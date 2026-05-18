# Technical Plan: Clipboard-First Rebrand, Default Save, Rich Metadata, and Colored Feedback

## Summary

Implement the approved spec as a compatible CLI evolution. The no-argument command will become the primary workflow: read YouTube input from the clipboard, fetch metadata and transcript data once per unique video, save Markdown to `transcripts.md`, and copy the same Markdown back to the clipboard. README and Cobra help will be rewritten around that workflow. Existing file-output workflows remain compatible.

This plan does not rename the binary, Go module, repository, or Homebrew formula in this implementation slice. A package rename is allowed by the spec, but it is not selected here because no replacement name was provided and the compatibility/release cost is high relative to the requested workflow improvement.

## Inputs Reviewed

- `specs/001-default-save-rich-feedback/spec.md`
- `specs/001-default-save-rich-feedback/checklists/requirements.md`
- `README.md`
- `GEMINI.md`
- `Makefile`
- `go.mod`
- `internal/app/export.go`
- `internal/cli/root.go`
- `internal/cli/export.go`
- `internal/cli/provider_prod.go`
- `internal/cli/provider_test_mode.go`
- `internal/clipboard/clipboard.go`
- `internal/input/parser.go`
- `internal/markdown/render.go`
- `internal/models/models.go`
- `internal/transcript/provider.go`
- `internal/transcript/youtube_provider.go`
- `internal/app/export_test.go`
- `internal/markdown/render_test.go`
- `tests/e2e/e2e_test.go`
- oEmbed specification and provider registry: `https://oembed.com/`, `https://oembed.com/providers.json`
- YouTube developer documentation: `https://developers.google.com/youtube/documentation`

## Assumptions

- The default output path remains `transcripts.md`.
- The default no-argument workflow should write the file first, then copy the same Markdown to the clipboard.
- Public oEmbed metadata is enough for the first enrichment slice: title, author/channel, author URL, provider, provider URL, thumbnail URL, thumbnail dimensions, and cache age when returned.
- Missing metadata should produce a warning and still render the transcript.
- No YouTube Data API key, OAuth, cookie, or authenticated request is required.
- ANSI color can be implemented locally without a new dependency.
- Package/binary/Homebrew renaming is deferred unless a future release explicitly chooses a new name and migration path.

## Architecture Approach

### 1. Preserve the Existing Orchestration Boundary

Keep `internal/app` as the workflow orchestrator. Extend it to coordinate three concerns:

- transcript fetching through the existing `transcript.Provider`
- optional metadata fetching through a new metadata provider interface
- output writing to file and clipboard

The app layer should stay testable with fakes; production wiring belongs in `internal/cli/provider_prod.go`, matching the existing test build-tag pattern.

### 2. Add a Metadata Provider Boundary

Add a small metadata package, tentatively `internal/metadata`, with:

- `Provider` interface: fetch metadata for one `models.VideoInput`
- `OEmbedProvider` production implementation using `https://www.youtube.com/oembed?format=json&url=...`
- bounded retries for network/5xx failures only, capped by `ExportOptions.Retries`
- no retry for 4xx responses or malformed permanent responses
- short HTTP timeout to avoid blocking transcript export

The app workflow should call the metadata provider at most once per unique parsed video. `input.ParseVideoInputs` already deduplicates by video ID before app fetches occur; tests should lock that behavior.

### 3. Extend the Domain Model and Markdown Renderer

Add `models.VideoMetadata` and attach it to `models.TranscriptDocument`.

The Markdown renderer should include a stable “Video Details” block before transcript text. It should render fields only when populated, plus existing language/snippet/source details. Metadata failures should not appear inside successful video details unless useful; warnings belong in CLI output, while failed transcript exports remain in `Failed Videos`.

### 4. Make Default Clipboard Mode Save and Copy

Change the no-argument app path so it:

1. reads clipboard input
2. parses and deduplicates videos
3. fetches metadata and transcripts sequentially
4. renders Markdown once
5. writes `opts.Out` (`transcripts.md` by default)
6. writes the same Markdown to clipboard
7. reports both output actions

If file writing fails, stop before clipboard write. If clipboard writing fails after file writing succeeds, return an error while leaving the saved file intact and reporting that the file was saved.

### 5. Add a Small Colored Status Reporter

Introduce a minimal app-level reporting abstraction, likely in `internal/app` or `internal/ui`, to avoid scattering raw `fmt.Fprintf` calls.

Reporter responsibilities:

- start: parsed N unique videos
- per-video metadata fetch start/warning
- per-video transcript fetch start/success/failure
- output saved path
- clipboard copied
- final summary

Color policy:

- green for success
- yellow for warnings/partial failures
- red for failures
- cyan or blue for progress/info
- plain text remains explicit without color
- disable color when `NO_COLOR` is set, `TERM=dumb`, or stdout is not a character device
- tests can force color on/off through constructor options rather than environment mutation

No new dependency is planned for terminal color.

### 6. Rebrand README and Help Text Around Clipboard-First Use

Rewrite README structure:

- opening value proposition: copy YouTube link, run one command, get Markdown saved and copied
- quick start with no-argument command first
- what the default command does
- generated output details
- advanced file/batch export
- metadata and rate-limit-conscious behavior
- installation and development sections retained

Update Cobra help:

- root `Short`, `Long`, and `Example` should lead with clipboard-first use
- root help should still expose flags for compatibility
- `export` command help should describe advanced explicit file/batch export
- e2e tests should assert the key help text order/phrasing

## Affected Modules

- `internal/models/models.go`: add metadata struct and document field.
- `internal/metadata/`: new oEmbed metadata provider and tests.
- `internal/app/export.go`: orchestrate metadata, default save-and-copy, and structured reporting.
- `internal/markdown/render.go`: render enriched video details.
- `internal/cli/root.go`: clipboard-first help text and examples.
- `internal/cli/export.go`: export help text and possibly provider/reporting wiring.
- `internal/cli/provider_prod.go`: wire production metadata provider and reporter/color decisions.
- `internal/cli/provider_test_mode.go`: wire test metadata/clipboard fakes for e2e tests.
- `internal/clipboard/clipboard.go`: no planned behavior change except continued default wiring.
- `README.md`: full repositioning around the clipboard workflow.
- `internal/app/export_test.go`, `internal/markdown/render_test.go`, `tests/e2e/e2e_test.go`: expanded coverage.

## API and Contract Changes

- CLI behavior change: `yt-transcript-md` with no args will now save `transcripts.md` as well as copy Markdown to clipboard.
- CLI compatibility retained: `export --links`, `export --input-file`, root `--links`, root `--input-file`, and `--out` continue to work.
- Markdown contract change: successful video sections gain a richer `Video Details` area when metadata is available.
- Help text contract change: root help leads with clipboard-first usage; export help remains for explicit/batch file output.
- No external HTTP API is exposed by this project.

## Data Model Changes

See `data-model.md`.

## Dependency Changes

No dependency changes are planned.

Alternatives considered:

- `github.com/atotto/clipboard`: useful but not needed; previous clipboard implementation already avoids adding it.
- `fatih/color` or another color package: not needed for simple ANSI output and would require dependency approval.
- YouTube Data API: rejected for this slice because it needs API keys/quotas and is unnecessary for title/channel/thumbnail metadata.

Approval status: no dependency approval required.

## Security Impact

- Clipboard data is treated as local input. Do not log full raw clipboard contents.
- URLs rendered into Markdown should be plain text/Markdown links generated from trusted parsed values or escaped as text. No HTML from oEmbed should be rendered.
- oEmbed `html` should be ignored to avoid embedding untrusted HTML in Markdown.
- Metadata provider should only request YouTube oEmbed for parsed YouTube video URLs derived from valid video IDs.
- No credentials, cookies, OAuth tokens, or secrets are introduced.

## Authorization Boundaries

- No authentication or authorization is added.
- The tool only accesses public transcript/metadata endpoints for user-provided public video IDs.
- Private or unavailable videos should fail gracefully as metadata/transcript failures.

## Observability Impact

- CLI output becomes the observability surface for local runs.
- The final summary should report unique videos, successful transcripts, failed transcripts, metadata warnings, saved file path, and clipboard copy status.
- No telemetry, remote logging, or analytics are added.

## Testing Strategy

See `test-plan.md`.

Verification commands:

- `go test ./...`
- `git diff --check`
- optionally `make verify` if local lint/build tools are available and not blocked by environment

## Failure Modes

- Clipboard read fails: command exits before network requests.
- Clipboard input invalid: command exits before network requests.
- Metadata fetch fails: warning only; transcript export continues.
- Transcript fetch fails: record failed video; continue unless strict.
- File write fails in default mode: command returns error and does not copy to clipboard.
- Clipboard write fails after file save: command returns error and leaves saved Markdown intact.
- Terminal color unsupported: output remains readable plain text.
- oEmbed returns malformed JSON: warning only; transcript export continues.

## Rollback and Recovery

- Code rollback is a normal Git revert of the implementation commit(s); no migration or external state exists.
- Runtime recovery from default file overwrite remains the same as current output behavior: users can choose `--out` for a different path.
- If metadata endpoint behavior changes, disable or bypass metadata provider without affecting transcript provider and core rendering.
- If colored output causes issues, color can be disabled by `NO_COLOR`, non-TTY detection, or a small follow-up flag if needed.

## Risks and Mitigations

- Risk: help/README rebrand overpromises metadata details that are not always available.
  Mitigation: document metadata as “when available” and render only populated fields.

- Risk: oEmbed adds extra request volume.
  Mitigation: one sequential metadata request per unique video, bounded retries, no search/discovery/polling.

- Risk: ANSI output pollutes redirected logs.
  Mitigation: disable color for non-character devices and honor `NO_COLOR`.

- Risk: changing no-arg behavior to write a file surprises users who expected clipboard-only.
  Mitigation: README/help must state save-and-copy clearly; existing explicit workflows remain unchanged.

- Risk: package rename could break installation and scripts.
  Mitigation: defer renaming to a separate release plan with a chosen name, aliases, and Homebrew migration steps.

## Human Architecture Approval Status

Approved by user on 2026-05-18. Task breakdown and implementation may proceed within this plan.
