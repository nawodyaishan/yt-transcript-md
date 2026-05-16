# yt-transcript-md

Export available YouTube transcripts from comma-separated YouTube links or video IDs into a Markdown file.

## Install for development

```bash
uv sync --all-groups
```

Or with Make:

```bash
make sync
```

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
make check
```

Individual targets are also available:

```bash
make format
make lint
make typecheck
make test
```

Build the package:

```bash
make build
```

## Limitations

This tool fetches available YouTube captions/transcripts. It does not download audio or perform speech-to-text.
