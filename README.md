# yt-transcript-md

Export available YouTube transcripts from YouTube links or video IDs into a Markdown file.

## Installation

### Via Homebrew

```bash
brew install nawodyaishan/tap/yt-transcript-md
```

### From Binary

Download the latest binary from the [Releases](https://github.com/nawodyaishan/yt-transcript-md/releases) page.

### From Source

Requires Go 1.21+.

```bash
git clone https://github.com/nawodyaishan/yt-transcript-md.git
cd yt-transcript-md
make build-go
```

The binary will be available at `bin/yt-transcript-md`.

## Usage

```bash
yt-transcript-md export --links "https://youtu.be/dQw4w9WgXcQ" --out transcripts.md
```

With multiple links:

```bash
yt-transcript-md export \
  --links "https://youtu.be/id111111111,https://www.youtube.com/watch?v=id222222222" \
  --out output/transcripts.md
```

With language priority:

```bash
yt-transcript-md export \
  --links "https://youtu.be/id111111111" \
  --languages "en,si,hi" \
  --out transcripts.md
```

With timestamps:

```bash
yt-transcript-md export \
  --links "https://youtu.be/id111111111" \
  --timestamps \
  --out transcripts.md
```

From a file:

```bash
yt-transcript-md export --input-file links.txt --out transcripts.md
```

## Quality checks

```bash
make verify
```

## Limitations

This tool fetches available YouTube captions/transcripts. It does not download audio or perform speech-to-text.
