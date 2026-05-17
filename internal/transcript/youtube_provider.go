package transcript

import (
	"context"
	"fmt"
	"time"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

// YouTubeProvider fetches transcripts using horiagug/youtube-transcript-api-go.
type YouTubeProvider struct {
	client *yt_transcript.YtTranscriptClient
}

// NewYouTubeProvider creates a new YouTubeProvider.
func NewYouTubeProvider() *YouTubeProvider {
	return &YouTubeProvider{
		client: yt_transcript.NewClient(),
	}
}

// Fetch retrieves the transcript for a video with retries.
func (p *YouTubeProvider) Fetch(ctx context.Context, video models.VideoInput, opts FetchOptions) (models.TranscriptDocument, error) {
	var lastErr error

	for i := 0; i <= opts.Retries; i++ {
		// Check for context cancellation before each attempt.
		if err := ctx.Err(); err != nil {
			return models.TranscriptDocument{}, err
		}

		if i > 0 {
			delay := time.Duration(opts.RetryDelaySeconds*float64(i)) * time.Second
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return models.TranscriptDocument{}, ctx.Err()
			}
		}

		// The library doesn't take a context, so we just call it.
		// We could use a goroutine and a channel to enforce a timeout if needed,
		// but for now we'll rely on the default HTTP client behavior or configure the client if possible.
		transcriptList, err := p.client.GetTranscripts(video.VideoID, opts.Languages)
		if err != nil {
			lastErr = err
			continue
		}

		if len(transcriptList) == 0 {
			lastErr = fmt.Errorf("no transcript found for video %s in languages %v", video.VideoID, opts.Languages)
			continue
		}

		// Use the first available transcript from the requested languages.
		t := transcriptList[0]

		snippets := make([]models.TranscriptSnippet, len(t.Lines))
		for j, line := range t.Lines {
			snippets[j] = models.TranscriptSnippet{
				Text:     line.Text,
				Start:    line.Start,
				Duration: line.Duration,
			}
		}

		return models.TranscriptDocument{
			Video:        video,
			Language:     t.Language,
			LanguageCode: t.LanguageCode,
			IsGenerated:  t.IsGenerated,
			Snippets:     snippets,
		}, nil
	}

	return models.TranscriptDocument{}, fmt.Errorf("failed to fetch transcript after %d retries: %w", opts.Retries, lastErr)
}
