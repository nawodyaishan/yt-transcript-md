# Feature Spec: Positional URL Arguments

## Problem Statement

The current CLI requires users to pass YouTube links through the `--links` flag with comma separation, which is syntactically awkward for quick one-off fetches. Users expect to be able to type links directly after the command name, as they would with tools like `curl`, `wget`, or `ffmpeg`. The shell already separates arguments by space, so requiring a comma-delimited flag value adds unnecessary friction and makes the command harder to use from shell history recall, browser copy-paste, and shell tab completion workflows.

## Relationship to Previous Specs

- `docs/001-default-save-rich-feedback/` establishes the default no-argument clipboard workflow. That workflow is unchanged.
- `docs/002-clipboard-multi-link-selection/` handles multi-link clipboard input. Positional URLs bypass the clipboard entirely.
- `docs/003-clipboard-history-tui-selection/` adds clipboard history scanning. Positional URLs also bypass history scanning.
- This spec adds a third input mode alongside the existing clipboard and explicit `--links` modes.

## Goals

- Allow one or more YouTube links or video IDs to be passed as positional arguments to both the root command and the `export` subcommand.
- Keep the existing `--links`, `--input-file`, and clipboard workflows unchanged.
- Reject ambiguous combinations that mix positional arguments with `--links` or `--input-file`.
- Require no new dependencies.

## Non-Goals

- Do not change the output format, file path logic, or transcript fetch behavior.
- Do not add shell completion integration in this slice.
- Do not allow positional arguments alongside clipboard or history flags (`--clipboard-selection`, `--history-source`, `--history-limit`, `--no-history`); those flags only apply to the clipboard workflow.
- Do not support bare video IDs that look like filesystem paths (e.g. `./dQw4w9WgXcQ`).
- Do not add playlist expansion or channel URLs in this slice.

## Users or Actors

- CLI users who copy one or more YouTube links from a browser and paste them as arguments after the command name.
- Power users who script or alias `yt-transcript-md` and prefer space-separated URL lists.
- Shell history users who recall and replay commands with different links.

## User Journeys

1. A user copies a single YouTube link and runs `yt-transcript-md https://youtu.be/dQw4w9WgXcQ`; the tool fetches and saves the transcript without opening a TUI or reading the clipboard.
2. A user copies two links and runs `yt-transcript-md https://youtu.be/abc https://youtu.be/def`; the tool fetches both and appends them to `transcripts.md`.
3. A user runs `yt-transcript-md export https://youtu.be/abc --out notes.md`; the tool writes only that transcript to `notes.md`.
4. A user mixes positional arguments with `--links` and receives a clear error.
5. A user mixes positional arguments with `--input-file` and receives a clear error.
6. A user passes positional arguments alongside clipboard-only flags (`--history-source`, etc.) and receives a clear error.
7. A user passes no arguments and no flags; the clipboard workflow runs as today.

## Functional Requirements

- FR-001: The root command must accept one or more positional arguments that are YouTube URLs or bare video IDs.
- FR-002: The `export` subcommand must accept one or more positional arguments that are YouTube URLs or bare video IDs.
- FR-003: When positional arguments are present, the command must not read the clipboard, scan clipboard history, or open the TUI.
- FR-004: When positional arguments are present, the command must not write back to the clipboard.
- FR-005: Positional arguments must be validated as YouTube links or video IDs using the same parser as `--links`.
- FR-006: If any positional argument is not a valid YouTube link or video ID, the command must fail with a clear error before making any network requests.
- FR-007: Positional arguments combined with `--links` must produce a clear error.
- FR-008: Positional arguments combined with `--input-file` must produce a clear error.
- FR-009: Positional arguments combined with `--clipboard-selection`, `--history-source`, `--history-limit`, or `--no-history` must produce a clear error.
- FR-010: Deduplication must apply across all positional arguments using canonical video IDs.
- FR-011: Processing order must match the order in which arguments appear on the command line.
- FR-012: All existing `--out`, `--languages`, `--timestamps`, `--preserve-formatting`, `--retries`, `--retry-delay-seconds`, and `--strict` flags must apply when positional arguments are used.
- FR-013: README and CLI help must document the positional argument syntax.

## Acceptance Criteria

- AC-001: `yt-transcript-md https://youtu.be/dQw4w9WgXcQ` fetches and saves the transcript for that video without clipboard interaction.
- AC-002: `yt-transcript-md https://youtu.be/abc https://youtu.be/def` fetches both videos in argument order.
- AC-003: `yt-transcript-md export https://youtu.be/abc --out notes.md` writes only that video to `notes.md`.
- AC-004: `yt-transcript-md https://youtu.be/abc --links "https://youtu.be/def"` exits with a non-zero code and a clear error message.
- AC-005: `yt-transcript-md https://youtu.be/abc --input-file links.txt` exits with a non-zero code and a clear error message.
- AC-006: `yt-transcript-md https://youtu.be/abc --history-source copyq` exits with a non-zero code and a clear error message.
- AC-007: `yt-transcript-md https://youtu.be/abc https://youtu.be/abc` processes the video once (deduplication).
- AC-008: An unrecognized argument that is not a YouTube link or video ID exits with a clear error.
- AC-009: `yt-transcript-md` with no arguments still runs the clipboard workflow unchanged.
- AC-010: `go test ./...` passes with coverage for positional argument parsing, validation, deduplication, conflict detection, and routing.

## Success Criteria

- A user can paste space-separated YouTube links directly after the command name and get transcripts immediately.
- No existing workflow is broken: clipboard, history TUI, `--links`, and `--input-file` all behave as before.
- The command produces a clear error for every unsupported argument combination.

## Edge Cases

- Single positional argument that is a bare 11-character video ID without a URL.
- Positional argument that is a YouTube URL with extra query parameters (e.g. `?si=`, `&t=`).
- Positional argument that is a YouTube Shorts URL (`/shorts/`).
- Mix of valid and invalid positional arguments.
- Very large number of positional arguments (e.g. 50+).
- Positional argument that duplicates a link already present in `--links` (conflict error takes priority).
- `--` separator followed by positional arguments.

## Implementation Notes

- In `export.go`, the `rootCmd.RunE` override and `exportCmd.RunE` should inspect `args` after flag parsing. If `len(args) > 0`, reject conflicting flags, join args as comma-separated input, set `opts.Links`, and call `runExport()`.
- No new flags, structs, or packages are required.
- The existing `input.ExtractClipboardVideoInputs` / `input.ParseLinks` parser handles URL normalization and video ID extraction; positional arguments go through the same path.
- Cobra's `Args` validator should not be set to `cobra.NoArgs` to allow positional arguments through.

## Human Approval Status

Pending.
