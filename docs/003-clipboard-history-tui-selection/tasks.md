# Tasks: Clipboard History TUI Selection

## Track Summary

Implement clipboard-history provider scanning and a TUI selector for macOS/Linux default clipboard mode. Existing explicit export workflows remain non-interactive.

## Prerequisites

- Spec approved: yes
- Dependency approval: yes
- Architecture review approved: yes
- High-risk areas: dependencies, clipboard-history privacy, interactive TUI
- Verification commands: `go test ./...`, `make test`, `make lint`, `make vet`, `make build`, `git diff --check`, `make verify`

## Task List

### T001: Add History Provider Interfaces and Fixtures

- Objective: Define provider abstractions and test fixtures without production command execution yet.
- Source artifacts: `spec.md`, `plan.md`, `data-model.md`, `test-plan.md`
- Allowed files or directories:
  - `internal/history/`
- Forbidden files or directories:
  - `go.mod`
  - `go.sum`
- Acceptance criteria:
  - Provider, Entry, Options, and Candidate models exist.
  - Tests cover aggregation with fake providers.
  - No provider command execution is added yet.
- Verification command or observable result: `go test ./internal/history`
- Dependencies: none
- Risk level: low
- Approval needed: spec approval
- Status: completed

### T002: Implement CopyQ Provider

- Objective: Read recent text entries from CopyQ through its CLI.
- Allowed files or directories:
  - `internal/history/`
- Forbidden files or directories:
  - TUI files
- Acceptance criteria:
  - Detects `copyq`.
  - Reads up to configured limit.
  - Handles missing/not-running CopyQ with clear provider error.
  - Tests use fake command runner or fixtures.
- Verification command or observable result: `go test ./internal/history`
- Dependencies: T001
- Risk level: medium
- Approval needed: architecture approval
- Status: completed

### T003: Implement cliphist Provider

- Objective: Read recent Wayland clipboard history from cliphist.
- Allowed files or directories:
  - `internal/history/`
- Forbidden files or directories:
  - TUI files
- Acceptance criteria:
  - Detects `cliphist`.
  - Uses list/decode flow or documented equivalent.
  - Handles missing/empty/malformed output.
  - Tests use fake command runner or fixtures.
- Verification command or observable result: `go test ./internal/history`
- Dependencies: T001
- Risk level: medium
- Approval needed: architecture approval
- Status: completed

### T004: Implement GPaste Provider

- Objective: Read recent GNOME clipboard history from `gpaste-client`.
- Allowed files or directories:
  - `internal/history/`
- Forbidden files or directories:
  - TUI files
- Acceptance criteria:
  - Detects `gpaste-client`.
  - Reads indexed/raw text entries up to configured limit.
  - Handles missing daemon/empty/malformed output.
  - Tests use fake command runner or fixtures.
- Verification command or observable result: `go test ./internal/history`
- Dependencies: T001
- Risk level: medium
- Approval needed: architecture approval
- Status: completed

### T005: Add Bubble Tea TUI Selector

- Objective: Add searchable multi-select terminal UI for video candidates.
- Allowed files or directories:
  - `internal/cli/`
  - `go.mod`
  - `go.sum`
- Forbidden files or directories:
  - provider command logic
- Acceptance criteria:
  - TUI supports navigation, search, multi-select, select visible, first N, confirm, and cancel.
  - TUI is behind an interface and can be faked in app/e2e tests.
  - Dependency changes match approved list.
- Verification command or observable result: `go test ./internal/cli`
- Dependencies: dependency approval, T001
- Risk level: medium
- Approval needed: dependency approval
- Status: completed

### T006: Wire Default Workflow and CLI Flags

- Objective: Integrate current clipboard, history providers, TUI selector, and selected videos into default root mode.
- Allowed files or directories:
  - `internal/app/`
  - `internal/cli/`
  - `tests/e2e/e2e_test.go`
- Forbidden files or directories:
  - transcript provider internals
  - metadata provider internals
- Acceptance criteria:
  - Root no-argument mode scans current clipboard and history when interactive.
  - `--history-source`, `--history-limit`, and `--no-history` are implemented.
  - Explicit workflows do not scan history or open TUI.
  - Non-interactive multi-candidate runs require explicit selection.
- Verification command or observable result: `go test ./internal/app ./internal/cli ./tests/e2e`
- Dependencies: T001-T005
- Risk level: high
- Approval needed: architecture approval
- Status: completed

### T007: Update README and Help

- Objective: Document provider setup, history flags, TUI controls, and privacy behavior.
- Allowed files or directories:
  - `README.md`
  - `internal/cli/root.go`
  - `tests/e2e/e2e_test.go`
- Forbidden files or directories:
  - unrelated docs
- Acceptance criteria:
  - README documents CopyQ, cliphist, and GPaste setup expectations.
  - README documents TUI controls.
  - Root help documents history flags.
  - E2E help tests cover key wording.
- Verification command or observable result: `go test ./tests/e2e`
- Dependencies: T006
- Risk level: low
- Approval needed: none after implementation approval
- Status: completed

### T008: Final Verification and Drift Check

- Objective: Verify all changes and update SDD artifacts if implementation diverged.
- Allowed files or directories:
  - any touched implementation, test, or docs files
- Forbidden files or directories:
  - unrelated files
- Acceptance criteria:
  - `go test ./...` passes.
  - `git diff --check` passes.
  - `make verify` passes, or any failure is explained and reduced to an expected repository-state check.
  - SDD artifacts reflect final behavior.
- Verification command or observable result: `go test ./...`, `git diff --check`, `make verify`
- Dependencies: T001-T007
- Risk level: low
- Approval needed: none
- Status: completed with note
- Completion note: `go test ./...`, `make test`, `make lint`, `make vet`, `make build`, and `git diff --check` passed. `make verify` reached `tidy-check` and failed because it asserts `git diff --exit-code go.mod go.sum`; those files intentionally changed to add the approved Bubble Tea/Bubbles dependencies and remain uncommitted.

## Dependency Order

1. T001
2. T002, T003, and T004 can proceed after T001
3. T005 after dependency approval
4. T006 after T001-T005
5. T007 after T006
6. T008

## Parallel-Safe Groups

- T002, T003, and T004 are parallel-safe if they touch only provider-specific files.
- T005 is parallel-safe with provider work after TUI dependency approval.
- T007 is not parallel-safe with T006 because both touch CLI help/e2e files.

## Verification Matrix

- History providers: `go test ./internal/history`
- TUI selector: `go test ./internal/cli`
- App workflow: `go test ./internal/app`
- E2E behavior: `go test ./tests/e2e`
- Full suite: `go test ./...`
- Full project verification: `make verify`
- Whitespace: `git diff --check`

## Blocked or Approval-Required Work

- Spec approval was granted by the user on 2026-06-25.
- Bubble Tea/Bubbles dependency approval was granted by the user on 2026-06-25.
- Architecture approval was granted by the user on 2026-06-25.
