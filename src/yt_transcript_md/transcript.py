from collections.abc import Sequence
from time import sleep

from youtube_transcript_api import YouTubeTranscriptApi

from yt_transcript_md.errors import TranscriptFetchError
from yt_transcript_md.models import TranscriptDocument, TranscriptSnippet, VideoInput


class YouTubeTranscriptProvider:
    """Fetches available YouTube transcripts/captions."""

    def __init__(self) -> None:
        self._client = YouTubeTranscriptApi()

    def fetch(
        self,
        video: VideoInput,
        languages: Sequence[str],
        preserve_formatting: bool,
        retries: int,
        retry_delay_seconds: float,
    ) -> TranscriptDocument:
        last_error: Exception | None = None

        for attempt in range(retries + 1):
            try:
                fetched = self._client.fetch(
                    video.video_id,
                    languages=list(languages),
                    preserve_formatting=preserve_formatting,
                )

                snippets = tuple(
                    TranscriptSnippet(
                        text=snippet.text,
                        start=float(snippet.start),
                        duration=float(snippet.duration),
                    )
                    for snippet in fetched
                )

                return TranscriptDocument(
                    video=video,
                    language=str(fetched.language),
                    language_code=str(fetched.language_code),
                    is_generated=bool(fetched.is_generated),
                    snippets=snippets,
                )

            except Exception as exc:
                last_error = exc

                if attempt < retries:
                    sleep(retry_delay_seconds * (attempt + 1))

        raise TranscriptFetchError(
            f"Failed to fetch transcript for {video.video_id}: {last_error}"
        ) from last_error
