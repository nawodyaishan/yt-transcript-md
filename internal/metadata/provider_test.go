package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

func TestOEmbedProviderFetch(t *testing.T) {
	video := models.VideoInput{Original: "https://youtu.be/dQw4w9WgXcQ", VideoID: "dQw4w9WgXcQ"}

	t.Run("decodes metadata", func(t *testing.T) {
		var gotURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotURL = r.URL.Query().Get("url")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"title":            "Example Video",
				"author_name":      "Example Channel",
				"author_url":       "https://www.youtube.com/@example",
				"provider_name":    "YouTube",
				"provider_url":     "https://www.youtube.com/",
				"thumbnail_url":    "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg",
				"thumbnail_width":  480,
				"thumbnail_height": 360,
				"cache_age":        3600,
				"html":             "<iframe></iframe>",
			})
		}))
		defer server.Close()

		provider := NewOEmbedProviderWithClient(server.Client(), server.URL)
		got, err := provider.Fetch(context.Background(), video, FetchOptions{})
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}

		if gotURL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
			t.Errorf("request URL = %q", gotURL)
		}
		if got.Title != "Example Video" {
			t.Errorf("Title = %q", got.Title)
		}
		if got.AuthorName != "Example Channel" {
			t.Errorf("AuthorName = %q", got.AuthorName)
		}
		if got.ThumbnailWidth != 480 || got.ThumbnailHeight != 360 {
			t.Errorf("thumbnail dimensions = %dx%d", got.ThumbnailWidth, got.ThumbnailHeight)
		}
	})

	t.Run("does not retry client errors", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer server.Close()

		provider := NewOEmbedProviderWithClient(server.Client(), server.URL)
		_, err := provider.Fetch(context.Background(), video, FetchOptions{Retries: 2})
		if err == nil {
			t.Fatal("Fetch() should return an error")
		}
		if requests != 1 {
			t.Errorf("requests = %d, want 1", requests)
		}
	})

	t.Run("retries server errors", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if requests == 1 {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			_, _ = w.Write([]byte(`{"title":"Recovered"}`))
		}))
		defer server.Close()

		provider := NewOEmbedProviderWithClient(server.Client(), server.URL)
		got, err := provider.Fetch(context.Background(), video, FetchOptions{Retries: 1})
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if requests != 2 {
			t.Errorf("requests = %d, want 2", requests)
		}
		if got.Title != "Recovered" {
			t.Errorf("Title = %q", got.Title)
		}
	})

	t.Run("respects canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		provider := NewOEmbedProviderWithClient(http.DefaultClient, "https://example.test/oembed")
		_, err := provider.Fetch(ctx, video, FetchOptions{Retries: 1})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Fetch() error = %v, want context.Canceled", err)
		}
	})
}
