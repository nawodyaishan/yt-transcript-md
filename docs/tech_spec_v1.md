# Technical specification

## Product requirements

### Input

Support:

```bash
yt-transcript-md --links "https://youtu.be/abc123abc12,https://www.youtube.com/watch?v=xyz123xyz12"
```

Also support raw IDs:

```bash
yt-transcript-md --links "abc123abc12,xyz123xyz12"
```

And file input:

```bash
yt-transcript-md --input-file links.txt --out transcripts.md
```

### Output

Generate one Markdown file:

```md
# YouTube Transcripts

## 1. Video `abc123abc12`

- URL: https://youtu.be/abc123abc12
- Language: English (`en`)
- Auto-generated: false

### Transcript

...
```

### Non-goals for MVP

This tool does **not** download video/audio.  
This tool does **not** run speech-to-text.  
This tool only fetches available YouTube captions/transcripts.

---

# Architecture

```text
yt-transcript-md/
  pyproject.toml
  README.md
  .python-version
  .gitignore
  src/
    yt_transcript_md/
      __init__.py
      __main__.py
      cli.py
      errors.py
      markdown.py
      models.py
      parser.py
      transcript.py
  tests/
    test_parser.py
    test_markdown.py
  .github/
    workflows/
      ci.yml
```

## Layer responsibilities

| Layer               | File            | Responsibility                                |
| ------------------- | --------------- | --------------------------------------------- |
| CLI                 | `cli.py`        | Parse options, orchestrate fetching/exporting |
| Parsing             | `parser.py`     | Convert links into valid YouTube video IDs    |
| Transcript provider | `transcript.py` | Fetch captions via `youtube-transcript-api`   |
| Domain models       | `models.py`     | Typed dataclasses                             |
| Markdown rendering  | `markdown.py`   | Convert successful/failed results to Markdown |
| Errors              | `errors.py`     | App-specific exceptions                       |
| Tests               | `tests/`        | Unit tests for parsing/rendering              |

Typer is a clean CLI choice because it is based on Python type hints and provides automatic help/completion behavior. ([typer.tiangolo.com](https://typer.tiangolo.com/)) Ruff is the recommended single-tool formatter/linter here; its formatter is designed as a fast Black-compatible formatter available through `ruff format`. ([Astral Docs](https://docs.astral.sh/ruff/formatter/))

---

# Setup with `uv`

```bash
mkdir yt-transcript-md
cd yt-transcript-md

uv init --package --name yt-transcript-md --python 3.12

uv add typer rich youtube-transcript-api
uv add --dev pytest pytest-cov ruff mypy
```

Run locally:

```bash
uv run yt-transcript-md --links "https://youtu.be/abc123abc12" --out transcripts.md
```

Run quality checks:

```bash
uv run ruff format .
uv run ruff check .
uv run mypy src
uv run pytest
```

Build package:

```bash
uv build
```

`uv build` creates distribution artifacts under `dist/` by default. ([Astral Docs](https://docs.astral.sh/uv/guides/projects/))

---

# `pyproject.toml`

```toml
[project]
name = "yt-transcript-md"
version = "0.1.0"
description = "Export available YouTube transcripts from comma-separated links into Markdown."
readme = "README.md"
requires-python = ">=3.11,<3.15"
license = { text = "MIT" }
authors = [
  { name = "Nawodya Ishan", email = "nawodyain@gmail.com" }
]
keywords = ["youtube", "transcripts", "markdown", "cli", "captions"]
dependencies = [
  "rich>=13.7.0",
  "typer>=0.12.0",
  "youtube-transcript-api>=1.2.4",
]

[project.scripts]
yt-transcript-md = "yt_transcript_md.cli:main"

[dependency-groups]
dev = [
  "mypy>=1.10.0",
  "pytest>=8.0.0",
  "pytest-cov>=5.0.0",
  "ruff>=0.6.0",
]

[build-system]
requires = ["hatchling>=1.25.0"]
build-backend = "hatchling.build"

[tool.ruff]
line-length = 100
target-version = "py311"
src = ["src", "tests"]

[tool.ruff.lint]
select = [
  "E",
  "F",
  "I",
  "B",
  "UP",
  "SIM",
  "RUF",
]
ignore = []

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
docstring-code-format = true

[tool.mypy]
python_version = "3.11"
strict = true
warn_unused_ignores = true
warn_redundant_casts = true
warn_return_any = true
disallow_untyped_defs = true
no_implicit_optional = true
mypy_path = "src"

[[tool.mypy.overrides]]
module = ["youtube_transcript_api.*"]
ignore_missing_imports = true

[tool.pytest.ini_options]
testpaths = ["tests"]
addopts = "-q --cov=yt_transcript_md --cov-report=term-missing"
```

---

# Source code

## `src/yt_transcript_md/__init__.py`

```python
__version__ = "0.1.0"
```

## `src/yt_transcript_md/__main__.py`

```python
from yt_transcript_md.cli import main

if __name__ == "__main__":
    main()
```

## `src/yt_transcript_md/errors.py`

```python
class YTTranscriptMDError(Exception):
    """Base application error."""


class InvalidYouTubeLinkError(YTTranscriptMDError):
    """Raised when a YouTube URL or video ID cannot be parsed."""


class TranscriptFetchError(YTTranscriptMDError):
    """Raised when a transcript cannot be fetched."""
```

## `src/yt_transcript_md/models.py`

```python
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
```

## `src/yt_transcript_md/parser.py`

```python
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
        candidate = parsed.path.strip("/").split("/")[0]
        return _validate_video_id(candidate, value)

    if host in YOUTUBE_HOSTS:
        if parsed.path == "/watch":
            candidate = parse_qs(parsed.query).get("v", [None])[0]
            return _validate_video_id(candidate, value)

        path_prefixes = ("/shorts/", "/embed/", "/live/")
        for prefix in path_prefixes:
            if parsed.path.startswith(prefix):
                candidate = parsed.path.removeprefix(prefix).split("/")[0]
                return _validate_video_id(candidate, value)

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
```

## `src/yt_transcript_md/transcript.py`

```python
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

            except Exception as exc:  # noqa: BLE001
                last_error = exc

                if attempt < retries:
                    sleep(retry_delay_seconds * (attempt + 1))

        raise TranscriptFetchError(
            f"Failed to fetch transcript for {video.video_id}: {last_error}"
        ) from last_error
```

## `src/yt_transcript_md/markdown.py`

```python
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
        lines.append(f"- `{failure.original}` — {failure.reason}")

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
```

## `src/yt_transcript_md/cli.py`

```python
from pathlib import Path
from typing import Annotated

import typer
from rich.console import Console

from yt_transcript_md.errors import InvalidYouTubeLinkError, TranscriptFetchError
from yt_transcript_md.markdown import render_markdown
from yt_transcript_md.models import FailedVideo, TranscriptDocument
from yt_transcript_md.parser import parse_video_inputs
from yt_transcript_md.transcript import YouTubeTranscriptProvider

app = typer.Typer(
    name="yt-transcript-md",
    help="Export available YouTube transcripts from links/video IDs to Markdown.",
    no_args_is_help=True,
)

console = Console()


@app.command()
def export(
    links: Annotated[
        str | None,
        typer.Option(
            "--links",
            "-l",
            help="Comma-separated YouTube links or video IDs.",
        ),
    ] = None,
    input_file: Annotated[
        Path | None,
        typer.Option(
            "--input-file",
            "-f",
            help="Text file containing comma-separated or newline-separated links.",
        ),
    ] = None,
    out: Annotated[
        Path,
        typer.Option(
            "--out",
            "-o",
            help="Output Markdown file path.",
        ),
    ] = Path("transcripts.md"),
    languages: Annotated[
        str,
        typer.Option(
            "--languages",
            help="Comma-separated language priority list. Example: en,si,hi",
        ),
    ] = "en",
    timestamps: Annotated[
        bool,
        typer.Option(
            "--timestamps",
            help="Include per-snippet timestamps in the Markdown output.",
        ),
    ] = False,
    preserve_formatting: Annotated[
        bool,
        typer.Option(
            "--preserve-formatting",
            help="Preserve YouTube transcript HTML formatting where available.",
        ),
    ] = False,
    retries: Annotated[
        int,
        typer.Option(
            "--retries",
            min=0,
            help="Number of retries per failed transcript fetch.",
        ),
    ] = 1,
    retry_delay_seconds: Annotated[
        float,
        typer.Option(
            "--retry-delay-seconds",
            min=0.0,
            help="Base retry delay in seconds.",
        ),
    ] = 1.5,
    strict: Annotated[
        bool,
        typer.Option(
            "--strict",
            help="Exit with a non-zero code if any video fails.",
        ),
    ] = False,
) -> None:
    """Fetch transcripts and write a Markdown file."""
    raw_input = _read_raw_input(links=links, input_file=input_file)
    language_priority = _parse_languages(languages)

    try:
        videos = parse_video_inputs(raw_input)
    except InvalidYouTubeLinkError as exc:
        console.print(f"[red]Input error:[/red] {exc}")
        raise typer.Exit(code=2) from exc

    if not videos:
        console.print("[red]No valid YouTube links or video IDs provided.[/red]")
        raise typer.Exit(code=2)

    provider = YouTubeTranscriptProvider()
    documents: list[TranscriptDocument] = []
    failures: list[FailedVideo] = []

    console.print(f"[bold]Fetching transcripts for {len(videos)} video(s)...[/bold]")

    for video in videos:
        try:
            document = provider.fetch(
                video=video,
                languages=language_priority,
                preserve_formatting=preserve_formatting,
                retries=retries,
                retry_delay_seconds=retry_delay_seconds,
            )
            documents.append(document)
            console.print(f"[green]✓[/green] {video.video_id}")

        except TranscriptFetchError as exc:
            reason = str(exc)
            failures.append(FailedVideo(original=video.original, reason=reason))
            console.print(f"[red]✗[/red] {video.video_id}: {reason}")

            if strict:
                break

    markdown = render_markdown(
        documents=documents,
        failures=failures,
        include_timestamps=timestamps,
    )

    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(markdown, encoding="utf-8")

    console.print(f"\n[bold green]Wrote Markdown:[/bold green] {out}")

    if failures and strict:
        raise typer.Exit(code=1)


def _read_raw_input(links: str | None, input_file: Path | None) -> str:
    values: list[str] = []

    if links:
        values.append(links)

    if input_file:
        if not input_file.exists():
            console.print(f"[red]Input file not found:[/red] {input_file}")
            raise typer.Exit(code=2)

        values.append(input_file.read_text(encoding="utf-8"))

    if not values:
        console.print("[red]Provide --links or --input-file.[/red]")
        raise typer.Exit(code=2)

    return "\n".join(values)


def _parse_languages(raw: str) -> list[str]:
    languages = [item.strip() for item in raw.split(",") if item.strip()]

    if not languages:
        return ["en"]

    return languages


def main() -> None:
    app()
```

---

# Tests

## `tests/test_parser.py`

```python
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
```

## `tests/test_markdown.py`

```python
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
```

---

# GitHub Actions CI

The official uv GitHub Actions guide recommends `astral-sh/setup-uv`; it also recommends pinning a specific uv version in CI. ([Astral Docs](https://docs.astral.sh/uv/guides/integration/github/?utm_source=chatgpt.com))

## `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: ["main"]
  pull_request:

jobs:
  quality:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v6

      - name: Install uv
        uses: astral-sh/setup-uv@08807647e7069bb48b6ef5acd8ec9567f424441b # v8.1.0
        with:
          version: "0.11.14"

      - name: Install Python
        run: uv python install

      - name: Sync dependencies
        run: uv sync --all-groups --locked

      - name: Check formatting
        run: uv run ruff format --check .

      - name: Lint
        run: uv run ruff check .

      - name: Type check
        run: uv run mypy src

      - name: Test
        run: uv run pytest
```

---

# README starter

````md
# yt-transcript-md

Export available YouTube transcripts from comma-separated YouTube links or video IDs into a Markdown file.

## Install for development

```bash
uv sync --all-groups
```
````

## Usage

```bash
uv run yt-transcript-md --links "https://youtu.be/dQw4w9WgXcQ" --out transcripts.md
```

With multiple links:

```bash
uv run yt-transcript-md \
  --links "https://youtu.be/id111111111,https://www.youtube.com/watch?v=id222222222" \
  --out output/transcripts.md
```

With language priority:

```bash
uv run yt-transcript-md \
  --links "https://youtu.be/id111111111" \
  --languages "en,si,hi" \
  --out transcripts.md
```

With timestamps:

```bash
uv run yt-transcript-md \
  --links "https://youtu.be/id111111111" \
  --timestamps \
  --out transcripts.md
```

From a file:

```bash
uv run yt-transcript-md --input-file links.txt --out transcripts.md
```

## Quality checks

```bash
uv run ruff format .
uv run ruff check .
uv run mypy src
uv run pytest
```

## Limitations

This tool fetches available YouTube captions/transcripts. It does not download audio or perform speech-to-text.

````

---

# Example commands

```bash
uv sync --all-groups
````

```bash
uv run yt-transcript-md \
  --links "https://youtu.be/dQw4w9WgXcQ,https://www.youtube.com/watch?v=jNQXAC9IVRw" \
  --languages "en" \
  --timestamps \
  --out output/transcripts.md
```

---

# Production improvements after MVP

Add these later:

| Feature                       | Why                                                                                                                                                                           |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--json-out`                  | Save machine-readable transcript data                                                                                                                                         |
| `--format plain,timestamped`  | More output control                                                                                                                                                           |
| `--continue-on-error` default | Batch jobs should not fail on one bad video                                                                                                                                   |
| `--proxy-*` options           | `youtube-transcript-api` supports proxy configuration, but proxies still do not guarantee avoiding YouTube blocks. ([PyPI](https://pypi.org/project/youtube-transcript-api/)) |
| Metadata provider             | Use YouTube Data API or `yt-dlp` metadata for video title/channel                                                                                                             |
| STT fallback                  | For videos without captions, optionally transcribe audio with Whisper-style pipeline                                                                                          |
| Dockerfile                    | Useful for server/batch execution                                                                                                                                             |
| Pre-commit                    | Run Ruff before commits                                                                                                                                                       |

This gives you a clean CLI package with `uv`, typed code, tests, linting, formatting, CI, and a realistic transcript-fetching architecture.
