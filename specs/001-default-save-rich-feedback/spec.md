# Feature Spec: Clipboard-First Rebrand, Default Save, Rich Metadata, and Colored Feedback

## Problem Statement

The default no-argument workflow currently focuses on clipboard input/output, but the product messaging still presents the tool primarily as a traditional file-export CLI. Users need the tool to clearly market and explain its simplest workflow: copy a YouTube link, run one command, get a Markdown transcript saved and copied. Users also need a durable Markdown file by default, richer video context in that Markdown, clearer progress feedback, and a rate-limit-conscious approach that avoids unnecessary YouTube requests.

## Goals

- Make `yt-transcript-md` with no arguments save the generated Markdown to disk by default.
- Keep the clipboard workflow useful by continuing to copy generated Markdown back to the clipboard.
- Reposition README and command help around the clipboard-first workflow as the main product promise.
- Add richer video details to each Markdown transcript when those details are available.
- Avoid request patterns that could abuse YouTube or related metadata endpoints.
- Add colored, readable CLI feedback for progress, success, warning, and failure states.

## Non-Goals

- Do not add audio downloading, speech-to-text, or transcript generation.
- Do not scrape private or authenticated YouTube-only data.
- Do not add long-running polling, background workers, or watch mode.
- Do not require a YouTube Data API key for the default workflow.
- Do not make color the only signal for status; text must remain meaningful without color.
- Do not rename the binary, Go module, Homebrew formula, repository, or package identifiers unless the technical plan explicitly accepts the compatibility and release cost.

## Users or Actors

- CLI users who copy a YouTube link, run `yt-transcript-md`, and expect an immediately usable transcript file.
- CLI users running batch exports who need clear feedback and rate-limit-conscious behavior.
- Automation users who depend on existing `export`, `--links`, `--input-file`, and `--out` behavior.

## User Journeys

1. A user copies a YouTube URL, runs `yt-transcript-md`, and gets both `transcripts.md` in the current directory and the same Markdown copied to the clipboard.
2. A new user opens `yt-transcript-md --help` and immediately understands the simplest workflow without needing to read advanced flags first.
3. A new user opens the README and sees the clipboard workflow positioned as the primary use case, with file/batch export presented as advanced or secondary workflows.
4. A user runs `yt-transcript-md export --links ... --out notes.md` and gets a Markdown file that includes transcript text plus video details where available.
5. A user exports multiple videos and sees colored progress messages for parsing input, fetching metadata, fetching transcripts, saving output, copying to clipboard, partial failures, and the final summary.
6. A user retries a failed video without the tool issuing duplicate metadata or transcript requests for duplicate input links in the same run.

## Functional Requirements

- FR-001: With no command-line arguments, the command must read YouTube input from the clipboard.
- FR-002: With no command-line arguments, the command must save Markdown to the existing default output path, `transcripts.md`, in the current working directory.
- FR-003: With no command-line arguments, the command must copy the generated Markdown back to the clipboard after a successful render.
- FR-004: Existing explicit file-output behavior must remain compatible for `export --links`, `export --input-file`, root `--links`, root `--input-file`, and `--out`.
- FR-005: Markdown output must include richer per-video details when available, including title, author/channel name, author/channel URL, thumbnail URL, provider name, and provider URL.
- FR-006: Missing metadata must not fail a transcript export unless transcript fetching itself fails or strict mode requires failure.
- FR-007: The tool must deduplicate input videos before any transcript or metadata fetches.
- FR-008: The tool must perform at most one transcript fetch sequence per unique video per run, excluding bounded retries for failures.
- FR-009: The tool must perform at most one metadata fetch sequence per unique video per run, excluding bounded retries for failures.
- FR-010: Metadata retries must be bounded and must not exceed the transcript retry count configured for the run.
- FR-011: Batch processing must remain sequential by default.
- FR-012: The CLI must show colored status feedback for start, per-video progress, success, warnings, failures, saved file path, clipboard copy, and final summary.
- FR-013: CLI feedback must be useful when colors are disabled or unsupported.
- FR-014: Partial failures must still produce Markdown containing successful videos and failed video details unless strict mode stops the run.
- FR-015: README must be overhauled to position the clipboard-first workflow as the primary purpose of the tool.
- FR-016: README must include a concise value proposition, a one-command quick start, what the default command does, and when to use advanced file/batch flags.
- FR-017: README must describe the default save-and-copy behavior, metadata behavior, and rate-limit-conscious safeguards.
- FR-018: CLI root help must be rewritten so the short and long descriptions market the simple clipboard workflow first.
- FR-019: CLI help must include examples for the no-argument clipboard workflow and advanced file-output workflow.
- FR-020: Export subcommand help must remain available for explicit file/batch use, but must not obscure the default clipboard-first path.

## Acceptance Criteria

- AC-001: Running `yt-transcript-md` with a clipboard YouTube URL creates `transcripts.md` and copies the generated Markdown to the clipboard.
- AC-002: Running `yt-transcript-md --links <url> --out custom.md` writes `custom.md` and does not require clipboard access.
- AC-003: Markdown for a video with available metadata contains a video details section with title, channel/author, thumbnail, provider, language, generated-caption flag, snippet count, and source URL.
- AC-004: If metadata fetching fails but transcript fetching succeeds, the Markdown is still written and the CLI shows a warning.
- AC-005: Duplicate inputs do not cause duplicate metadata or transcript fetches in the same run.
- AC-006: A batch of N unique videos emits progress for each unique video and a final summary with success and failure counts.
- AC-007: `yt-transcript-md --help` presents clipboard-first usage before advanced export options.
- AC-008: `yt-transcript-md export --help` explains explicit file/batch export without contradicting the default clipboard workflow.
- AC-009: README opens with the simple clipboard workflow and treats batch/file export as secondary.
- AC-010: `go test ./...` passes with coverage for default clipboard save, metadata-enriched Markdown, metadata failure tolerance, deduped fetches, colored feedback text, and help text expectations.

## Success Criteria

- Default no-argument use produces a durable Markdown file without extra flags.
- README and CLI help make the one-command clipboard workflow obvious before users inspect advanced flags.
- Users can see what the tool is doing during longer transcript fetches.
- Metadata enrichment adds useful context without making transcript export fragile.
- Request volume scales linearly with unique video count and bounded retries only.

## Edge Cases

- Clipboard is empty or does not contain a YouTube URL/video ID.
- Clipboard input contains duplicate links or mixed comma/newline input.
- `transcripts.md` already exists in the current directory.
- Metadata endpoint returns non-200, invalid JSON, or incomplete fields.
- Transcript exists but metadata does not.
- Metadata exists but transcript does not.
- Terminal does not support color.
- Clipboard write succeeds but file write fails, or file write succeeds but clipboard write fails.
- Existing users expect file-export positioning from previous README/help text.
- Help text becomes too verbose for terminal use.

## Data Sensitivity and Compliance Notes

- Clipboard input may contain user-private data; the tool must only parse the clipboard locally and must not log the full clipboard contents beyond the extracted YouTube inputs.
- The default behavior contacts YouTube-related endpoints only for the unique video IDs being exported.
- No credentials, cookies, OAuth tokens, or YouTube Data API keys are required for the default behavior.

## API or Integration Expectations

- Metadata fetching should prefer public, no-auth metadata that can provide title, author/channel, provider, and thumbnail fields.
- oEmbed is an acceptable metadata source because the oEmbed spec defines title, author, provider, thumbnail, and cache fields, and the public oEmbed provider registry lists YouTube support at `https://www.youtube.com/oembed`.
- Metadata support must be optional and failure-tolerant.
- The tool must not add automatic polling, bulk discovery, search API calls, or unbounded parallel requests.

## Assumptions

- The default save path should be the existing default `transcripts.md`.
- Existing overwrite behavior for output files remains acceptable for default saves.
- The no-argument workflow should both save the file and copy Markdown to the clipboard.
- “Complete rebrand” may include a product/package rename, but a rename is optional rather than required.
- “More video details” means public metadata available without authentication: title, author/channel, author/channel URL, thumbnail URL, provider name, and provider URL.
- Colored feedback can use ANSI terminal colors directly or an existing dependency if already present; adding a new dependency requires separate approval.

## Open Questions

- None.

## Human Approval Status

Approved by user on 2026-05-18 for technical planning. Architecture review and task breakdown are still required before implementation.
