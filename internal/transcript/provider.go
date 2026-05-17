package transcript

import (
	"context"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

// FetchOptions represents the options for fetching a transcript.
type FetchOptions struct {
	Languages          []string
	PreserveFormatting bool
	Retries            int
	RetryDelaySeconds  float64
}

// Provider is an interface for fetching YouTube transcripts.
type Provider interface {
	Fetch(ctx context.Context, video models.VideoInput, opts FetchOptions) (models.TranscriptDocument, error)
}
