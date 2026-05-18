package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

const defaultOEmbedEndpoint = "https://www.youtube.com/oembed"

// FetchOptions controls metadata fetch retries.
type FetchOptions struct {
	Retries           int
	RetryDelaySeconds float64
}

// Provider fetches optional public metadata for a video.
type Provider interface {
	Fetch(ctx context.Context, video models.VideoInput, opts FetchOptions) (models.VideoMetadata, error)
}

// OEmbedProvider fetches public YouTube metadata from the oEmbed endpoint.
type OEmbedProvider struct {
	client   *http.Client
	endpoint string
}

type oEmbedResponse struct {
	Title           string `json:"title"`
	AuthorName      string `json:"author_name"`
	AuthorURL       string `json:"author_url"`
	ProviderName    string `json:"provider_name"`
	ProviderURL     string `json:"provider_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
	CacheAgeSeconds int    `json:"cache_age"`
}

// NewOEmbedProvider creates a provider using YouTube's public oEmbed endpoint.
func NewOEmbedProvider() *OEmbedProvider {
	return NewOEmbedProviderWithClient(&http.Client{Timeout: 5 * time.Second}, defaultOEmbedEndpoint)
}

// NewOEmbedProviderWithClient creates a provider with injectable HTTP settings.
func NewOEmbedProviderWithClient(client *http.Client, endpoint string) *OEmbedProvider {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	if endpoint == "" {
		endpoint = defaultOEmbedEndpoint
	}
	return &OEmbedProvider{
		client:   client,
		endpoint: endpoint,
	}
}

// Fetch returns public metadata for a video.
func (p *OEmbedProvider) Fetch(ctx context.Context, video models.VideoInput, opts FetchOptions) (models.VideoMetadata, error) {
	var lastErr error

	for attempt := 0; attempt <= opts.Retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return models.VideoMetadata{}, err
		}

		if attempt > 0 {
			if err := waitForRetry(ctx, opts.RetryDelaySeconds, attempt); err != nil {
				return models.VideoMetadata{}, err
			}
		}

		metadata, retryable, err := p.fetchOnce(ctx, video)
		if err == nil {
			return metadata, nil
		}

		lastErr = err
		if !retryable {
			break
		}
	}

	return models.VideoMetadata{}, fmt.Errorf("metadata fetch failed: %w", lastErr)
}

func (p *OEmbedProvider) fetchOnce(ctx context.Context, video models.VideoInput) (models.VideoMetadata, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.requestURL(video), nil)
	if err != nil {
		return models.VideoMetadata{}, false, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return models.VideoMetadata{}, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		err := fmt.Errorf("oEmbed returned %s", resp.Status)
		return models.VideoMetadata{}, resp.StatusCode >= http.StatusInternalServerError, err
	}

	var payload oEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return models.VideoMetadata{}, false, err
	}

	return models.VideoMetadata{
		Title:           payload.Title,
		AuthorName:      payload.AuthorName,
		AuthorURL:       payload.AuthorURL,
		ProviderName:    payload.ProviderName,
		ProviderURL:     payload.ProviderURL,
		ThumbnailURL:    payload.ThumbnailURL,
		ThumbnailWidth:  payload.ThumbnailWidth,
		ThumbnailHeight: payload.ThumbnailHeight,
		CacheAgeSeconds: payload.CacheAgeSeconds,
	}, false, nil
}

func (p *OEmbedProvider) requestURL(video models.VideoInput) string {
	values := url.Values{}
	values.Set("format", "json")
	values.Set("url", "https://www.youtube.com/watch?v="+video.VideoID)
	return p.endpoint + "?" + values.Encode()
}

func waitForRetry(ctx context.Context, delaySeconds float64, attempt int) error {
	delay := time.Duration(delaySeconds*float64(attempt)) * time.Second
	if delay <= 0 {
		return nil
	}

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
