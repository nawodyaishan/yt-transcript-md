package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

const (
	colorReset  = "\x1b[0m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorRed    = "\x1b[31m"
	colorCyan   = "\x1b[36m"
)

// Reporter writes user-facing CLI progress messages.
type Reporter struct {
	writer io.Writer
	color  bool
}

// NewReporter creates a reporter with color enabled only for suitable terminals.
func NewReporter(writer io.Writer) *Reporter {
	return &Reporter{writer: writer, color: detectColor(writer)}
}

// NewReporterWithColor creates a reporter with explicit color behavior for tests.
func NewReporterWithColor(writer io.Writer, color bool) *Reporter {
	return &Reporter{writer: writer, color: color}
}

func (r *Reporter) Start(videoCount int) {
	r.info("START", "Processing %d unique video(s)", videoCount)
}

func (r *Reporter) MetadataStart(videoID string) {
	r.info("INFO", "Fetching video details for %s", videoID)
}

func (r *Reporter) MetadataWarning(videoID string, err error) {
	r.warn("WARN", "Video details unavailable for %s: %v", videoID, err)
}

func (r *Reporter) MetadataSuccess(videoID string, metadata models.VideoMetadata) {
	details := metadataSummary(metadata)
	if details == "" {
		return
	}
	r.success("DETAILS", "%s: %s", videoID, details)
}

func (r *Reporter) TranscriptStart(videoID string) {
	r.info("INFO", "Fetching transcript for %s", videoID)
}

func (r *Reporter) TranscriptSuccess(videoID string) {
	r.success("OK", "Transcript fetched for %s", videoID)
}

func (r *Reporter) TranscriptFailure(videoID string, err error) {
	r.failure("ERROR", "Transcript failed for %s: %v", videoID, err)
}

func (r *Reporter) Saved(path string) {
	r.success("SAVED", "Markdown written to %s", path)
}

func (r *Reporter) Copied() {
	r.success("COPIED", "Markdown copied to clipboard")
}

func (r *Reporter) Summary(successes int, failures int, warnings int) {
	if failures > 0 || warnings > 0 {
		r.warn("SUMMARY", "%d succeeded, %d failed, %d warning(s)", successes, failures, warnings)
		return
	}
	r.success("SUMMARY", "%d succeeded, %d failed, %d warning(s)", successes, failures, warnings)
}

func (r *Reporter) info(status string, format string, args ...any) {
	r.line(colorCyan, status, format, args...)
}

func (r *Reporter) success(status string, format string, args ...any) {
	r.line(colorGreen, status, format, args...)
}

func (r *Reporter) warn(status string, format string, args ...any) {
	r.line(colorYellow, status, format, args...)
}

func (r *Reporter) failure(status string, format string, args ...any) {
	r.line(colorRed, status, format, args...)
}

func (r *Reporter) line(color string, status string, format string, args ...any) {
	if r == nil || r.writer == nil {
		return
	}

	message := fmt.Sprintf(format, args...)
	if r.color {
		_, _ = fmt.Fprintf(r.writer, "%s[%s]%s %s\n", color, status, colorReset, message)
		return
	}
	_, _ = fmt.Fprintf(r.writer, "[%s] %s\n", status, message)
}

func detectColor(writer io.Writer) bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

func metadataSummary(metadata models.VideoMetadata) string {
	var parts []string
	if metadata.Title != "" {
		parts = append(parts, "Title: "+metadata.Title)
	}
	if metadata.AuthorName != "" {
		parts = append(parts, "Channel: "+metadata.AuthorName)
	}
	return strings.Join(parts, " | ")
}
