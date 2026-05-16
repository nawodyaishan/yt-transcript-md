from datetime import UTC, datetime

from yt_transcript_md.models import FailedVideo, TranscriptDocument, TranscriptSnippet


def render_markdown(
    documents: list[TranscriptDocument],
    failures: list[FailedVideo],
    include_timestamps: bool,
) -> str:
    lines: list[str] = [
        "# YouTube Transcripts",
        "",
        f"Generated at: `{datetime.now(UTC).isoformat(timespec='seconds')}`",
        "",
        f"Successful videos: **{len(documents)}**",
        f"Failed videos: **{len(failures)}**",
        "",
        "---",
        "",
    ]

    for index, document in enumerate(documents, start=1):
        lines.extend(_render_document(index, document, include_timestamps))

    if failures:
        lines.extend(_render_failures(failures))

    return "\n".join(lines).rstrip() + "\n"


def _render_document(
    index: int,
    document: TranscriptDocument,
    include_timestamps: bool,
) -> list[str]:
    video = document.video

    lines = [
        f"## {index}. Video `{video.video_id}`",
        "",
        f"- Source: {video.original}",
        f"- Language: {document.language} (`{document.language_code}`)",
        f"- Auto-generated: `{str(document.is_generated).lower()}`",
        f"- Snippets: `{len(document.snippets)}`",
        "",
        "### Transcript",
        "",
    ]

    if include_timestamps:
        lines.extend(_render_timestamped_transcript(document.snippets))
    else:
        lines.append(_render_plain_transcript(document.snippets))

    lines.extend(["", "---", ""])
    return lines


def _render_plain_transcript(snippets: tuple[TranscriptSnippet, ...]) -> str:
    text = " ".join(_normalize_text(snippet.text) for snippet in snippets if snippet.text.strip())
    return _wrap_into_paragraphs(text)


def _render_timestamped_transcript(snippets: tuple[TranscriptSnippet, ...]) -> list[str]:
    lines: list[str] = []

    for snippet in snippets:
        text = _normalize_text(snippet.text)
        if not text:
            continue

        timestamp = _format_timestamp(snippet.start)
        lines.append(f"- `[{timestamp}]` {text}")

    return lines


def _render_failures(failures: list[FailedVideo]) -> list[str]:
    lines = ["## Failed Videos", ""]

    for failure in failures:
        lines.append(f"- `{failure.original}` - {failure.reason}")

    lines.append("")
    return lines


def _normalize_text(text: str) -> str:
    return " ".join(text.replace("\n", " ").split()).strip()


def _wrap_into_paragraphs(text: str, words_per_paragraph: int = 120) -> str:
    words = text.split()
    paragraphs = [
        " ".join(words[index : index + words_per_paragraph])
        for index in range(0, len(words), words_per_paragraph)
    ]
    return "\n\n".join(paragraphs)


def _format_timestamp(seconds: float) -> str:
    total_seconds = int(seconds)
    hours, remainder = divmod(total_seconds, 3600)
    minutes, seconds = divmod(remainder, 60)

    if hours:
        return f"{hours:02d}:{minutes:02d}:{seconds:02d}"

    return f"{minutes:02d}:{seconds:02d}"
