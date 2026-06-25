# Feature Spec: Clipboard History TUI Selection

## Problem Statement

The current default workflow reads only the current clipboard item. That works when a user copies one YouTube link or copies several links together, but it does not support the common workflow where a user copies several YouTube links one at a time from a browser, chat, or notes app. Users need `yt-transcript-md` to scan recent clipboard history from supported clipboard managers on macOS and Linux, find recently copied YouTube videos, and present a searchable multi-select TUI before fetching transcripts.

## Relationship to Previous Specs

- `docs/001-default-save-rich-feedback/` remains the baseline save-and-copy workflow.
- `docs/002-clipboard-multi-link-selection/` handles multiple videos in the current clipboard text.
- This spec extends default clipboard mode to optionally read recent clipboard history entries from supported clipboard managers. It supersedes `002` only for the source of selectable links; current-clipboard extraction remains the first input source and fallback.

## Goals

- Detect supported clipboard-history providers on macOS and Linux.
- Read recent text clipboard history entries without requiring a `yt-transcript-md` watcher.
- Extract and deduplicate YouTube videos across current clipboard text and provider history entries.
- Show a full-screen TUI with arrow-key navigation, search/filtering, and multi-select.
- Let users select exact videos, all filtered videos, or a first-N subset before network requests.
- Keep history scanning configurable by source and limit.
- Preserve explicit `--links`, `--input-file`, and `export` behavior as non-interactive workflows.
- Avoid logging or printing unrelated clipboard-history contents.

## Non-Goals

- Do not implement a background clipboard watcher in this slice.
- Do not read macOS private databases for apps that do not expose a stable CLI/API.
- Do not support Windows clipboard history in this slice.
- Do not process image, HTML-only, rich text, file, or binary clipboard entries except for text representations exposed by providers.
- Do not fetch YouTube metadata before the user selects videos.
- Do not add playlist expansion, channel discovery, browser automation, cookies, OAuth, or YouTube Data API usage.
- Do not make a clipboard manager a hard dependency for explicit export workflows.

## Users or Actors

- macOS users who use a clipboard manager and copy several YouTube links over time.
- Linux users on Wayland or GNOME who use clipboard-history tools such as `cliphist`, `CopyQ`, or `GPaste`.
- CLI users who prefer keyboard-driven selection from recent copied links.
- Automation users who rely on existing explicit non-interactive inputs.

## Supported Provider Strategy

Provider priority for `--history-source auto`:

1. Current clipboard text through the existing clipboard reader.
2. `copyq` when installed and running.
3. `cliphist` on Linux Wayland when installed.
4. `gpaste-client` on Linux GNOME when installed.

Provider support notes:

- `CopyQ` is the preferred cross-platform provider for macOS and Linux because it has a documented command-line/scripting interface.
- `cliphist` is preferred on Linux Wayland because it is CLI-native and designed for pipe-based history recall.
- `GPaste` is supported for GNOME because it exposes history through `gpaste-client`.
- Maccy, Raycast, Alfred, and Paste are not first-slice providers unless a stable public CLI/API is identified; their GUI history features are useful but not enough for reliable CLI integration.

## User Journeys

1. A user copies one YouTube link and runs `yt-transcript-md`; the tool processes the current clipboard link immediately as today.
2. A user copies five YouTube links one by one, then runs `yt-transcript-md`; the tool scans supported clipboard history and opens a TUI with the five unique videos.
3. A user types in the TUI search box to filter videos by URL, video ID, source provider, or copied entry preview.
4. A user selects multiple videos with the keyboard and presses Enter; the tool fetches only selected videos.
5. A user presses `a` to select all currently filtered videos and then Enter to process them.
6. A user presses `n`, enters `3`, and processes the first three filtered videos in history order.
7. A user presses Esc or `q`; the command exits before network requests, file writes, or clipboard writes.
8. A user runs `yt-transcript-md --history-source copyq --history-limit 25`; the tool reads up to 25 CopyQ entries and skips other providers.
9. A user runs `yt-transcript-md --no-history`; the tool only uses current clipboard behavior from `002`.
10. A user runs `yt-transcript-md export --links ...`; the tool remains non-interactive and does not scan clipboard history.

## Functional Requirements

- FR-001: Default no-argument mode must read the current clipboard text first.
- FR-002: Default no-argument mode must scan supported clipboard-history providers when history scanning is enabled.
- FR-003: History scanning must be enabled by default only for interactive terminals.
- FR-003a: Non-interactive default mode must not scan history unless the user provides explicit history and selection flags that avoid opening the TUI.
- FR-004: If exactly one unique YouTube video is found across enabled sources, the command may process it without opening the TUI.
- FR-005: If multiple unique videos are found across enabled sources, the command must open a TUI before metadata or transcript fetches.
- FR-006: The TUI must support keyboard navigation.
- FR-007: The TUI must support text search/filtering.
- FR-008: The TUI must support multi-select.
- FR-009: The TUI must support select all visible/filtered videos.
- FR-010: The TUI must support selecting the first N visible/filtered videos.
- FR-011: The TUI must support cancellation that exits before network and output side effects.
- FR-012: TUI selection must happen before metadata or transcript requests.
- FR-013: Only selected unique videos may trigger metadata or transcript requests.
- FR-014: Deduplication must use canonical video ID across all providers and current clipboard text.
- FR-015: First-seen order must be preserved, with current clipboard entries before history entries.
- FR-016: Each selectable row must show enough context to distinguish videos without fetching metadata: video ID, sanitized source URL or preview, and source provider.
- FR-017: The command must provide `--history-source` with values `auto`, `current`, `copyq`, `cliphist`, and `gpaste`.
- FR-018: The command must provide `--history-limit` as a positive integer.
- FR-019: The command must provide `--no-history` to disable provider scanning and use only current clipboard behavior.
- FR-020: `--clipboard-selection` from `002` must continue to work for non-interactive selection when multiple current/history videos are available.
- FR-021: Explicit `--links`, `--input-file`, and `export` workflows must not scan clipboard history or open the TUI.
- FR-022: If history scanning is requested but no supported provider is available, the command must clearly report that and fall back to current clipboard behavior when possible.
- FR-023: Provider command failures must produce warnings and continue to the next provider in `auto` mode.
- FR-024: Provider command failures must be fatal when the user explicitly selects that provider with `--history-source`.
- FR-025: The tool must not log raw clipboard-history entries.
- FR-026: README and CLI help must document supported providers, required setup, flags, and privacy behavior.
- FR-027: Tests must cover provider detection, provider parsing, deduplication, TUI selection behavior, cancellation, and explicit workflow compatibility.

## Acceptance Criteria

- AC-001: With only one current clipboard YouTube link and no history provider, `yt-transcript-md` behaves as the current default workflow.
- AC-002: With CopyQ test history containing three unique YouTube videos, `yt-transcript-md --history-source copyq --history-limit 10` opens the TUI and processes only selected videos.
- AC-003: With cliphist test history containing duplicate video links, the selectable list shows one row per unique video.
- AC-004: With GPaste test history containing valid links mixed with unrelated text, the selectable list includes only valid YouTube videos and does not render unrelated raw text outside sanitized previews.
- AC-005: Canceling the TUI exits without metadata fetches, transcript fetches, file writes, or clipboard writes.
- AC-006: Selecting all visible rows processes exactly all currently filtered unique videos.
- AC-007: Selecting first N visible rows processes exactly the first N filtered unique videos.
- AC-008: `--history-limit 5` reads no more than five entries from each selected provider.
- AC-009: `--history-source current` and `--no-history` do not invoke provider commands.
- AC-010: Explicit `--links`, `--input-file`, and `export` workflows do not invoke provider commands or TUI code.
- AC-011: `--history-source copyq` fails clearly when `copyq` is missing or not running.
- AC-012: `--history-source auto` warns and falls back when a provider command fails.
- AC-013: `go test ./...` passes with unit and e2e coverage for provider and selector behavior.
- AC-014: README and root help explain clipboard-history scanning and supported providers.

## Success Criteria

- A user can copy several YouTube links one by one and process selected recent links with a single `yt-transcript-md` run.
- The default workflow remains simple for one current clipboard link.
- The tool never fetches transcripts for unselected history videos.
- Users can choose and configure their clipboard-history source.
- The implementation remains testable without requiring real clipboard-manager installations in CI.

## Edge Cases

- No supported clipboard-history provider is installed.
- Provider is installed but not running.
- Provider history contains no text entries.
- Provider history contains sensitive unrelated text.
- Provider history contains binary or non-text entries.
- Provider history contains duplicate YouTube links or duplicate raw video IDs.
- Current clipboard link also appears in provider history.
- Provider command is slow, hangs, or returns malformed output.
- `--history-limit` is zero, negative, or too large.
- Terminal is not interactive.
- TUI is interrupted with Ctrl+C, Esc, or EOF.
- TUI search filters to zero rows.
- User selects zero rows and presses Enter.
- Selected videos later have transcript or metadata failures.

## Data Sensitivity and Compliance Notes

- Clipboard history may contain private or sensitive information. The tool must extract video candidates locally and must not log full raw history entries.
- Provider adapters should request text content only when possible.
- Sanitized previews must be truncated and must not include unrelated large clipboard text.
- The tool must not write clipboard-history data to project output except selected YouTube source URLs already required for transcript Markdown.
- No telemetry, analytics, remote logging, credentials, cookies, OAuth tokens, or API keys are introduced.

## API or Integration Expectations

- Provider adapters should live behind an internal interface.
- Provider adapters should shell out only to known commands with bounded arguments; no shell interpolation is required.
- Provider output parsing should be structured where possible and line-oriented only when provider output format is documented.
- TUI selection should operate on already extracted `models.VideoInput` candidates plus local source metadata.
- Transcript and metadata providers should remain unchanged and should receive only selected `models.VideoInput` values.

## Assumptions

- CopyQ, cliphist, and GPaste are acceptable first-slice providers for macOS/Linux.
- CopyQ must be installed and running for its CLI history access to work.
- cliphist users already have `wl-paste --watch cliphist store` or equivalent configured.
- GPaste users have the GPaste daemon available.
- Adding Bubble Tea/Bubbles dependencies is acceptable only after dependency approval.
- Clipboard history order means newest-to-oldest as returned by the provider, unless a provider documents a different default.

## Deferred Questions

- Should Maccy, Raycast, Alfred, or Paste be supported through optional user-provided commands in a later slice?
- Should a future `yt-transcript-md watch` mode be added for users without a clipboard manager?
- Should provider history be merged newest-first globally by timestamp when providers expose timestamps?

## Human Approval Status

Approved by user on 2026-06-25. Architecture and dependency approval were also granted on 2026-06-25; implementation may proceed within the approved plan and task breakdown.
