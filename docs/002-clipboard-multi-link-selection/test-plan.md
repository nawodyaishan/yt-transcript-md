# Test Plan: Clipboard Multi-Link Selection Prompt

## Unit Tests

### Selection Logic

- `all` returns every parsed unique video in order
- `one` with a valid index returns exactly that video
- `one` with zero, negative, or too-large input is rejected
- `recent` with a valid count returns the first N videos in order
- `recent` with zero, negative, or too-large count is rejected
- `cancel` returns no selected videos and a cancellation result
- duplicate videos are deduplicated before selection

### Clipboard Extraction

- extracts YouTube URLs from surrounding prose
- extracts raw 11-character video IDs from surrounding prose when separated by token boundaries
- preserves first-seen order
- strips surrounding angle brackets and punctuation
- deduplicates extracted videos before selection
- returns the existing no-valid-input error when no valid videos are present
- does not change explicit `--links`, `--input-file`, or `export` parsing behavior

### App Workflow

- one clipboard video bypasses selector and processes normally
- multiple clipboard videos invoke selector before metadata or transcript providers
- selector `all` processes every detected unique video
- selector `one` processes only the selected video
- selector `recent` processes only the selected first N videos
- selector cancellation causes no metadata fetches, transcript fetches, file writes, or clipboard writes
- selector error causes no network or output side effects
- reporter start count reflects selected videos only
- explicit `Export` with `--links` equivalent behavior does not invoke selector

### Terminal Prompt

- prompt shows a multi-video summary and available choices
- one-video choice lists indexed videos in parsed order
- invalid menu choice retries or returns a clear error
- invalid index retries or returns a clear error
- invalid recent count retries or returns a clear error
- EOF maps to cancellation
- prompt output does not include unrelated raw clipboard content

### CLI Flags and Help

- root help documents the multi-link clipboard prompt
- root help documents the non-interactive `--clipboard-selection` flag
- selection flag accepts `all`, `one:<index>`, and `recent:<count>`
- invalid selection flag values fail before network requests
- selection flag is rejected for explicit root/export workflows

## E2E Tests

- test build with fake clipboard containing one link runs without prompt and writes expected Markdown
- test build with fake clipboard containing multiple links and selection `all` writes all selected transcripts
- test build with fake clipboard containing multiple links and selection `one:2` writes only the second transcript
- test build with fake clipboard containing multiple links and selection `recent:2` writes only the first two transcripts
- test build with fake clipboard containing surrounding prose and two YouTube links prompts for or selects only the two valid videos
- non-interactive multi-video clipboard run without selection flag exits with a clear error before output files are created
- non-interactive multi-video clipboard run with `--clipboard-selection all`, `one:2`, and `recent:2` processes the selected subset
- root `--links --clipboard-selection all` and `export --clipboard-selection all` fail before output files are created
- root `--links` and `export --links` continue to process multiple links without prompting

## Verification Commands

- `go test ./...`
- `git diff --check`
- `make verify` when local lint/build tooling is available

## Manual Checks

- Copy one YouTube link and run `go run ./cmd/yt-transcript-md`; verify no prompt appears.
- Copy three YouTube links on separate lines and run `go run ./cmd/yt-transcript-md`; verify the prompt appears before any fetch progress.
- Choose each prompt path: one, all, recent N, and cancel.
- Run with redirected stdin/stdout or a scripted environment and verify the non-interactive behavior is clear.
