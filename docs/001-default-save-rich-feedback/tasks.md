# Tasks: Clipboard-First Rebrand, Default Save, Rich Metadata, and Colored Feedback

## Track Summary

Implement the approved plan in small, verifiable slices with no new dependencies and no package/binary rename. Existing file-output commands must keep working while the no-argument workflow becomes save-and-copy by default.

## Prerequisites

- Spec approved: yes
- Architecture review approved: yes
- Dependencies approved: no new dependencies planned
- High-risk areas: none
- Verification commands: `go test ./...`, `git diff --check`

## Task List

### T001: Add Metadata Model and oEmbed Provider

- Objective: Add optional video metadata support and a production oEmbed metadata provider.
- Source artifacts: `spec.md`, `plan.md`, `data-model.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/models/models.go`
  - `internal/metadata/`
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
  - production config, deployment files
- Acceptance criteria:
  - `models.VideoMetadata` exists and is attachable to `models.TranscriptDocument`.
  - Metadata provider fetches YouTube oEmbed JSON for a parsed video input.
  - Provider decodes title, author, provider, thumbnail, dimensions, and cache age.
  - Provider ignores oEmbed HTML.
  - 4xx/malformed responses are non-fatal provider errors.
  - Network/5xx retries are bounded by configured retry count.
- Verification command or observable result: `go test ./internal/metadata ./internal/models`
- Dependencies: none
- Risk level: low
- Approval needed: none
- Status: completed

### T002: Render Enriched Markdown Details

- Objective: Render available metadata in each successful video section without breaking existing transcript output.
- Source artifacts: `spec.md`, `plan.md`, `data-model.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/markdown/render.go`
  - `internal/markdown/render_test.go`
  - `internal/markdown/testdata/`
  - `internal/models/models.go`
- Forbidden files or directories:
  - CLI files
  - provider files
- Acceptance criteria:
  - Successful videos include a stable `Video Details` block.
  - Metadata fields render only when present.
  - Existing source, language, generated-caption flag, snippet count, transcript, timestamp, and failure output remain available.
  - Golden or direct tests cover metadata and no-metadata rendering.
- Verification command or observable result: `go test ./internal/markdown`
- Dependencies: T001
- Risk level: low
- Approval needed: none
- Status: completed

### T003: Add Structured Colored Feedback

- Objective: Replace ad hoc app logging with a small reporter that supports readable colored and plain output.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/app/`
  - optional `internal/ui/`
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
- Acceptance criteria:
  - Reporter emits meaningful progress, success, warning, failure, saved, clipboard, and summary messages.
  - Color can be forced on/off in tests.
  - Color is disabled for `NO_COLOR`, `TERM=dumb`, or non-character stdout.
  - Existing tests can pass with `io.Discard`.
- Verification command or observable result: `go test ./internal/app`
- Dependencies: none
- Risk level: low
- Approval needed: none
- Status: completed

### T004: Integrate Metadata and Default Save-and-Copy Workflow

- Objective: Update app and CLI orchestration so default no-argument mode saves `transcripts.md` and copies the same Markdown to clipboard, while file-output modes remain compatible.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/app/export.go`
  - `internal/app/export_test.go`
  - `internal/cli/export.go`
  - `internal/cli/provider_prod.go`
  - `internal/cli/provider_test_mode.go`
  - `tests/e2e/e2e_test.go`
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
  - production config, deployment files
- Acceptance criteria:
  - Default mode reads clipboard, writes `transcripts.md`, then writes clipboard.
  - File write failure prevents clipboard write.
  - Clipboard write failure after file save returns an error while leaving file intact.
  - Metadata failures warn but do not fail transcript export.
  - Duplicate inputs trigger one transcript fetch and one metadata fetch per unique video.
  - Existing root/export file-output workflows remain compatible.
- Verification command or observable result: `go test ./internal/app ./tests/e2e`
- Dependencies: T001, T002, T003
- Risk level: medium
- Approval needed: none
- Status: completed

### T005: Rebrand README and Cobra Help

- Objective: Reposition README and CLI help around the clipboard-first workflow while keeping advanced file/batch export documented.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `README.md`
  - `internal/cli/root.go`
  - `internal/cli/export.go`
  - `tests/e2e/e2e_test.go`
- Forbidden files or directories:
  - module/package rename files
  - Homebrew formula files
- Acceptance criteria:
  - README opens with copy-link/run-command/save-and-copy workflow.
  - README documents metadata and rate-limit-conscious behavior.
  - Root help leads with clipboard-first usage and includes examples.
  - Export help clearly presents explicit file/batch export.
  - E2E tests assert help text expectations.
- Verification command or observable result: `go test ./tests/e2e`
- Dependencies: T004
- Risk level: low
- Approval needed: none
- Status: completed

### T006: Final Verification and SDD Drift Check

- Objective: Verify all changes and update SDD artifacts if implementation diverged from plan.
- Source artifacts: all SDD artifacts
- Allowed files or directories:
  - `specs/001-default-save-rich-feedback/`
  - any touched source/test/doc files
- Forbidden files or directories:
  - unrelated files
- Acceptance criteria:
  - `go test ./...` passes.
  - `git diff --check` passes.
  - Any implementation deviation is recorded in completion notes or SDD artifacts.
- Verification command or observable result: `go test ./...` and `git diff --check`
- Dependencies: T001-T005
- Risk level: low
- Approval needed: none
- Status: completed

## Dependency Order

1. T001 and T003 can start independently.
2. T002 depends on T001.
3. T004 depends on T001, T002, and T003.
4. T005 depends on T004.
5. T006 depends on all implementation tasks.

## Parallel-Safe Groups

- T001 and T003 are parallel-safe if implemented by separate workers because their write sets do not overlap.
- T005 documentation/help work is not parallel-safe with T004 because both touch CLI/e2e files.

## Verification Matrix

- Metadata provider: `go test ./internal/metadata`
- Markdown renderer: `go test ./internal/markdown`
- App workflow: `go test ./internal/app`
- CLI/e2e behavior: `go test ./tests/e2e`
- Full suite: `go test ./...`
- Whitespace: `git diff --check`

## Blocked or Approval-Required Work

- Product/binary/Homebrew rename is deferred and requires a separate chosen name plus release/migration plan.
- New dependencies remain out of scope unless separately approved.

## Completion Notes

- Implemented within the approved plan without adding dependencies.
- Product/binary/Homebrew rename remains deferred.
- Verification completed with `go test ./...` and `git diff --check`.
