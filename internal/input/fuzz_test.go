package input

import (
	"testing"
)

func FuzzExtractVideoID(f *testing.F) {
	seeds := []string{
		"dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=12",
		"https://www.youtube.com/shorts/dQw4w9WgXcQ",
		"<https://youtu.be/dQw4w9WgXcQ>",
		"",
		"too-long-id-123",
		"too-short",
		"https://example.com",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		_, _ = ExtractVideoID(value)
	})
}

func FuzzParseVideoInputs(f *testing.F) {
	seeds := []string{
		"dQw4w9WgXcQ,https://youtu.be/dQw4w9WgXcQ",
		"abc123abc12\nxyz123xyz12",
		"invalid-url,abc123abc12",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		_, _ = ParseVideoInputs(value)
	})
}
