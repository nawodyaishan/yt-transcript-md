# Test Plan: Clipboard-First Rebrand, Default Save, Rich Metadata, and Colored Feedback

## Unit Tests

### App Workflow

- default clipboard workflow reads clipboard input, writes `transcripts.md`, and writes the same Markdown back to clipboard
- file-output workflow does not require clipboard access
- metadata success is attached to transcript documents and rendered
- metadata failure logs a warning and does not fail successful transcript export
- duplicate input links produce one transcript fetch and one metadata fetch
- strict transcript failure still returns an error and preserves existing strict semantics
- file write failure prevents clipboard write in default mode
- clipboard write failure after file save returns an error while file content exists

### Metadata Provider

- builds a YouTube oEmbed URL from parsed video input
- decodes title, author, provider, thumbnail, and cache-age fields
- ignores oEmbed HTML
- treats 4xx responses as non-retryable errors
- retries network/5xx failures within configured bounds
- respects context cancellation

### Markdown Renderer

- renders video details when metadata is available
- omits empty metadata fields cleanly
- keeps existing language, generated-caption flag, snippet count, and source details
- preserves timestamp and plain transcript behavior
- keeps failed video rendering unchanged

### Reporter / Colored Feedback

- success, warning, failure, progress, saved, clipboard, and summary messages include meaningful text
- color-enabled mode includes expected ANSI sequences
- color-disabled mode has no ANSI sequences
- `NO_COLOR`, `TERM=dumb`, and non-TTY detection disable color

### CLI Help

- root help leads with clipboard-first workflow
- root help includes no-argument and advanced file-output examples
- `export --help` describes explicit file/batch export without contradicting default behavior

## E2E Tests

- test build with `-tags test` runs `yt-transcript-md` with fake clipboard env vars and verifies:
  - output file is created in working directory or configured temp directory
  - fake clipboard output receives the generated Markdown
  - Markdown includes metadata-enriched details from the test metadata provider
- root `--links --out` still writes file without clipboard env vars
- `export --links --out` still writes file
- strict failure still exits non-zero
- root `--help` and `export --help` contain the expected workflow language

## Verification Commands

- `go test ./...`
- `git diff --check`
- `make verify` when local lint/build tooling is available

## Manual Checks

- On macOS, copy a YouTube URL and run `go run ./cmd/yt-transcript-md`; verify `transcripts.md` is saved and clipboard contains Markdown.
- Run with `NO_COLOR=1` and verify output is readable without ANSI sequences.
- Run with duplicate comma-separated links and verify output has one video section.
