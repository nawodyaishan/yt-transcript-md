package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/metadata"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

type fakeProvider struct {
	responses map[string]models.TranscriptDocument
	errs      map[string]error
	calls     map[string]int
}

func (f *fakeProvider) Fetch(ctx context.Context, video models.VideoInput, opts transcript.FetchOptions) (models.TranscriptDocument, error) {
	if f.calls == nil {
		f.calls = make(map[string]int)
	}
	f.calls[video.VideoID]++
	if err, ok := f.errs[video.VideoID]; ok {
		return models.TranscriptDocument{}, err
	}
	if doc, ok := f.responses[video.VideoID]; ok {
		return doc, nil
	}
	return models.TranscriptDocument{}, errors.New("not found")
}

type fakeMetadataProvider struct {
	responses map[string]models.VideoMetadata
	errs      map[string]error
	calls     map[string]int
}

func (f *fakeMetadataProvider) Fetch(ctx context.Context, video models.VideoInput, opts metadata.FetchOptions) (models.VideoMetadata, error) {
	if f.calls == nil {
		f.calls = make(map[string]int)
	}
	f.calls[video.VideoID]++
	if err, ok := f.errs[video.VideoID]; ok {
		return models.VideoMetadata{}, err
	}
	if response, ok := f.responses[video.VideoID]; ok {
		return response, nil
	}
	return models.VideoMetadata{}, nil
}

type fakeClipboard struct {
	read     string
	readErr  error
	writeErr error
	wrote    string
}

func (f *fakeClipboard) ReadAll() (string, error) {
	if f.readErr != nil {
		return "", f.readErr
	}
	return f.read, nil
}

func (f *fakeClipboard) WriteAll(text string) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.wrote = text
	return nil
}

func TestExport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "yt-transcript-md-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	provider := &fakeProvider{
		responses: map[string]models.TranscriptDocument{
			"dQw4w9WgXcQ": {
				Video: models.VideoInput{
					Original: "dQw4w9WgXcQ",
					VideoID:  "dQw4w9WgXcQ",
				},
				Language:     "English",
				LanguageCode: "en",
				Snippets:     []models.TranscriptSnippet{{Text: "Hello"}},
			},
		},
		errs: map[string]error{
			"fail1234567": errors.New("fetch failed"),
		},
	}

	t.Run("success", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "out.md")
		opts := ExportOptions{
			Links: "dQw4w9WgXcQ",
			Out:   outPath,
		}

		err := Export(context.Background(), opts, provider, nil, io.Discard)
		if err != nil {
			t.Errorf("Export() error = %v", err)
		}

		if _, err := os.Stat(outPath); os.IsNotExist(err) {
			t.Errorf("output file was not created")
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "partial.md")
		opts := ExportOptions{
			Links:  "dQw4w9WgXcQ,fail1234567",
			Out:    outPath,
			Strict: false,
		}

		err := Export(context.Background(), opts, provider, nil, io.Discard)
		if err != nil {
			t.Errorf("Export() error = %v", err)
		}

		data, _ := os.ReadFile(outPath)
		if !contains(string(data), "Failed Videos") {
			t.Errorf("output should contain failed videos")
		}
	})

	t.Run("strict failure", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "strict.md")
		opts := ExportOptions{
			Links:  "dQw4w9WgXcQ,fail1234567",
			Out:    outPath,
			Strict: true,
		}

		err := Export(context.Background(), opts, provider, nil, io.Discard)
		if err == nil {
			t.Errorf("Export() should return error in strict mode")
		}
	})

	t.Run("missing input", func(t *testing.T) {
		opts := ExportOptions{}
		err := Export(context.Background(), opts, provider, nil, io.Discard)
		if err == nil {
			t.Errorf("Export() should return error with no input")
		}
	})
}

func TestExportClipboard(t *testing.T) {
	provider := &fakeProvider{
		responses: map[string]models.TranscriptDocument{
			"dQw4w9WgXcQ": {
				Video: models.VideoInput{
					Original: "https://youtu.be/dQw4w9WgXcQ",
					VideoID:  "dQw4w9WgXcQ",
				},
				Language:     "English",
				LanguageCode: "en",
				Snippets:     []models.TranscriptSnippet{{Text: "Clipboard transcript"}},
			},
		},
	}

	t.Run("saves and copies rendered transcript", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "transcripts.md")
		clipboard := &fakeClipboard{read: "https://youtu.be/dQw4w9WgXcQ"}
		metadataProvider := &fakeMetadataProvider{
			responses: map[string]models.VideoMetadata{
				"dQw4w9WgXcQ": {
					Title:      "Clipboard Video",
					AuthorName: "Clipboard Channel",
				},
			},
		}
		opts := ExportOptions{Languages: "en", Out: outPath}

		err := ExportClipboard(context.Background(), opts, clipboard, provider, metadataProvider, io.Discard)
		if err != nil {
			t.Fatalf("ExportClipboard() error = %v", err)
		}

		saved, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("output file was not created: %v", err)
		}

		if !contains(clipboard.wrote, "Video `dQw4w9WgXcQ`") {
			t.Errorf("clipboard output missing video ID: %s", clipboard.wrote)
		}
		if !contains(clipboard.wrote, "Clipboard transcript") {
			t.Errorf("clipboard output missing transcript: %s", clipboard.wrote)
		}
		if string(saved) != clipboard.wrote {
			t.Errorf("saved output and clipboard output differ")
		}
		if !contains(clipboard.wrote, "Clipboard Video") {
			t.Errorf("clipboard output missing metadata: %s", clipboard.wrote)
		}
	})

	t.Run("empty clipboard", func(t *testing.T) {
		clipboard := &fakeClipboard{read: " \n\t "}
		err := ExportClipboard(context.Background(), ExportOptions{}, clipboard, provider, nil, io.Discard)
		if err == nil {
			t.Fatal("ExportClipboard() should return error with empty clipboard")
		}
	})

	t.Run("metadata failure warns and continues", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "metadata-failure.md")
		clipboard := &fakeClipboard{read: "https://youtu.be/dQw4w9WgXcQ"}
		metadataProvider := &fakeMetadataProvider{
			errs: map[string]error{"dQw4w9WgXcQ": errors.New("metadata unavailable")},
		}
		var log strings.Builder

		err := ExportClipboard(context.Background(), ExportOptions{Out: outPath}, clipboard, provider, metadataProvider, &log)
		if err != nil {
			t.Fatalf("ExportClipboard() error = %v", err)
		}
		if !contains(log.String(), "[WARN] Video details unavailable for dQw4w9WgXcQ") {
			t.Errorf("log missing metadata warning: %s", log.String())
		}
		if !contains(clipboard.wrote, "Clipboard transcript") {
			t.Errorf("clipboard output missing transcript after metadata failure")
		}
	})

	t.Run("saved message uses absolute path", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "absolute.md")
		var log strings.Builder

		err := Export(context.Background(), ExportOptions{
			Links: "dQw4w9WgXcQ",
			Out:   outPath,
		}, provider, nil, &log)
		if err != nil {
			t.Fatalf("Export() error = %v", err)
		}
		absolute, err := filepath.Abs(outPath)
		if err != nil {
			t.Fatal(err)
		}
		if !contains(log.String(), "[SAVED] Markdown written to "+absolute) {
			t.Errorf("saved message missing absolute path: %s", log.String())
		}
	})

	t.Run("deduplicates transcript and metadata fetches", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "dedupe.md")
		provider := &fakeProvider{
			responses: map[string]models.TranscriptDocument{
				"dQw4w9WgXcQ": {
					Video: models.VideoInput{Original: "dQw4w9WgXcQ", VideoID: "dQw4w9WgXcQ"},
					Snippets: []models.TranscriptSnippet{
						{Text: "Hello"},
					},
				},
			},
		}
		metadataProvider := &fakeMetadataProvider{
			responses: map[string]models.VideoMetadata{
				"dQw4w9WgXcQ": {Title: "Deduped"},
			},
		}

		err := Export(context.Background(), ExportOptions{
			Links: "dQw4w9WgXcQ,https://youtu.be/dQw4w9WgXcQ",
			Out:   outPath,
		}, provider, metadataProvider, io.Discard)
		if err != nil {
			t.Fatalf("Export() error = %v", err)
		}
		if provider.calls["dQw4w9WgXcQ"] != 1 {
			t.Errorf("transcript fetches = %d, want 1", provider.calls["dQw4w9WgXcQ"])
		}
		if metadataProvider.calls["dQw4w9WgXcQ"] != 1 {
			t.Errorf("metadata fetches = %d, want 1", metadataProvider.calls["dQw4w9WgXcQ"])
		}
	})

	t.Run("file write failure prevents clipboard write", func(t *testing.T) {
		badDir := filepath.Join(t.TempDir(), "not-dir")
		if err := os.WriteFile(badDir, []byte("not a directory"), 0644); err != nil {
			t.Fatal(err)
		}
		outPath := filepath.Join(badDir, "out.md")
		clipboard := &fakeClipboard{read: "https://youtu.be/dQw4w9WgXcQ"}

		err := ExportClipboard(context.Background(), ExportOptions{Out: outPath}, clipboard, provider, nil, io.Discard)
		if err == nil {
			t.Fatal("ExportClipboard() should return file write error")
		}
		if clipboard.wrote != "" {
			t.Errorf("clipboard was written after file failure")
		}
	})

	t.Run("clipboard write failure leaves saved file", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "transcripts.md")
		clipboard := &fakeClipboard{
			read:     "https://youtu.be/dQw4w9WgXcQ",
			writeErr: errors.New("clipboard failed"),
		}

		err := ExportClipboard(context.Background(), ExportOptions{Out: outPath}, clipboard, provider, nil, io.Discard)
		if err == nil {
			t.Fatal("ExportClipboard() should return clipboard write error")
		}
		if _, statErr := os.Stat(outPath); statErr != nil {
			t.Fatalf("output file should remain after clipboard failure: %v", statErr)
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
