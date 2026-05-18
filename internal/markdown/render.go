package markdown

import (
	"fmt"
	"strings"
	"time"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

// Options represents the rendering options.
type Options struct {
	IncludeTimestamps bool
	Now               time.Time
}

// Render renders the successful transcripts and failures to a Markdown string.
func Render(documents []models.TranscriptDocument, failures []models.FailedVideo, opts Options) string {
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}

	var sb strings.Builder
	sb.WriteString("# YouTube Transcripts\n\n")
	_, _ = fmt.Fprintf(&sb, "Generated at: `%s`\n\n", opts.Now.Format("2006-01-02T15:04:05Z07:00"))
	_, _ = fmt.Fprintf(&sb, "Successful videos: **%d**\n", len(documents))
	_, _ = fmt.Fprintf(&sb, "Failed videos: **%d**\n\n", len(failures))
	sb.WriteString("---\n\n")

	for i, doc := range documents {
		sb.WriteString(renderDocument(i+1, doc, opts.IncludeTimestamps))
	}

	if len(failures) > 0 {
		sb.WriteString(renderFailures(failures))
	}

	return strings.TrimSpace(sb.String()) + "\n"
}

func renderDocument(index int, doc models.TranscriptDocument, includeTimestamps bool) string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "## %d. Video `%s`\n\n", index, doc.Video.VideoID)
	sb.WriteString("### Video Details\n\n")
	if doc.Metadata.Title != "" {
		_, _ = fmt.Fprintf(&sb, "- Title: %s\n", doc.Metadata.Title)
	}
	if doc.Metadata.AuthorName != "" {
		if doc.Metadata.AuthorURL != "" {
			_, _ = fmt.Fprintf(&sb, "- Channel: [%s](%s)\n", doc.Metadata.AuthorName, doc.Metadata.AuthorURL)
		} else {
			_, _ = fmt.Fprintf(&sb, "- Channel: %s\n", doc.Metadata.AuthorName)
		}
	}
	if doc.Metadata.ProviderName != "" {
		if doc.Metadata.ProviderURL != "" {
			_, _ = fmt.Fprintf(&sb, "- Provider: [%s](%s)\n", doc.Metadata.ProviderName, doc.Metadata.ProviderURL)
		} else {
			_, _ = fmt.Fprintf(&sb, "- Provider: %s\n", doc.Metadata.ProviderName)
		}
	}
	if doc.Metadata.ThumbnailURL != "" {
		_, _ = fmt.Fprintf(&sb, "- Thumbnail: %s", doc.Metadata.ThumbnailURL)
		if doc.Metadata.ThumbnailWidth > 0 && doc.Metadata.ThumbnailHeight > 0 {
			_, _ = fmt.Fprintf(&sb, " (`%dx%d`)", doc.Metadata.ThumbnailWidth, doc.Metadata.ThumbnailHeight)
		}
		sb.WriteString("\n")
	}
	_, _ = fmt.Fprintf(&sb, "- Source: %s\n", doc.Video.Original)
	_, _ = fmt.Fprintf(&sb, "- Language: %s (`%s`)\n", doc.Language, doc.LanguageCode)
	_, _ = fmt.Fprintf(&sb, "- Auto-generated: `%t`\n", doc.IsGenerated)
	_, _ = fmt.Fprintf(&sb, "- Snippets: `%d`\n\n", len(doc.Snippets))
	sb.WriteString("### Transcript\n\n")

	if includeTimestamps {
		sb.WriteString(renderTimestampedTranscript(doc.Snippets))
	} else {
		sb.WriteString(renderPlainTranscript(doc.Snippets))
	}

	sb.WriteString("\n---\n\n")
	return sb.String()
}

func renderPlainTranscript(snippets []models.TranscriptSnippet) string {
	var texts []string
	for _, s := range snippets {
		normalized := normalizeText(s.Text)
		if normalized != "" {
			texts = append(texts, normalized)
		}
	}
	fullText := strings.Join(texts, " ")
	return wrapIntoParagraphs(fullText, 120)
}

func renderTimestampedTranscript(snippets []models.TranscriptSnippet) string {
	var lines []string
	for _, s := range snippets {
		normalized := normalizeText(s.Text)
		if normalized == "" {
			continue
		}
		timestamp := formatTimestamp(s.Start)
		lines = append(lines, fmt.Sprintf("- `[%s]` %s", timestamp, normalized))
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderFailures(failures []models.FailedVideo) string {
	var sb strings.Builder
	sb.WriteString("## Failed Videos\n\n")
	for _, f := range failures {
		_, _ = fmt.Fprintf(&sb, "- `%s` — %s\n", f.Original, f.Reason)
	}
	sb.WriteString("\n")
	return sb.String()
}

func normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	words := strings.Fields(text)
	return strings.Join(words, " ")
}

func wrapIntoParagraphs(text string, wordsPerParagraph int) string {
	words := strings.Fields(text)
	var paragraphs []string
	for i := 0; i < len(words); i += wordsPerParagraph {
		end := i + wordsPerParagraph
		if end > len(words) {
			end = len(words)
		}
		paragraphs = append(paragraphs, strings.Join(words[i:end], " "))
	}
	return strings.Join(paragraphs, "\n\n")
}

func formatTimestamp(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
