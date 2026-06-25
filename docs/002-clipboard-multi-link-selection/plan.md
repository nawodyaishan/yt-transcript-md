# Technical Plan: Clipboard Multi-Link Selection Prompt

## Summary

Add a selection step to the default no-argument clipboard workflow. The command will continue to parse and deduplicate clipboard input first. If exactly one unique video is present, it proceeds unchanged. If multiple unique videos are present, it resolves a user selection before any metadata or transcript fetches. Explicit `--links`, `--input-file`, and `export` workflows remain non-interactive and unchanged.

## Inputs Reviewed

- `docs/001-default-save-rich-feedback/spec.md`
- `docs/001-default-save-rich-feedback/plan.md`
- `internal/app/export.go`
- `internal/app/export_test.go`
- `internal/input/parser.go`
- `internal/cli/export.go`
- `internal/app/reporter.go`
- Exa MCP search results for CLI clipboard detection and batch selection patterns

## Assumptions

- "Recent N" means first N unique videos in parsed clipboard order.
- No video title is available at prompt time unless metadata is fetched early; this plan avoids early metadata fetches.
- Root no-argument mode may prompt because it is human-oriented.
- Explicit input flags and `export` remain automation-oriented and must not prompt.
- No new runtime dependency is needed for prompt handling.
- Clipboard mode should extract valid YouTube videos from surrounding prose before passing candidates through canonical parsing and deduplication.

## Architecture Approach

### 1. Split Clipboard Parsing from Export Rendering

Keep `renderExport` responsible for fetching and rendering selected videos. Add a helper that accepts already-parsed `[]models.VideoInput`, avoiding a second parse after the user chooses a subset.

Candidate shape:

- `renderVideos(ctx, videos, opts, provider, metadataProvider, reporter)`
- `renderExport(ctx, rawInput, opts, provider, metadataProvider, reporter)` parses then delegates to `renderVideos`

This preserves explicit export behavior and gives clipboard mode a clean place to select a subset.

### 2. Add Clipboard Extraction Before Canonical Parsing

Clipboard mode should scan raw clipboard text for YouTube URLs and raw 11-character video IDs, normalize surrounding punctuation, and preserve first-seen order. The extracted candidates should then use the existing canonical parser/deduper so URL support remains centralized.

Explicit `--links`, `--input-file`, and `export` inputs should keep the current strict parse behavior unless a separate spec changes them.

### 3. Add a Selection Abstraction

Add an app-level selector interface so tests can drive decisions without terminal input.

Candidate shape:

```go
type ClipboardSelection struct {
    Mode ClipboardSelectionMode
    Index int
    Count int
}

type ClipboardSelector interface {
    Select(videos []models.VideoInput) (ClipboardSelection, error)
}
```

Modes:

- `all`
- `one`
- `recent`
- `cancel`

The selected subset should be computed by a pure function, for example `ApplyClipboardSelection(videos, selection)`, so validation can be unit tested separately from terminal prompting.

### 4. Implement a Minimal Terminal Prompt

Production CLI should provide a standard-library prompt implementation using readable menu text and `bufio.Reader`.

Prompt behavior:

- summarize detected count
- list numeric choices for one/all/recent/cancel
- list videos by index only when one-video mode is selected
- retry invalid menu/index/count input with a bounded or cancelable loop
- treat EOF as cancellation

The prompt should display only parsed video IDs and sanitized original URLs.

### 5. Add Non-Interactive Safety

If multiple videos are detected and prompt input/output is not interactive, default clipboard mode should fail before network requests unless an explicit selection mode is provided.

Required root clipboard selection flag:

- `--clipboard-selection all`
- `--clipboard-selection one:<index>`
- `--clipboard-selection recent:<count>`

This flag should apply only to root clipboard mode. Explicit `--links`, `--input-file`, and `export` should not require or use it.

### 6. Preserve Request Discipline

Selection must happen before the existing metadata and transcript loops. Reporter output should use selected counts. Deduplication remains parser-owned, so no unselected or duplicate video should trigger network requests.

### 7. Update Docs and Help

Update README and root help to explain:

- one copied link processes immediately
- multiple copied links trigger a selection prompt
- automation can use explicit `--links`/`--input-file` or the root clipboard selection flag

## Affected Modules

- `internal/app/export.go`: parse clipboard input, invoke selector, render selected videos.
- `internal/app/export_test.go`: cover selection, cancellation, and request counts.
- `internal/app/`: optional new `selection.go` for selection models and pure validation.
- `internal/cli/export.go`: wire terminal selector and optional root selection flag.
- `internal/cli/root.go`: help text updates.
- `internal/cli/provider_test_mode.go`: test selector or flag support for e2e tests if needed.
- `internal/input/parser.go` or new `internal/input` helper: clipboard extraction before canonical parsing.
- `README.md`: document multi-link clipboard workflow.
- `tests/e2e/e2e_test.go`: verify prompt and non-interactive flag behavior.

## API and Contract Changes

- Default root clipboard mode becomes interactive only when multiple unique videos are detected.
- Single-video clipboard mode remains behaviorally compatible.
- Explicit batch modes remain non-interactive.
- New root-only `--clipboard-selection` flag is added for non-interactive clipboard runs.
- `--clipboard-selection` is rejected when combined with explicit `--links`, `--input-file`, or `export` input.
- Clipboard mode accepts surrounding text when at least one valid YouTube video can be extracted.
- No public Go API is exposed by the module.

## Data Model Changes

See `data-model.md`.

## Dependency Changes

No dependency changes are planned.

Alternatives considered:

- Add `survey` or another prompt package: rejected for this slice because the menu is simple.
- Fetch metadata before prompting to show titles: rejected because it violates the request discipline goal.
- Use publish date for recent N: rejected for this slice because it needs extra metadata/API behavior before selection.
- Process all videos automatically in non-interactive clipboard mode: rejected because it preserves the accidental-large-batch risk this feature is meant to reduce.

## Security Impact

- Clipboard content remains local except selected video IDs are sent to existing metadata/transcript providers.
- Prompt output must not echo unrelated raw clipboard content.
- Selection flag values are local CLI input only.

## Authorization Boundaries

- No authentication or authorization is added.
- No cookies, OAuth tokens, or API keys are introduced.

## Observability Impact

- Add clear prompt messages before the existing reporter start.
- Existing reporter summary should count selected videos.
- Cancellation should print a short message and exit without a misleading success summary.

## Testing Strategy

See `test-plan.md`.

Verification commands:

- `go test ./...`
- `git diff --check`
- optionally `make verify`

## Failure Modes

- Multi-video clipboard input in non-interactive mode without selection flag: clear error before network requests.
- Invalid selection flag value: clear error before network requests.
- Selection flag combined with explicit input: clear error before network requests.
- User cancellation/EOF: exit before network requests, file writes, or clipboard writes.
- Invalid prompt input: retry or allow cancellation.
- Clipboard contains prose but no valid YouTube videos: existing no-valid-input error before network requests.
- Selected transcript fails later: existing partial/strict failure behavior applies.

## Risks and Mitigations

- Risk: Prompt breaks users who expected automatic full-batch clipboard processing.
  Mitigation: provide `all` in the prompt and a non-interactive `--clipboard-selection all` escape hatch.

- Risk: "Recent" is misunderstood as publish date.
  Mitigation: label it as "recent from clipboard order" in help and prompt text.

- Risk: Prompt text leaks unrelated clipboard content.
  Mitigation: render only parsed `VideoInput` values.

- Risk: Terminal detection is unreliable across platforms.
  Mitigation: keep explicit selection flags and test both interactive and non-interactive paths.

## Human Architecture Approval Status

Approved by user on 2026-06-25. Task execution may proceed within this plan.
