from yt_transcript_md.markdown import render_markdown
from yt_transcript_md.models import TranscriptDocument, TranscriptSnippet, VideoInput


def test_render_markdown_plain_transcript() -> None:
    document = TranscriptDocument(
        video=VideoInput(
            original="https://youtu.be/dQw4w9WgXcQ",
            video_id="dQw4w9WgXcQ",
        ),
        language="English",
        language_code="en",
        is_generated=False,
        snippets=(
            TranscriptSnippet(text="Hello", start=0.0, duration=1.0),
            TranscriptSnippet(text="world", start=1.0, duration=1.0),
        ),
    )

    output = render_markdown(
        documents=[document],
        failures=[],
        include_timestamps=False,
    )

    assert "# YouTube Transcripts" in output
    assert "Video `dQw4w9WgXcQ`" in output
    assert "Hello world" in output


def test_render_markdown_with_timestamps() -> None:
    document = TranscriptDocument(
        video=VideoInput(
            original="dQw4w9WgXcQ",
            video_id="dQw4w9WgXcQ",
        ),
        language="English",
        language_code="en",
        is_generated=True,
        snippets=(TranscriptSnippet(text="Hello", start=65.0, duration=1.0),),
    )

    output = render_markdown(
        documents=[document],
        failures=[],
        include_timestamps=True,
    )

    assert "`[01:05]` Hello" in output
