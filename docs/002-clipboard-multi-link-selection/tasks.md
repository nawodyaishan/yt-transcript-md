# Tasks: Clipboard Multi-Link Selection Prompt

## Track Summary

Implement the approved plan in small slices. The default clipboard workflow gains a multi-link selection prompt, while explicit batch export remains non-interactive.

## Prerequisites

- Spec approved: yes
- Architecture review approved: yes
- Dependencies approved: no new dependencies planned
- High-risk areas: interactive CLI behavior, automation compatibility
- Verification commands: `go test ./...`, `git diff --check`

## Task List

### T001: Add Selection Model and Pure Validation

- Objective: Define clipboard selection modes and validate/apply them to parsed videos.
- Source artifacts: `spec.md`, `plan.md`, `data-model.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/app/`
  - `internal/models/models.go` only if model placement is approved there
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
- Acceptance criteria:
  - Selection modes cover all, one, recent, and cancel.
  - Applying a selection returns the correct subset.
  - Invalid indexes and counts return clear errors.
  - Unit tests cover selection behavior.
- Verification command or observable result: `go test ./internal/app`
- Dependencies: none
- Risk level: low
- Approval needed: spec approval
- Status: completed

### T002: Add Clipboard Extraction for Prose Input

- Objective: Extract valid YouTube videos from arbitrary clipboard text before canonical parsing and selection.
- Source artifacts: `spec.md`, `plan.md`, `data-model.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/input/`
  - `internal/app/export_test.go` only for workflow coverage if needed
- Forbidden files or directories:
  - explicit export CLI behavior
  - provider implementations
- Acceptance criteria:
  - Clipboard extraction preserves first-seen order.
  - Surrounding prose and unsupported URLs do not make clipboard mode fail when valid YouTube videos exist.
  - No-valid-video clipboard input returns the existing no-valid-input error.
  - Explicit `--links`, `--input-file`, and `export` parsing behavior remains unchanged.
- Verification command or observable result: `go test ./internal/input ./internal/app`
- Dependencies: none
- Risk level: medium
- Approval needed: spec approval
- Status: completed

### T003: Refactor App Export to Accept Preselected Videos

- Objective: Let clipboard mode select videos after parsing and before existing fetch/render work.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/app/export.go`
  - `internal/app/export_test.go`
- Forbidden files or directories:
  - provider implementations
  - markdown renderer
- Acceptance criteria:
  - Existing explicit export tests still pass.
  - Clipboard mode can render a supplied selected subset.
  - Metadata/transcript providers are not called for unselected videos.
  - Reporter start count reflects selected videos.
- Verification command or observable result: `go test ./internal/app`
- Dependencies: T001, T002
- Risk level: medium
- Approval needed: spec approval
- Status: completed

### T004: Add Terminal Selector and Non-Interactive Selection Flag

- Objective: Wire production root clipboard mode to an interactive prompt and a non-interactive selection escape hatch.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/cli/export.go`
  - `internal/cli/root.go`
  - optional new `internal/cli/selection.go`
  - `internal/cli/provider_test_mode.go`
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
- Acceptance criteria:
  - Multi-video clipboard input prompts only in interactive root no-argument mode.
  - Non-interactive multi-video clipboard input without explicit selection fails before network requests.
  - Selection flag supports all, one index, and recent count.
  - Explicit `--links`, `--input-file`, and `export` do not prompt and reject the selection flag.
- Verification command or observable result: `go test ./internal/cli ./tests/e2e`
- Dependencies: T001-T003
- Risk level: medium
- Approval needed: architecture approval
- Status: completed

### T005: Expand Workflow Tests and E2E Coverage

- Objective: Lock the new selection behavior and compatibility guarantees.
- Source artifacts: `spec.md`, `plan.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/app/export_test.go`
  - `tests/e2e/e2e_test.go`
  - test fixtures as needed
- Forbidden files or directories:
  - production behavior outside accepted fixups
- Acceptance criteria:
  - Unit tests cover all selection modes and cancellation side effects.
  - Unit tests cover prose extraction and explicit-input compatibility.
  - E2E tests cover one, all, recent, and non-interactive error behavior.
  - Existing explicit batch e2e tests continue passing.
- Verification command or observable result: `go test ./...`
- Dependencies: T001-T004
- Risk level: medium
- Approval needed: none after implementation approval
- Status: completed

### T006: Update README and Help Text

- Objective: Document the multi-link clipboard prompt and automation-safe alternatives.
- Source artifacts: `spec.md`, `plan.md`
- Allowed files or directories:
  - `README.md`
  - `internal/cli/root.go`
  - `internal/cli/export.go` only if wording needs clarification
- Forbidden files or directories:
  - unrelated docs
- Acceptance criteria:
  - README explains one-link no-prompt behavior.
  - README explains multi-link prompt choices.
  - Root help mentions selection behavior and any selection flag.
  - Export help remains focused on explicit batch workflows.
- Verification command or observable result: `go test ./tests/e2e`
- Dependencies: T004
- Risk level: low
- Approval needed: none after implementation approval
- Status: completed

### T007: Final Verification and SDD Drift Check

- Objective: Verify implementation and update artifacts if behavior diverges from the approved spec.
- Source artifacts: all `docs/002-clipboard-multi-link-selection/` artifacts
- Allowed files or directories:
  - any touched implementation, test, or docs files
- Forbidden files or directories:
  - unrelated files
- Acceptance criteria:
  - `go test ./...` passes.
  - `git diff --check` passes.
  - SDD artifacts reflect final behavior.
- Verification command or observable result: `go test ./...` and `git diff --check`
- Dependencies: T001-T006
- Risk level: low
- Approval needed: none
- Status: completed

## Dependency Order

1. T001
2. T002
3. T003
4. T004
5. T005 and T006 can proceed after T004
6. T007

## Parallel-Safe Groups

- T005 test expansion can begin with fake app-level selectors after T001/T003.
- T006 docs/help is not parallel-safe with T004 because both may touch CLI help files.

## Verification Matrix

- Selection logic: `go test ./internal/app`
- Clipboard extraction: `go test ./internal/input`
- CLI prompt and flags: `go test ./internal/cli`
- E2E behavior: `go test ./tests/e2e`
- Full suite: `go test ./...`
- Whitespace: `git diff --check`

## Blocked or Approval-Required Work

- Implementation proceeded after human approval of the spec and architecture.
- Any new prompt dependency requires separate approval.
- Publish-date-based recency is out of scope unless the spec is revised.

## Completion Notes

- Implemented without adding dependencies.
- Clipboard mode now extracts valid YouTube videos from surrounding text.
- Default clipboard mode prompts for multi-video input only when interactive; non-interactive runs use `--clipboard-selection`.
- Verification completed with `go test ./...` and `git diff --check`.
