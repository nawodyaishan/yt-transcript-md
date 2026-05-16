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
            console.print(f"[green]OK[/green] {video.video_id}")

        except TranscriptFetchError as exc:
            reason = str(exc)
            failures.append(FailedVideo(original=video.original, reason=reason))
            console.print(f"[red]FAILED[/red] {video.video_id}: {reason}")

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
