package models

// VideoInput represents a YouTube video input.
type VideoInput struct {
	Original string
	VideoID  string
}

// TranscriptSnippet represents a single snippet of a transcript.
type TranscriptSnippet struct {
	Text     string
	Start    float64
	Duration float64
}

// TranscriptDocument represents a full transcript for a video.
type TranscriptDocument struct {
	Video        VideoInput
	Metadata     VideoMetadata
	Language     string
	LanguageCode string
	IsGenerated  bool
	Snippets     []TranscriptSnippet
}

// VideoMetadata represents optional public metadata for a video.
type VideoMetadata struct {
	Title           string
	AuthorName      string
	AuthorURL       string
	ProviderName    string
	ProviderURL     string
	ThumbnailURL    string
	ThumbnailWidth  int
	ThumbnailHeight int
	CacheAgeSeconds int
}

// FailedVideo represents a video that failed to fetch.
type FailedVideo struct {
	Original string
	Reason   string
}
