class YTTranscriptMDError(Exception):
    """Base application error."""


class InvalidYouTubeLinkError(YTTranscriptMDError):
    """Raised when a YouTube URL or video ID cannot be parsed."""


class TranscriptFetchError(YTTranscriptMDError):
    """Raised when a transcript cannot be fetched."""
