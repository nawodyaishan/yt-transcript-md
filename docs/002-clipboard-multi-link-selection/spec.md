# Feature Spec: Clipboard Multi-Link Selection Prompt

## Problem Statement

Users often copy several YouTube links at once from notes, search results, chat, or a browser page. The default no-argument workflow currently treats every valid unique video in the clipboard as the batch to process. That is efficient for intentional batches, but risky when the clipboard contains more videos than the user meant to export. Users need a single `yt-transcript-md` run to detect multi-video clipboard input and ask whether to process one video, all videos, or the first N videos from clipboard order before any transcript or metadata requests begin.

## Goals

- Detect when the no-argument clipboard workflow contains multiple unique YouTube videos.
- Ask the user which subset to process before fetching metadata or transcripts.
- Support processing one selected video, all detected videos, or the first N videos from clipboard order.
- Preserve existing non-interactive behavior for explicit `--links`, `--input-file`, and `export` workflows.
- Avoid accidental large YouTube request batches from clipboard content.
- Tolerate surrounding clipboard text by extracting valid YouTube videos without printing unrelated clipboard content.
- Keep the prompt clear, terminal-friendly, and testable without adding a new dependency.

## Non-Goals

- Do not add clipboard watching, background polling, or automatic queueing.
- Do not use YouTube publish dates to determine recency in this slice.
- Do not fetch metadata before the user selects the subset.
- Do not change explicit file or batch export semantics.
- Do not add playlist expansion, channel discovery, search, or browser integration.
- Do not add a terminal UI framework or new prompt dependency unless separately approved.

## Users or Actors

- Clipboard-first CLI users who copy several YouTube links and run `yt-transcript-md`.
- Users who want a quick transcript for only one video from a copied group.
- Users who copy a feed-like list and want only the first few videos in that list.
- Automation users who rely on existing non-interactive flags and must not be prompted.

## User Journeys

1. A user copies one YouTube URL and runs `yt-transcript-md`; the tool processes it without a prompt.
2. A user copies five YouTube URLs and runs `yt-transcript-md`; the tool reports five unique videos and asks whether to process one, all, or recent N.
3. A user chooses one video; the tool lists detected videos by clipboard order, accepts a number, and processes only that video.
4. A user chooses all videos; the tool processes the existing deduplicated clipboard batch in order.
5. A user chooses recent N; the tool asks for N and processes the first N unique videos in parsed clipboard order.
6. A user copies a paragraph that contains several YouTube URLs mixed with other text; the tool extracts the valid YouTube videos, prompts on those videos only, and never prints the unrelated text.
7. A user runs `yt-transcript-md --links ... --out notes.md`; the tool remains non-interactive and processes the provided links as it does today.
8. A user pipes or runs the command in a non-interactive terminal with multi-video clipboard content; the tool fails clearly unless an explicit selection mode is provided.

## Functional Requirements

- FR-001: The default no-argument clipboard workflow must parse and deduplicate clipboard YouTube videos before prompting.
- FR-002: If the clipboard contains exactly one unique video, the command must process that video without prompting.
- FR-003: If the clipboard contains multiple unique videos and stdin/stdout are interactive, the command must prompt before metadata or transcript fetches.
- FR-004: The prompt must offer at least three choices: one video, all videos, and recent N videos.
- FR-005: The one-video path must present detected videos in clipboard order with stable numeric indexes.
- FR-006: The one-video path must accept a valid numeric index and reject out-of-range or non-numeric input with a clear retry or cancellation path.
- FR-007: The all-videos path must process every unique parsed video in clipboard order.
- FR-008: The recent-N path must ask for a positive integer N and process the first N unique videos in parsed clipboard order.
- FR-009: If N is greater than the number of unique videos, the command must reject it with a clear message rather than silently processing all videos.
- FR-010: The prompt must include a cancel option that exits before network requests and does not write an output file or clipboard content.
- FR-011: No metadata or transcript request may start until the selection has been resolved.
- FR-012: The final reporter start message must count only selected videos, not all detected clipboard videos.
- FR-013: Duplicate clipboard links must not appear more than once in the selectable list.
- FR-014: Clipboard mode must extract valid YouTube videos from surrounding text while preserving first-seen order.
- FR-015: Explicit root flags `--links` and `--input-file` must remain non-interactive.
- FR-016: The `export` subcommand must remain non-interactive.
- FR-017: Clipboard mode must fail with `no valid YouTube links or video IDs provided` when no valid YouTube videos can be extracted.
- FR-018: The prompt must not print raw unrelated clipboard text.
- FR-019: The command must provide a root-only `--clipboard-selection` flag for non-interactive multi-video clipboard input.
- FR-020: `--clipboard-selection` must accept exactly `all`, `one:<index>`, and `recent:<count>` values.
- FR-021: Invalid `--clipboard-selection` values must fail before metadata fetches, transcript fetches, file writes, or clipboard writes.
- FR-022: `--clipboard-selection` must apply only to default clipboard mode and must be rejected with a clear error when explicit `--links`, `--input-file`, or `export` input is used.
- FR-023: README and root help must document the multi-link clipboard prompt and non-interactive selection flag.
- FR-024: Tests must cover single-video no-prompt, multi-video prompt choices, invalid prompt input, cancellation, non-interactive behavior, clipboard text extraction, and unchanged explicit batch behavior.

## Acceptance Criteria

- AC-001: With a clipboard containing one YouTube URL, `yt-transcript-md` does not prompt and produces the same save-and-copy output as today.
- AC-002: With a clipboard containing three unique YouTube URLs, choosing all processes exactly those three videos in parsed order.
- AC-003: With a clipboard containing three unique YouTube URLs, choosing one and index `2` processes only the second unique video.
- AC-004: With a clipboard containing five unique YouTube URLs, choosing recent N and entering `2` processes only the first two unique videos in parsed order.
- AC-005: Duplicate links in the clipboard are deduplicated before the prompt and before any fetches.
- AC-006: Canceling the prompt exits without metadata fetches, transcript fetches, file writes, or clipboard writes.
- AC-007: In a non-interactive run with multi-video clipboard input and no explicit selection mode, the command exits with a clear error before network requests.
- AC-008: Explicit `--links`, `--input-file`, and `export` workflows process batches without prompting.
- AC-009: With a clipboard containing unrelated prose and two YouTube URLs, clipboard mode prompts for exactly the two valid unique videos and does not print the unrelated prose.
- AC-010: In a non-interactive run with multi-video clipboard input and `--clipboard-selection recent:2`, the command processes exactly the first two unique videos in parsed order.
- AC-011: Invalid `--clipboard-selection` values fail before network or output side effects.
- AC-012: Combining `--clipboard-selection` with explicit `--links`, `--input-file`, or `export` input fails with a clear error before network or output side effects.
- AC-013: `go test ./...` passes with unit and e2e coverage for the prompt workflow.
- AC-014: README and root help explain how multi-link clipboard input is handled.

## Success Criteria

- Users do not accidentally process large clipboard batches with the default command.
- Users can intentionally process one, all, or a bounded recent subset without re-copying links.
- Automation and explicit batch exports remain stable and non-interactive.
- Request volume remains limited to selected unique videos.

## Edge Cases

- Clipboard contains only duplicate links.
- Clipboard contains a mix of valid YouTube links, unsupported URLs, and unrelated text.
- Clipboard contains no valid YouTube videos.
- Clipboard contains links separated by newlines, commas, angle brackets, or surrounding punctuation.
- User enters an invalid menu choice.
- User enters an invalid video index.
- User enters zero, a negative number, or a number larger than the detected video count for recent N.
- User sends EOF or interrupt during the prompt.
- Stdin is not interactive, stdout is redirected, or both.
- Metadata or transcript fetch later fails for a selected video.
- Selected subset contains videos that would otherwise be deduplicated.

## Data Sensitivity and Compliance Notes

- Clipboard content may include private unrelated text. The prompt must display only parsed YouTube video identifiers and sanitized source URLs.
- The tool must not log the full raw clipboard content.
- No network request should be made for unselected videos.
- No credentials, cookies, OAuth tokens, or YouTube Data API keys are introduced.

## API or Integration Expectations

- The existing `input.ParseVideoInputs` behavior should remain the canonical parser and deduplication layer unless the plan identifies a parser gap.
- Clipboard mode needs an extraction layer before canonical parsing so prose surrounding valid links does not make the whole clipboard invalid.
- Selection should operate on `models.VideoInput` values after parsing.
- The app layer should own workflow decisions, while CLI wiring should provide interactive input/output streams and non-interactive detection.
- Production prompt handling should use standard input/output primitives or a tiny local abstraction.

## Assumptions

- "Recent N" means the first N unique videos in parsed clipboard order, not YouTube publish date.
- Clipboard order is the only available ordering signal before metadata or transcript fetches.
- The first implementation can show video IDs and source URLs; video titles are unavailable until after selection.
- Non-interactive safety is required because the root command can be used in scripts even when it reads the clipboard.

## Deferred Questions

- Does "recent N" need to mean newest by YouTube publish date in a future slice? That would require metadata requests before final selection and is intentionally out of scope here.

## Human Approval Status

Approved by user on 2026-06-25. Architecture approval was also granted on 2026-06-25; implementation may proceed within the approved plan and task breakdown.
