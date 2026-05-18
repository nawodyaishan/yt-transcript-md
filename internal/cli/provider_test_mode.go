//go:build test

package cli

import (
	"context"
	"errors"
	"os"

	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/nawodyaishan/yt-transcript-md/internal/metadata"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

type testProvider struct{}

func (p *testProvider) Fetch(ctx context.Context, video models.VideoInput, opts transcript.FetchOptions) (models.TranscriptDocument, error) {
	if video.VideoID == "fail1234567" {
		return models.TranscriptDocument{}, errors.New("fetch failed")
	}
	return models.TranscriptDocument{
		Video:        video,
		Language:     "English",
		LanguageCode: "en",
		IsGenerated:  false,
		Snippets: []models.TranscriptSnippet{
			{Text: "Test transcript snippet", Start: 0, Duration: 1},
		},
	}, nil
}

func getProvider() transcript.Provider {
	return &testProvider{}
}

type testClipboard struct{}

func (testClipboard) ReadAll() (string, error) {
	text := os.Getenv("YT_TRANSCRIPT_MD_TEST_CLIPBOARD")
	if text == "" {
		return "", errors.New("YT_TRANSCRIPT_MD_TEST_CLIPBOARD is required")
	}
	return text, nil
}

func (testClipboard) WriteAll(text string) error {
	path := os.Getenv("YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT")
	if path == "" {
		return errors.New("YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT is required")
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func getClipboard() app.Clipboard {
	return testClipboard{}
}

type testMetadataProvider struct{}

func (testMetadataProvider) Fetch(ctx context.Context, video models.VideoInput, opts metadata.FetchOptions) (models.VideoMetadata, error) {
	if video.VideoID == "failmeta123" {
		return models.VideoMetadata{}, errors.New("metadata failed")
	}
	return models.VideoMetadata{
		Title:        "Test Video",
		AuthorName:   "Test Channel",
		AuthorURL:    "https://www.youtube.com/@test",
		ProviderName: "YouTube",
		ProviderURL:  "https://www.youtube.com/",
		ThumbnailURL: "https://i.ytimg.com/vi/" + video.VideoID + "/hqdefault.jpg",
	}, nil
}

func getMetadataProvider() metadata.Provider {
	return testMetadataProvider{}
}
