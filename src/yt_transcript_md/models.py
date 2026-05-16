from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class VideoInput:
    original: str
    video_id: str


@dataclass(frozen=True, slots=True)
class TranscriptSnippet:
    text: str
    start: float
    duration: float


@dataclass(frozen=True, slots=True)
class TranscriptDocument:
    video: VideoInput
    language: str
    language_code: str
    is_generated: bool
    snippets: tuple[TranscriptSnippet, ...]


@dataclass(frozen=True, slots=True)
class FailedVideo:
    original: str
    reason: str
