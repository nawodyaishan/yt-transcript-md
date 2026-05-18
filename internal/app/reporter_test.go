package app

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

func TestReporterColorModes(t *testing.T) {
	t.Run("forced color", func(t *testing.T) {
		var out bytes.Buffer
		reporter := NewReporterWithColor(&out, true)

		reporter.TranscriptSuccess("dQw4w9WgXcQ")

		got := out.String()
		if !strings.Contains(got, "\x1b[32m[OK]\x1b[0m") {
			t.Fatalf("colored output missing ANSI status: %q", got)
		}
		if !strings.Contains(got, "Transcript fetched for dQw4w9WgXcQ") {
			t.Fatalf("colored output missing message: %q", got)
		}
	})

	t.Run("forced plain", func(t *testing.T) {
		var out bytes.Buffer
		reporter := NewReporterWithColor(&out, false)

		reporter.MetadataWarning("dQw4w9WgXcQ", io.ErrUnexpectedEOF)

		got := out.String()
		if strings.Contains(got, "\x1b[") {
			t.Fatalf("plain output includes ANSI sequence: %q", got)
		}
		if !strings.Contains(got, "[WARN] Video details unavailable for dQw4w9WgXcQ") {
			t.Fatalf("plain output missing warning text: %q", got)
		}
	})

	t.Run("non file writer disables color", func(t *testing.T) {
		var out bytes.Buffer
		reporter := NewReporter(&out)

		reporter.Copied()

		got := out.String()
		if strings.Contains(got, "\x1b[") {
			t.Fatalf("non-file writer output includes ANSI sequence: %q", got)
		}
	})

	t.Run("metadata success includes title and channel", func(t *testing.T) {
		var out bytes.Buffer
		reporter := NewReporterWithColor(&out, false)

		reporter.MetadataSuccess("dQw4w9WgXcQ", models.VideoMetadata{
			Title:      "Example Video",
			AuthorName: "Example Channel",
		})

		got := out.String()
		if !strings.Contains(got, "[DETAILS] dQw4w9WgXcQ: Title: Example Video | Channel: Example Channel") {
			t.Fatalf("metadata success missing title/channel: %q", got)
		}
	})

	t.Run("metadata success omits empty details", func(t *testing.T) {
		var out bytes.Buffer
		reporter := NewReporterWithColor(&out, false)

		reporter.MetadataSuccess("dQw4w9WgXcQ", models.VideoMetadata{})

		if out.String() != "" {
			t.Fatalf("empty metadata should not emit output: %q", out.String())
		}
	})

	t.Run("NO_COLOR disables color detection", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if detectColor(io.Discard) {
			t.Fatal("detectColor() should be false when NO_COLOR is set")
		}
	})
}
