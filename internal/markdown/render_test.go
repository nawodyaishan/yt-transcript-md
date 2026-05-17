package markdown

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

var update = flag.Bool("update", false, "update golden files")

func TestRender(t *testing.T) {
	now := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		docs      []models.TranscriptDocument
		failures  []models.FailedVideo
		opts      Options
		goldenFile string
	}{
		{
			name: "plain transcript",
			docs: []models.TranscriptDocument{
				{
					Video: models.VideoInput{
						Original: "https://youtu.be/dQw4w9WgXcQ",
						VideoID:  "dQw4w9WgXcQ",
					},
					Language:     "English",
					LanguageCode: "en",
					IsGenerated:  false,
					Snippets: []models.TranscriptSnippet{
						{Text: "Hello", Start: 0, Duration: 1},
						{Text: "world", Start: 1, Duration: 1},
					},
				},
			},
			opts:       Options{IncludeTimestamps: false, Now: now},
			goldenFile: "plain.golden",
		},
		{
			name: "timestamped transcript",
			docs: []models.TranscriptDocument{
				{
					Video: models.VideoInput{
						Original: "dQw4w9WgXcQ",
						VideoID:  "dQw4w9WgXcQ",
					},
					Language:     "English",
					LanguageCode: "en",
					IsGenerated:  true,
					Snippets: []models.TranscriptSnippet{
						{Text: "Hello", Start: 65, Duration: 1},
					},
				},
			},
			opts:       Options{IncludeTimestamps: true, Now: now},
			goldenFile: "timestamps.golden",
		},
		{
			name: "failures only",
			failures: []models.FailedVideo{
				{Original: "invalid", Reason: "could not extract ID"},
			},
			opts:       Options{Now: now},
			goldenFile: "failures.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.docs, tt.failures, tt.opts)
			goldenPath := filepath.Join("testdata", tt.goldenFile)

			if *update {
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			if got != string(want) {
				t.Errorf("Render() output does not match golden file %s. Run with -update to update.", tt.goldenFile)
			}
		})
	}
}
