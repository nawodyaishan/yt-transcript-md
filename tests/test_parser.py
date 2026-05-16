import pytest

from yt_transcript_md.errors import InvalidYouTubeLinkError
from yt_transcript_md.parser import extract_video_id, parse_video_inputs, split_input_links


@pytest.mark.parametrize(
    ("value", "expected"),
    [
        ("dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ("https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ("https://m.youtube.com/watch?v=dQw4w9WgXcQ&t=12", "dQw4w9WgXcQ"),
        ("https://www.youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ("https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
        ("https://www.youtube.com/live/dQw4w9WgXcQ", "dQw4w9WgXcQ"),
    ],
)
def test_extract_video_id(value: str, expected: str) -> None:
    assert extract_video_id(value) == expected


def test_extract_video_id_rejects_invalid_url() -> None:
    with pytest.raises(InvalidYouTubeLinkError):
        extract_video_id("https://example.com/not-youtube")


def test_split_input_links_supports_commas_and_newlines() -> None:
    raw = "a,b\nc"
    assert split_input_links(raw) == ["a", "b", "c"]


def test_parse_video_inputs_deduplicates_ids() -> None:
    raw = "dQw4w9WgXcQ,https://youtu.be/dQw4w9WgXcQ"
    videos = parse_video_inputs(raw)

    assert len(videos) == 1
    assert videos[0].video_id == "dQw4w9WgXcQ"
