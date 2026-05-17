# yt-transcript-md

[![Go Version](https://img.shields.io/github/go-mod/go-version/nawodyaishan/yt-transcript-md)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A fast, lightweight CLI tool written in Go to extract available YouTube transcripts and captions and export them into cleanly formatted Markdown files. Perfect for archiving video content, creating notes, or feeding data into LLMs.

## Features

- **Blazing Fast:** Written in Go for optimal performance and minimal footprint.
- **Multiple Inputs:** Provide comma-separated links, raw video IDs, or a text file containing URLs.
- **Smart URL Parsing:** Handles standard `youtube.com/watch`, `youtu.be`, `/shorts`, `/live`, and `/embed` links effortlessly.
- **Language Fallbacks:** Specify a priority list of language codes (e.g., `en,es,fr`) to fetch the best available transcript.
- **Timestamp Support:** Optionally include precise start times for every snippet of text.
- **Resilient:** Built-in retry mechanisms and partial success handling for batch processing large lists of videos.

## Installation

### Via Homebrew (macOS/Linux)

The easiest way to install on macOS and Linux is using Homebrew:

```bash
brew install nawodyaishan/tap/yt-transcript-md
```

### From Binary

Pre-compiled binaries for macOS, Linux, and Windows are available on the [Releases page](https://github.com/nawodyaishan/yt-transcript-md/releases).

### From Source

Ensure you have [Go 1.21+](https://go.dev/dl/) installed, then run:

```bash
git clone https://github.com/nawodyaishan/yt-transcript-md.git
cd yt-transcript-md
make build
```

The compiled binary will be placed at `bin/yt-transcript-md`.

## Usage

Extract a single transcript:

```bash
yt-transcript-md export --links "https://youtu.be/dQw4w9WgXcQ" --out transcripts.md
```

### Advanced Examples

**Batch extraction from multiple links:**
```bash
yt-transcript-md export \
  --links "dQw4w9WgXcQ,https://www.youtube.com/watch?v=jNQXAC9IVRw" \
  --out output/transcripts.md
```

**Extract from a text file:**
*(Assuming `links.txt` contains one URL or ID per line)*
```bash
yt-transcript-md export --input-file links.txt --out transcripts.md
```

**Include timestamps and prefer Spanish, falling back to English:**
```bash
yt-transcript-md export \
  --links "https://youtu.be/dQw4w9WgXcQ" \
  --languages "es,en" \
  --timestamps \
  --out transcripts.md
```

## Command Line Flags

| Flag | Short | Default | Description |
| :--- | :---: | :--- | :--- |
| `--links` | `-l` | | Comma-separated list of YouTube links or video IDs. |
| `--input-file` | `-f` | | Path to a text file containing YouTube links/IDs. |
| `--out` | `-o` | `transcripts.md` | The path where the Markdown file will be saved. |
| `--languages` | | `en` | Comma-separated list of ISO language codes to try. |
| `--timestamps` | | `false` | Include `[MM:SS]` timestamps before each snippet. |
| `--preserve-formatting` | | `false` | Preserve original HTML tags from YouTube. |
| `--retries` | | `1` | Number of times to retry a failed fetch. |
| `--retry-delay-seconds` | | `1.5` | Base delay between retries. |
| `--strict` | | `false` | Fail the entire process if any single video fails to fetch. |

## Limitations

- **No Audio Processing:** This tool fetches existing closed captions/transcripts directly from YouTube. It does not download audio or perform AI Speech-to-Text transcription. If a video has no captions available, the tool will report a failure for that video.
- **Availability:** Relies on YouTube's internal caption API endpoints, which may be subject to rate limiting if abused heavily.

## Development

Run all quality checks (formatting, linting, tests, build):

```bash
make verify
```

## License

[MIT License](LICENSE)
