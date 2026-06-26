# yt-transcript-md

[![Go Version](https://img.shields.io/github/go-mod/go-version/nawodyaishan/yt-transcript-md)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Turn YouTube links into clean Markdown transcripts — from your clipboard, your terminal args, or a file. No API key. No audio download.

```bash
# paste a link and go
yt-transcript-md https://youtu.be/dQw4w9WgXcQ https://youtu.be/jNQXAC9IVRw

# or just copy a link and run
yt-transcript-md
```

```text
[START] Processing 2 unique video(s)
[DETAILS] dQw4w9WgXcQ: Title: Never Gonna Give You Up | Channel: Rick Astley
[OK] Transcript fetched for dQw4w9WgXcQ
[DETAILS] jNQXAC9IVRw: Title: Me at the zoo | Channel: jawed
[OK] Transcript fetched for jNQXAC9IVRw
[SUMMARY] 2 succeeded, 0 failed, 0 warning(s)
[SAVED] Markdown written to /Users/you/project/transcripts.md
[COPIED] Markdown copied to clipboard
```

---

## Install

**Homebrew** (recommended)
```bash
brew install nawodyaishan/tap/yt-transcript-md
```

**Binary releases** — macOS, Linux, Windows on the [Releases page](https://github.com/nawodyaishan/yt-transcript-md/releases).

**From source** (requires Go 1.25.11+)
```bash
git clone https://github.com/nawodyaishan/yt-transcript-md.git
cd yt-transcript-md
make build          # output: bin/yt-transcript-md
```

---

## Three ways to use it

### 1. Pass links as arguments

No clipboard involved. Duplicates are deduplicated automatically.

```bash
yt-transcript-md https://youtu.be/abc https://youtu.be/def

# write to a specific file
yt-transcript-md https://youtu.be/abc --out notes.md

# works with the export subcommand
yt-transcript-md export https://youtu.be/abc --out notes.md
```

### 2. Copy a link, then run

Copy any YouTube URL or video ID to your clipboard, then:

```bash
yt-transcript-md
```

Saves `transcripts.md` and copies the Markdown back to your clipboard.

### 3. Batch from a file

```bash
yt-transcript-md export --input-file links.txt --out transcripts.md
```

---

## Clipboard history picker

In interactive terminals, `yt-transcript-md` scans your clipboard history and opens a full-screen picker when multiple YouTube videos are found. Copy links one by one from your browser — the tool collects them for you.

```text
Select YouTube videos from clipboard history

Search: █

> [ ] dQw4w9WgXcQ  current  https://youtu.be/dQw4w9WgXcQ
  [x] jNQXAC9IVRw  copyq    https://youtu.be/jNQXAC9IVRw
  [ ] BaW_jenozKc  copyq    https://youtu.be/BaW_jenozKc

space select  / search  a all visible  n first N  enter process  q cancel
```

**Supported clipboard managers:**

| Manager | Platform | Setup |
| :--- | :--- | :--- |
| **CopyQ** | macOS + Linux | Install and keep running |
| **cliphist** | Linux Wayland | `wl-paste --watch cliphist store` |
| **GPaste** | Linux GNOME | Keep the GPaste daemon running |

Control history scanning:

```bash
yt-transcript-md --history-source copyq --history-limit 25
yt-transcript-md --history-source cliphist
yt-transcript-md --no-history          # current clipboard only
```

History is processed locally. Raw entries are never logged.

---

## Non-interactive selection

For scripts and CI, skip the picker with an explicit selection flag:

```bash
yt-transcript-md --clipboard-selection all
yt-transcript-md --clipboard-selection one:2      # second video found
yt-transcript-md --clipboard-selection recent:3   # first three videos found
```

---

## Output options

```bash
# choose output path
yt-transcript-md https://youtu.be/abc --out notes.md

# add [MM:SS] timestamps to each line
yt-transcript-md https://youtu.be/abc --timestamps

# prefer Spanish, fall back to English
yt-transcript-md https://youtu.be/abc --languages "es,en"

# exit non-zero if any video fails (useful in CI)
yt-transcript-md https://youtu.be/abc --strict
```

Each output file includes title, channel, thumbnail URL, language, and snippet count when available.

---

## All flags

| Flag | Short | Default | Description |
| :--- | :---: | :---: | :--- |
| `--out` | `-o` | `transcripts.md` | Output file path |
| `--links` | `-l` | | Comma-separated links or video IDs |
| `--input-file` | `-f` | | Text file of links |
| `--languages` | | `en` | Comma-separated language priority |
| `--timestamps` | | `false` | Add `[MM:SS]` to each snippet |
| `--preserve-formatting` | | `false` | Keep YouTube HTML formatting |
| `--retries` | | `1` | Retries per failed fetch |
| `--retry-delay-seconds` | | `1.5` | Base retry delay |
| `--strict` | | `false` | Exit non-zero on any failure |
| `--clipboard-selection` | | | Non-interactive picker: `all`, `one:<n>`, `recent:<n>` |
| `--history-source` | | `auto` | History provider: `auto` `current` `copyq` `cliphist` `gpaste` |
| `--history-limit` | | `50` | Max history entries per provider |
| `--no-history` | | `false` | Disable history scanning |

Positional URL arguments cannot be combined with `--links`, `--input-file`, or clipboard flags.

---

## Clipboard system requirements

| Platform | Tool used |
| :--- | :--- |
| macOS | `pbpaste` / `pbcopy` |
| Linux | `wl-clipboard`, `xclip`, `xsel`, or Termux |
| Windows | PowerShell + `clip.exe` |

---

## Limitations

- Only fetches existing YouTube captions. Videos without captions fail gracefully.
- Public metadata (title, channel) is best-effort and may be unavailable.
- Does not download audio or run speech-to-text.

---

## Development

```bash
make verify       # vet + lint + test + vuln scan
make test         # tests only
make build        # compile to bin/
make run ARGS="--help"
```

Requires Go `1.25.11`+. Vulnerability scans use the version pinned in `go.mod`:

```bash
GOTOOLCHAIN=go1.25.11 govulncheck ./...
```

---

## License

[MIT](LICENSE)
