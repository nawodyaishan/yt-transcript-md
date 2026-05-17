package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"strings"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

type fakeProvider struct {
	responses map[string]models.TranscriptDocument
	errs      map[string]error
}

func (f *fakeProvider) Fetch(ctx context.Context, video models.VideoInput, opts transcript.FetchOptions) (models.TranscriptDocument, error) {
	if err, ok := f.errs[video.VideoID]; ok {
		return models.TranscriptDocument{}, err
	}
	if doc, ok := f.responses[video.VideoID]; ok {
		return doc, nil
	}
	return models.TranscriptDocument{}, errors.New("not found")
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

		err := Export(context.Background(), opts, provider, io.Discard)
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

		err := Export(context.Background(), opts, provider, io.Discard)
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

		err := Export(context.Background(), opts, provider, io.Discard)
		if err == nil {
			t.Errorf("Export() should return error in strict mode")
		}
	})

	t.Run("missing input", func(t *testing.T) {
		opts := ExportOptions{}
		err := Export(context.Background(), opts, provider, io.Discard)
		if err == nil {
			t.Errorf("Export() should return error with no input")
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
