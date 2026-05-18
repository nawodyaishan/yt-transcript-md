package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nawodyaishan/yt-transcript-md/internal/input"
	"github.com/nawodyaishan/yt-transcript-md/internal/markdown"
	"github.com/nawodyaishan/yt-transcript-md/internal/metadata"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

// ExportOptions represents the options for the export command.
type ExportOptions struct {
	Links              string
	InputFile          string
	Out                string
	Languages          string
	Timestamps         bool
	PreserveFormatting bool
	Retries            int
	RetryDelaySeconds  float64
	Strict             bool
}

// Clipboard represents the clipboard operations needed by the default command.
type Clipboard interface {
	ReadAll() (string, error)
	WriteAll(string) error
}

type exportResult struct {
	Markdown string
}

// Export orchestrates the fetching and rendering of transcripts.
func Export(ctx context.Context, opts ExportOptions, provider transcript.Provider, metadataProvider metadata.Provider, log io.Writer) error {
	rawInput, err := readRawInput(opts)
	if err != nil {
		return err
	}

	reporter := NewReporter(log)
	result, err := renderExport(ctx, rawInput, opts, provider, metadataProvider, reporter)
	if result.Markdown != "" {
		outPath := outputPath(opts.Out)
		if writeErr := writeOutputFile(outPath, result.Markdown); writeErr != nil {
			return fmt.Errorf("failed to write output: %w", writeErr)
		}

		reporter.Saved(displayPath(outPath))
	}

	return err
}

// ExportClipboard reads input from the clipboard, saves Markdown, and copies it back.
func ExportClipboard(ctx context.Context, opts ExportOptions, clipboard Clipboard, provider transcript.Provider, metadataProvider metadata.Provider, log io.Writer) error {
	if clipboard == nil {
		return fmt.Errorf("clipboard is not configured")
	}

	rawInput, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %w", err)
	}

	if strings.TrimSpace(rawInput) == "" {
		return fmt.Errorf("clipboard is empty")
	}

	reporter := NewReporter(log)
	result, err := renderExport(ctx, rawInput, opts, provider, metadataProvider, reporter)
	if result.Markdown != "" {
		outPath := outputPath(opts.Out)
		if writeErr := writeOutputFile(outPath, result.Markdown); writeErr != nil {
			return fmt.Errorf("failed to write output: %w", writeErr)
		}
		reporter.Saved(displayPath(outPath))

		if writeErr := clipboard.WriteAll(result.Markdown); writeErr != nil {
			return fmt.Errorf("failed to write clipboard: %w", writeErr)
		}

		reporter.Copied()
	}

	return err
}

func renderExport(ctx context.Context, rawInput string, opts ExportOptions, provider transcript.Provider, metadataProvider metadata.Provider, reporter *Reporter) (exportResult, error) {
	videos, err := input.ParseVideoInputs(rawInput)
	if err != nil {
		return exportResult{}, fmt.Errorf("input error: %w", err)
	}

	if len(videos) == 0 {
		return exportResult{}, fmt.Errorf("no valid YouTube links or video IDs provided")
	}

	languagePriority := parseLanguages(opts.Languages)
	fetchOpts := transcript.FetchOptions{
		Languages:          languagePriority,
		PreserveFormatting: opts.PreserveFormatting,
		Retries:            opts.Retries,
		RetryDelaySeconds:  opts.RetryDelaySeconds,
	}
	metadataOpts := metadata.FetchOptions{
		Retries:           opts.Retries,
		RetryDelaySeconds: opts.RetryDelaySeconds,
	}

	var documents []models.TranscriptDocument
	var failures []models.FailedVideo
	metadataWarnings := 0

	reporter.Start(len(videos))

	for _, video := range videos {
		var videoMetadata models.VideoMetadata
		if metadataProvider != nil {
			reporter.MetadataStart(video.VideoID)
			fetchedMetadata, err := metadataProvider.Fetch(ctx, video, metadataOpts)
			if err != nil {
				metadataWarnings++
				reporter.MetadataWarning(video.VideoID, err)
			} else {
				videoMetadata = fetchedMetadata
				reporter.MetadataSuccess(video.VideoID, fetchedMetadata)
			}
		}

		reporter.TranscriptStart(video.VideoID)
		doc, err := provider.Fetch(ctx, video, fetchOpts)
		if err != nil {
			failures = append(failures, models.FailedVideo{
				Original: video.Original,
				Reason:   err.Error(),
			})
			reporter.TranscriptFailure(video.VideoID, err)
			if opts.Strict {
				break
			}
			continue
		}

		doc.Metadata = videoMetadata
		documents = append(documents, doc)
		reporter.TranscriptSuccess(video.VideoID)
	}

	reporter.Summary(len(documents), len(failures), metadataWarnings)

	md := markdown.Render(documents, failures, markdown.Options{
		IncludeTimestamps: opts.Timestamps,
	})

	result := exportResult{Markdown: md}

	if opts.Strict && len(failures) > 0 {
		return result, fmt.Errorf("strict mode: %d videos failed", len(failures))
	}

	return result, nil
}

func outputPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "transcripts.md"
	}
	return path
}

func readRawInput(opts ExportOptions) (string, error) {
	var parts []string

	if opts.Links != "" {
		parts = append(parts, opts.Links)
	}

	if opts.InputFile != "" {
		data, err := os.ReadFile(opts.InputFile)
		if err != nil {
			return "", fmt.Errorf("input file not found: %w", err)
		}
		parts = append(parts, string(data))
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("provide --links or --input-file")
	}

	return strings.Join(parts, "\n"), nil
}

func parseLanguages(raw string) []string {
	var result []string
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{"en"}
	}
	return result
}

func writeOutputFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func displayPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absolute
}
