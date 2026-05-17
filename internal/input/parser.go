package input

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

var videoIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

var youtubeHosts = map[string]bool{
	"youtube.com":       true,
	"www.youtube.com":   true,
	"m.youtube.com":     true,
	"music.youtube.com": true,
}

// SplitInputLinks splits comma/newline-separated input into cleaned link strings.
func SplitInputLinks(raw string) []string {
	// Replace commas and newlines with a common separator then split
	raw = strings.ReplaceAll(raw, ",", "\n")
	lines := strings.Split(raw, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.Trim(trimmed, "<>")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ExtractVideoID extracts a YouTube video ID from a URL or raw video ID.
func ExtractVideoID(value string) (string, error) {
	cleaned := strings.TrimSpace(value)
	cleaned = strings.Trim(cleaned, "<>")

	if videoIDPattern.MatchString(cleaned) {
		return cleaned, nil
	}

	parsed, err := url.Parse(cleaned)
	if err != nil {
		return "", fmt.Errorf("could not parse URL: %w", err)
	}

	host := strings.ToLower(parsed.Host)

	if host == "youtu.be" {
		path := strings.TrimPrefix(parsed.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			return validateVideoID(parts[0], value)
		}
	}

	if youtubeHosts[host] {
		if parsed.Path == "/watch" {
			candidate := parsed.Query().Get("v")
			return validateVideoID(candidate, value)
		}

		pathPrefixes := []string{"/shorts/", "/embed/", "/live/"}
		for _, prefix := range pathPrefixes {
			if strings.HasPrefix(parsed.Path, prefix) {
				path := strings.TrimPrefix(parsed.Path, prefix)
				parts := strings.Split(path, "/")
				if len(parts) > 0 {
					return validateVideoID(parts[0], value)
				}
			}
		}
	}

	return "", fmt.Errorf("could not extract a YouTube video ID from: %s", value)
}

// ParseVideoInputs parses user input into unique VideoInput objects while preserving order.
func ParseVideoInputs(raw string) ([]models.VideoInput, error) {
	seen := make(map[string]bool)
	var videos []models.VideoInput

	for _, item := range SplitInputLinks(raw) {
		videoID, err := ExtractVideoID(item)
		if err != nil {
			return nil, err
		}

		if seen[videoID] {
			continue
		}

		seen[videoID] = true
		videos = append(videos, models.VideoInput{
			Original: item,
			VideoID:  videoID,
		})
	}

	return videos, nil
}

func validateVideoID(candidate string, original string) (string, error) {
	if candidate != "" && videoIDPattern.MatchString(candidate) {
		return candidate, nil
	}
	return "", fmt.Errorf("invalid YouTube video ID in: %s", original)
}
