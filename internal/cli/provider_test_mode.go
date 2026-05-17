//go:build test

package cli

import (
	"context"
	"errors"

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
