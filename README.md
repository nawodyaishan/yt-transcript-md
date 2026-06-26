# yt-transcript-md

[![Go Version](https://img.shields.io/github/go-mod/go-version/nawodyaishan/yt-transcript-md)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Copy one or more YouTube links. Run one command. Get clean Markdown transcripts saved to disk and copied back to your clipboard.

`yt-transcript-md` is a clipboard-first CLI for turning YouTube captions into Markdown notes for research, archiving, and LLM workflows. It fetches existing YouTube transcripts; it does not download audio or run speech-to-text.

You can also pass links directly as arguments without touching the clipboard:

```bash
yt-transcript-md https://youtu.be/dQw4w9WgXcQ https://youtu.be/jNQXAC9IVRw
```

## Quick Start

Copy a YouTube link or video ID, then run:

```bash
yt-transcript-md
```

The default command:

- reads YouTube links from your clipboard
- scans supported clipboard-history managers in interactive terminals
- asks what to process when multiple unique videos are detected
- fetches available public video details and transcript text
- saves Markdown to `transcripts.md` and prints the full saved file path
- copies the same Markdown back to your clipboard

If the current clipboard or clipboard history contains multiple unique videos, interactive runs open a searchable TUI picker. Non-interactive runs can choose explicitly:

```bash
yt-transcript-md --clipboard-selection all
yt-transcript-md --clipboard-selection one:2
yt-transcript-md --clipboard-selection recent:3
```

The `recent` selector uses the order found in your clipboard, not the video's publish date.

Example status output:

```text
[START] Processing 1 unique video(s)
[INFO] Fetching video details for rJrd2QMVbGM
[DETAILS] rJrd2QMVbGM: Title: Example Video | Channel: Example Channel
[INFO] Fetching transcript for rJrd2QMVbGM
[OK] Transcript fetched for rJrd2QMVbGM
[SUMMARY] 1 succeeded, 0 failed, 0 warning(s)
[SAVED] Markdown written to /Users/you/project/transcripts.md
[COPIED] Markdown copied to clipboard
```

## Why Use It

- **Fast capture:** copy a link, run one command, paste or open the Markdown.
- **Markdown-ready output:** transcripts are formatted for notes, archives, and LLM context.
- **Useful video details:** title, channel, thumbnail, provider, language, and snippet counts are included when available.
- **Clear progress feedback:** colored terminal status shows metadata, title/channel, transcript fetch status, saved path, clipboard copy, and final summary.
- **Clipboard and file output:** default mode does both; advanced modes support chosen files and non-interactive batches.
- **Rate-limit conscious:** one sequential metadata request and one transcript fetch sequence per selected unique video, with bounded retries.
- **No audio processing:** uses available captions/transcripts directly.

## Markdown Output

Generated Markdown includes:

- generation timestamp and success/failure counts
- one section per unique video
- a `Video Details` block with title, channel, provider, thumbnail, source URL, language, generated-caption flag, and snippet count when available
- transcript text, optionally with timestamps
- failed video details for partial batch failures

## Clipboard Multi-Link Selection

Clipboard mode accepts a copied YouTube URL, raw video ID, or surrounding text that contains YouTube links. It extracts valid YouTube videos, deduplicates them, and keeps their first-seen order.

When one unique video is found, `yt-transcript-md` processes it immediately. When multiple unique videos are found in an interactive terminal, it opens a TUI before making any metadata or transcript requests.

```text
Select YouTube videos from clipboard history

Search:

> [ ] dQw4w9WgXcQ  current  https://youtu.be/dQw4w9WgXcQ
  [x] jNQXAC9IVRw  copyq    https://youtu.be/jNQXAC9IVRw

space select  / search  a all visible  n first N  enter process  q cancel
```

For scripts, CI, or redirected terminals, pass an explicit selection:

```bash
# Process every unique video found in the clipboard
yt-transcript-md --clipboard-selection all

# Process only the second unique video found in the clipboard
yt-transcript-md --clipboard-selection one:2

# Process the first three unique videos found in clipboard order
yt-transcript-md --clipboard-selection recent:3
```

`--clipboard-selection` only applies to the default clipboard workflow. Use `--links`, `--input-file`, or `export` for explicit non-clipboard batch input.

## Clipboard History Sources

In interactive terminals, `yt-transcript-md` can scan supported clipboard-history managers for recently copied YouTube links. It always reads the current clipboard first, then scans provider history when enabled.

Supported first-slice providers:

- **CopyQ** on macOS and Linux: install and keep CopyQ running.
- **cliphist** on Linux Wayland: configure history capture, for example `wl-paste --watch cliphist store`.
- **GPaste** on Linux GNOME: keep the GPaste daemon running.

Control history scanning:

```bash
# Auto-detect supported providers
yt-transcript-md --history-source auto

# Use only current clipboard, no provider history
yt-transcript-md --no-history

# Scan CopyQ history only
yt-transcript-md --history-source copyq --history-limit 25

# Scan cliphist or GPaste explicitly
yt-transcript-md --history-source cliphist
yt-transcript-md --history-source gpaste
```

History scanning is local-only. The tool extracts YouTube links locally, shows only sanitized previews, and does not log raw clipboard-history entries.

## Advanced Workflows

### Positional URL arguments

Pass one or more YouTube links or video IDs directly as arguments. The clipboard is not read or written:

```bash
# Single video
yt-transcript-md https://youtu.be/dQw4w9WgXcQ

# Multiple videos — space-separated, processed in argument order
yt-transcript-md https://youtu.be/dQw4w9WgXcQ https://youtu.be/jNQXAC9IVRw

# Write to a specific file
yt-transcript-md https://youtu.be/dQw4w9WgXcQ --out notes.md

# Works with the export subcommand too
yt-transcript-md export https://youtu.be/dQw4w9WgXcQ --out notes.md
```

Duplicate links are deduplicated by video ID. Positional arguments cannot be combined with `--links`, `--input-file`, or clipboard workflow flags (`--clipboard-selection`, `--history-source`, `--history-limit`, `--no-history`).

### Explicit links flag

Write a specific link to a chosen file:

```bash
yt-transcript-md --links "https://youtu.be/dQw4w9WgXcQ" --out notes.md
```

Batch export multiple videos:

```bash
yt-transcript-md export \
  --links "dQw4w9WgXcQ,https://www.youtube.com/watch?v=jNQXAC9IVRw" \
  --out output/transcripts.md
```

Export from a text file:

```bash
yt-transcript-md export --input-file links.txt --out transcripts.md
```

Include timestamps and prefer Spanish, falling back to English:

```bash
yt-transcript-md export \
  --links "https://youtu.be/dQw4w9WgXcQ" \
  --languages "es,en" \
  --timestamps \
  --out transcripts.md
```

## Installation

### Homebrew

```bash
brew install nawodyaishan/tap/yt-transcript-md
```

### Binary Releases

Pre-built binaries for macOS, Linux, and Windows are available on the [Releases page](https://github.com/nawodyaishan/yt-transcript-md/releases).

### From Source

Requires Go `1.25.11` or newer.

```bash
git clone https://github.com/nawodyaishan/yt-transcript-md.git
cd yt-transcript-md
make build
```

The compiled binary will be placed at `bin/yt-transcript-md`.

Build and run locally in one step:

```bash
make run
make run ARGS="--help"
```

## Command Line Flags

With no flags, `yt-transcript-md` runs the clipboard workflow. Pass YouTube links as positional arguments to bypass the clipboard entirely. Use root-level input flags or the `export` command for explicit file and batch workflows.

| Flag | Short | Default | Description |
| :--- | :---: | :--- | :--- |
| `--links` | `-l` | | Comma-separated YouTube links or video IDs. Cannot be combined with positional arguments. |
| `--input-file` | `-f` | | Text file containing YouTube links or IDs. |
| `--out` | `-o` | `transcripts.md` | Markdown output path. |
| `--clipboard-selection` | | | Resolve multi-link clipboard input without prompting: `all`, `one:<index>`, or `recent:<count>`. |
| `--history-source` | | `auto` | Clipboard history source: `auto`, `current`, `copyq`, `cliphist`, or `gpaste`. |
| `--history-limit` | | `50` | Maximum clipboard history entries to scan per provider. |
| `--no-history` | | `false` | Disable clipboard history scanning and use only the current clipboard. |
| `--languages` | | `en` | Comma-separated language priority list. |
| `--timestamps` | | `false` | Include `[MM:SS]` timestamps before each snippet. |
| `--preserve-formatting` | | `false` | Preserve YouTube transcript HTML formatting. |
| `--retries` | | `1` | Number of retries per failed transcript fetch. |
| `--retry-delay-seconds` | | `1.5` | Base retry delay in seconds. |
| `--strict` | | `false` | Exit non-zero if any video fails. |

## Clipboard Support

Clipboard mode uses native clipboard tools:

- macOS: `pbpaste` and `pbcopy`
- Windows: PowerShell clipboard support and `clip.exe`
- Linux: one of `wl-clipboard`, `xclip`, `xsel`, or Termux clipboard commands

## Limitations

- The tool only fetches existing YouTube captions/transcripts.
- Videos without available captions will be reported as failures.
- Public metadata is best-effort and may be omitted when unavailable.
- YouTube endpoints may rate limit abusive usage; this tool keeps requests sequential and deduplicates repeated inputs per run.

## Development

This project is built and scanned with the Go version declared in `go.mod`. Keep the local toolchain at Go `1.25.11` or newer so standard-library vulnerability scans use the patched runtime.

Run all quality checks:

```bash
make verify
```

Run the Go test suite directly:

```bash
go test ./...
```

Run the vulnerability scan used by CI and release workflows:

```bash
govulncheck ./...
```

If your local Go installation is newer but `govulncheck` has toolchain-loading issues, run the scan with the project-pinned patched toolchain:

```bash
GOTOOLCHAIN=go1.25.11 govulncheck ./...
```

## License

[MIT License](LICENSE)
