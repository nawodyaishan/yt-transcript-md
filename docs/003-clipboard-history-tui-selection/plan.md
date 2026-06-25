# Technical Plan: Clipboard History TUI Selection

## Summary

Add clipboard-history scanning to the default no-argument workflow using a provider architecture. The command will read the current clipboard first, then optionally query supported clipboard-history providers on macOS/Linux, extract YouTube video candidates, and present a Bubble Tea TUI picker when multiple unique videos are available. Existing explicit input workflows remain non-interactive and do not scan clipboard history.

## Inputs Reviewed

- `docs/001-default-save-rich-feedback/`
- `docs/002-clipboard-multi-link-selection/`
- `internal/app/export.go`
- `internal/input/parser.go`
- `internal/cli/export.go`
- Exa research:
  - CopyQ command-line/scripting support: `https://hluk.github.io/CopyQ/`
  - cliphist CLI history design: `https://github.com/sentriz/cliphist`
  - GPaste CLI client: `https://man.archlinux.org/man/gpaste-client.1.en`
  - Bubble Tea and Bubbles TUI libraries: `https://github.com/charmbracelet/bubbletea`, `https://github.com/charmbracelet/bubbles`
  - Maccy/Raycast/Alfred clipboard history documentation for non-first-slice provider decisions

## Assumptions

- User approved macOS and Linux as target platforms.
- User wants existing clipboard-manager integration, not a `yt-transcript-md` watcher.
- User wants a full TUI with search and multi-select.
- User wants configurable history retention/scan limit.
- CopyQ is the most practical first macOS provider because it has a CLI and supports macOS/Linux.
- Linux support should include cliphist and GPaste because both expose CLI history.

## Architecture Approach

### 1. Keep Current Clipboard as Source Zero

The existing clipboard reader remains the first source. This preserves simple one-link usage and avoids requiring a clipboard manager for basic use.

### 2. Add Clipboard History Provider Boundary

Add `internal/history` with interfaces like:

```go
type Provider interface {
    Name() string
    Available(ctx context.Context) error
    Entries(ctx context.Context, limit int) ([]Entry, error)
}

type Entry struct {
    Provider string
    ID string
    Text string
    Preview string
    Rank int
}
```

Provider implementations:

- `CopyQProvider`
- `CliphistProvider`
- `GPasteProvider`

Use `exec.CommandContext` with direct args and timeouts. Do not invoke through shell strings.

### 3. Add Candidate Aggregation

Add a candidate aggregator that:

1. reads current clipboard text
2. queries selected history provider(s) when enabled
3. extracts YouTube candidates using the existing clipboard extraction logic
4. dedupes by canonical video ID
5. records source provider, source rank, and sanitized preview

This should be pure/testable outside CLI code.

### 4. Add TUI Selector Boundary

Add an app-level selector interface that can be implemented by a Bubble Tea TUI in CLI wiring and faked in tests.

The selected output should be `[]models.VideoInput` or an existing selection type that can be applied before transcript/metadata fetches.

### 5. Use Bubble Tea and Bubbles for TUI

Recommended dependencies:

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- likely `github.com/charmbracelet/lipgloss` if styling is needed through Bubbles

TUI features:

- list rows with video ID, source provider, and sanitized preview
- `/` or direct typing for search
- arrow keys/j/k for navigation
- space toggles selected row
- `a` toggles/selects all visible rows
- `n` opens first-N input
- Enter confirms selected rows
- Esc/q cancels

### 6. Add CLI Flags and Config Surface

Root-only flags:

- `--history-source auto|current|copyq|cliphist|gpaste`
- `--history-limit N`
- `--no-history`

Existing `--clipboard-selection` should continue to work as a non-TUI escape hatch for non-interactive runs.

Non-interactive mode should not scan history by default. If the user explicitly provides both a history source and a non-TUI selection flag, the command may scan history without opening the TUI.

Default values:

- `history-source=auto`
- `history-limit=50`
- `no-history=false`

### 7. Preserve Explicit Workflow Boundaries

If `--links`, `--input-file`, or `export` is used:

- do not scan history
- do not open TUI
- reject history flags if they would create ambiguity, or document them as ignored with a warning

Recommendation: reject root history flags when combined with explicit root input and keep them unavailable on `export`.

## Affected Modules

- `internal/history/`: new providers, detection, aggregation, tests.
- `internal/app/export.go`: accept selected history candidates before fetch/render.
- `internal/app/selection.go`: evolve selection model if needed for multi-select candidate metadata.
- `internal/input/parser.go`: reuse or extend clipboard extraction helpers.
- `internal/cli/export.go`: add root flags and provider/TUI wiring.
- `internal/cli/`: new TUI selector implementation.
- `internal/cli/provider_test_mode.go`: fake history providers/selectors for e2e tests.
- `README.md`: document provider setup, flags, and TUI controls.
- `tests/e2e/e2e_test.go`: cover fake provider history and default workflow behavior.
- `go.mod`, `go.sum`: dependency additions if Bubble Tea/Bubbles are approved.

## API and Contract Changes

- Default no-argument mode may scan clipboard history when interactive and history scanning is enabled.
- Current clipboard behavior remains supported and first-priority.
- Explicit input modes remain non-interactive.
- Root help gains history flags and provider documentation.
- New dependency approval is required for TUI libraries.

## Data Model Changes

See `data-model.md`.

## Dependency Changes

Planned pending approval:

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`

Possible transitive dependency:

- `github.com/charmbracelet/lipgloss`

No provider SDK dependencies are planned. Provider access should use local CLI commands.

## Security Impact

- Clipboard history is sensitive. Never log raw entries.
- Only sanitized/truncated previews should be shown in the TUI.
- Provider command paths should be discovered with `exec.LookPath`.
- Provider commands should run with bounded context timeouts.
- No history content is persisted by `yt-transcript-md`.

## Observability Impact

- Reporter should show selected video count only after TUI confirmation.
- Provider warnings should name the provider but not raw entry content.
- In `auto` mode, skipped/missing providers should be quiet unless no usable source remains, or shown at verbose level if a verbose flag is later added.

## Testing Strategy

See `test-plan.md`.

Verification commands:

- `go test ./...`
- `git diff --check`
- `make verify`

## Failure Modes

- No provider available: fall back to current clipboard if usable, otherwise clear error.
- Explicit provider missing/failing: clear fatal error.
- Auto provider failing: warning and continue to next provider.
- TUI cancellation: no network or output side effects.
- Non-interactive terminal with multiple candidates: require `--clipboard-selection` or explicit input flags.
- Dependency rendering issues: tests should fake selector at app boundary and keep TUI unit tests focused.

## Risks and Mitigations

- Risk: macOS clipboard managers do not expose stable CLI history.
  Mitigation: support CopyQ first on macOS and defer private app integrations.

- Risk: Provider output formats vary.
  Mitigation: isolate adapters, parse conservatively, and add command-output fixtures.

- Risk: TUI dependency increases project size.
  Mitigation: require explicit dependency approval and keep TUI isolated behind an interface.

- Risk: Clipboard history leaks sensitive content.
  Mitigation: local-only extraction, no raw logs, truncated previews, no persistence.

## Human Architecture Approval Status

Approved by user on 2026-06-25. Dependency approval for Bubble Tea/Bubbles was also granted on 2026-06-25. Task execution may proceed within this plan.
