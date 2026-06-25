# Test Plan: Clipboard History TUI Selection

## Unit Tests

### Provider Detection

- detects `copyq` when available
- detects `cliphist` when available
- detects `gpaste-client` when available
- treats missing provider as unavailable
- honors explicit provider source
- auto mode skips unavailable providers

### Provider Parsing

- CopyQ adapter parses recent text entries from fixture output
- cliphist adapter parses list/decode fixture output
- GPaste adapter parses indexed/raw history fixture output
- provider adapter respects configured history limit
- provider adapter times out or returns error on command failure
- provider adapter never logs raw entry text in errors

### Candidate Aggregation

- current clipboard candidates are included first
- provider candidates are included after current clipboard candidates
- duplicate videos across providers dedupe by video ID
- unrelated clipboard-history text is ignored
- sanitized previews are truncated
- empty sources return no candidates
- `--no-history` disables provider calls
- `--history-source current` disables provider calls

### TUI Selection

- keyboard navigation moves cursor
- search filters rows
- space toggles selected row
- `a` selects visible rows
- `n` selects first N visible rows
- Enter confirms selected rows
- Enter with zero selected rows shows an error or does not exit
- Esc/q cancels
- cancellation returns no selected videos

### App Workflow

- one candidate processes without TUI
- multiple candidates invoke selector before metadata/transcript providers
- selected candidates only trigger selected provider calls
- cancellation prevents metadata fetches, transcript fetches, file writes, and clipboard writes
- explicit `--links`, `--input-file`, and `export` do not invoke history providers or TUI selector

### CLI Flags and Help

- root help documents `--history-source`, `--history-limit`, `--no-history`, and TUI behavior
- invalid `--history-limit` fails before provider calls
- invalid `--history-source` fails before provider calls
- history flags are rejected with explicit root input
- history flags are unavailable or rejected for `export`

## E2E Tests

- fake history provider with three YouTube links opens fake selector and processes selected videos
- fake history provider with duplicates processes only unique selected videos
- fake history provider cancellation creates no output file
- no provider and single current clipboard link preserves current workflow
- non-interactive multi-candidate history run requires explicit selection
- root `--links` and `export --links` do not invoke fake history provider
- help text includes supported history flags and provider names

## Verification Commands

- `go test ./...`
- `git diff --check`
- `make verify`

## Manual Checks

- macOS with CopyQ running: copy several YouTube links one by one, run `yt-transcript-md`, select from TUI, verify only selected videos are exported.
- Linux Wayland with cliphist configured: copy several YouTube links one by one, run `yt-transcript-md --history-source cliphist`, verify TUI selection.
- Linux GNOME with GPaste running: run `yt-transcript-md --history-source gpaste`, verify provider history is read.
- Run `yt-transcript-md --no-history` and verify only current clipboard behavior is used.
