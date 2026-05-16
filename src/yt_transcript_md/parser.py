import re
from urllib.parse import parse_qs, urlparse

from yt_transcript_md.errors import InvalidYouTubeLinkError
from yt_transcript_md.models import VideoInput

VIDEO_ID_PATTERN = re.compile(r"^[a-zA-Z0-9_-]{11}$")

YOUTUBE_HOSTS = {
    "youtube.com",
    "www.youtube.com",
    "m.youtube.com",
    "music.youtube.com",
}


def split_input_links(raw: str) -> list[str]:
    """Split comma/newline-separated input into cleaned link strings."""
    return [item.strip().strip("<>") for item in re.split(r"[,\n]+", raw) if item.strip()]


def extract_video_id(value: str) -> str:
    """Extract a YouTube video ID from a URL or raw video ID."""
    cleaned = value.strip().strip("<>")

    if VIDEO_ID_PATTERN.fullmatch(cleaned):
        return cleaned

    parsed = urlparse(cleaned)
    host = parsed.netloc.lower()

    if host == "youtu.be":
        path_candidate = parsed.path.strip("/").split("/")[0]
        return _validate_video_id(path_candidate, value)

    if host in YOUTUBE_HOSTS:
        if parsed.path == "/watch":
            values = parse_qs(parsed.query).get("v", [])
            query_candidate = values[0] if values else None
            return _validate_video_id(query_candidate, value)

        path_prefixes = ("/shorts/", "/embed/", "/live/")
        for prefix in path_prefixes:
            if parsed.path.startswith(prefix):
                path_candidate = parsed.path.removeprefix(prefix).split("/")[0]
                return _validate_video_id(path_candidate, value)

    raise InvalidYouTubeLinkError(f"Could not extract a YouTube video ID from: {value}")


def parse_video_inputs(raw: str) -> list[VideoInput]:
    """Parse user input into unique VideoInput objects while preserving order."""
    seen: set[str] = set()
    videos: list[VideoInput] = []

    for item in split_input_links(raw):
        video_id = extract_video_id(item)
        if video_id in seen:
            continue

        seen.add(video_id)
        videos.append(VideoInput(original=item, video_id=video_id))

    return videos


def _validate_video_id(candidate: str | None, original: str) -> str:
    if candidate and VIDEO_ID_PATTERN.fullmatch(candidate):
        return candidate

    raise InvalidYouTubeLinkError(f"Invalid YouTube video ID in: {original}")
