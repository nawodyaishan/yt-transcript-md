# Go CLI Migration Plan

## Objective

Migrate `yt-transcript-md` from a Python package into a higher-quality Go CLI while preserving current user-facing behavior and adding a stronger QA, release, and Homebrew distribution pipeline.

The migrated tool should remain focused on one job: read YouTube links or video IDs, fetch available captions/transcripts, and render a Markdown transcript file.

## Current Baseline

The existing Python implementation is compact and behaviorally clear:

- CLI: `src/yt_transcript_md/cli.py`
- Input parsing: `src/yt_transcript_md/parser.py`
- Transcript fetch: `src/yt_transcript_md/transcript.py`
- Markdown rendering: `src/yt_transcript_md/markdown.py`
- Domain models: `src/yt_transcript_md/models.py`
- Unit tests: `tests/test_parser.py`, `tests/test_markdown.py`
- Quality commands: `make check`

Current CLI surface:

```bash
yt-transcript-md --links "https://youtu.be/dQw4w9WgXcQ" --out transcripts.md
yt-transcript-md --input-file links.txt --out transcripts.md
yt-transcript-md --links "dQw4w9WgXcQ" --languages "en,si,hi" --timestamps
```

Behavior to preserve:

- Accept comma-separated and newline-separated input.
- Accept raw 11-character YouTube video IDs.
- Accept `youtu.be`, `/watch?v=`, `/shorts/`, `/embed/`, and `/live/` URLs.
- Deduplicate video IDs while preserving input order.
- Support language priority via `--languages`.
- Support `--timestamps`.
- Support `--preserve-formatting`.
- Retry transcript fetches with configurable retry count and delay.
- Continue on per-video failures by default.
- Exit non-zero in strict mode when any video fails.
- Create parent directories for the output Markdown file.

## Reference Inputs

Research and local references used for this plan:

- Local reference repo: `/Users/nawodyaishan/Documents/GitHub/mcp-config-tui`
- Cobra docs via Context7: command constructors, `RunE`, flag validation, output wiring, and testability.
- GoReleaser docs via Context7: reproducible builds, cross-platform artifacts, checksums, Homebrew `brews` publishing, and token setup.
- Exa research: Go CLI structure, Go testing practices, YouTube transcript Go library options, golden tests, fuzz tests, and race detection.
- Go official docs surfaced via Exa: table-driven tests, subtests, fuzzing, and race detector.

## Target Product Decisions

### CLI framework

Use `spf13/cobra` for the production CLI.

Rationale:

- It gives durable help output, flag validation, shell completions, and command constructors.
- It supports `RunE`, which keeps command execution testable and keeps exit code handling centralized.
- It is the standard shape expected by Go CLI users and release tooling.

Initial command shape:

```bash
yt-transcript-md export --links "..." --out transcripts.md
yt-transcript-md export --input-file links.txt --out transcripts.md
yt-transcript-md version
yt-transcript-md completion bash|zsh|fish|powershell
```

Compatibility alias:

```bash
yt-transcript-md --links "..." --out transcripts.md
```

Keep the root command able to run export behavior for backward compatibility with the current Python CLI. Internally this can route to the same export options as the `export` subcommand.

### Transcript provider strategy

Do not blindly depend on the first Go YouTube transcript library.

Current Go library candidates found during research are young and vary in maturity:

- `github.com/horiagug/youtube-transcript-api-go`
- `github.com/rahadiangg/youtube-transcript-go`
- `github.com/paulstuart/yt-transcript`

Recommended approach:

- Define an internal `transcript.Provider` interface first.
- Implement the provider behind that interface.
- Start by evaluating `horiagug/youtube-transcript-api-go` and `rahadiangg/youtube-transcript-go` in a spike.
- Keep a fallback path open for a small internal provider based on YouTube transcript endpoint behavior if third-party libraries prove unstable.
- Record provider decision in `docs/transcript_provider_decision.md` before removing Python.

Provider interface target:

```go
type Provider interface {
	Fetch(ctx context.Context, video VideoInput, opts FetchOptions) (TranscriptDocument, error)
}
```

### Markdown output compatibility

The first Go release should generate Markdown that is intentionally close to the current Python output. A golden compatibility suite should pin the format before the Python code is removed.

Allowed output changes:

- Timestamp generation will naturally differ.
- Error strings may differ if they are clearer and tested.

Not allowed without a documented breaking change:

- Reordering videos.
- Dropping failed video reporting.
- Changing timestamp format from `MM:SS` and `HH:MM:SS`.
- Changing default output file from `transcripts.md`.
- Changing default language from `en`.

## Proposed Go Project Layout

Use a layout similar in spirit to `mcp-config-tui`, but keep this repo simpler because it is a non-TUI CLI.

```text
yt-transcript-md/
  cmd/
    yt-transcript-md/
      main.go
      main_test.go
  internal/
    cli/
      root.go
      export.go
      version.go
      completion.go
      export_test.go
    input/
      parser.go
      parser_test.go
      fuzz_test.go
    markdown/
      render.go
      render_test.go
      testdata/
        plain.golden
        timestamps.golden
        failures.golden
    transcript/
      provider.go
      youtube_provider.go
      provider_fake_test.go
      retry.go
      retry_test.go
    app/
      export.go
      export_test.go
    version/
      version.go
      version_test.go
  tests/
    e2e/
      e2e_test.go
      testdata/
        export_plain.golden
        export_timestamps.golden
        export_failures.golden
  scripts/
    lib/common.sh
    build.sh
    test.sh
    lint.sh
    vet.sh
    tidy.sh
    tidy-check.sh
    mod-verify.sh
    verify.sh
    release.sh
    tag.sh
    clean.sh
  docs/
    golang_migration_plan.md
    transcript_provider_decision.md
  go.mod
  go.sum
  Makefile
  .golangci.yml
  .goreleaser.yml
  .github/workflows/ci.yml
  .github/workflows/release.yml
```

Package responsibilities:

- `cmd/yt-transcript-md`: tiny process entrypoint only.
- `internal/cli`: Cobra commands, flag parsing, user-facing errors, and command output.
- `internal/app`: orchestration of parse, fetch, render, write.
- `internal/input`: YouTube input splitting, normalization, validation, and dedupe.
- `internal/transcript`: transcript provider abstraction, retry behavior, and concrete YouTube provider.
- `internal/markdown`: deterministic Markdown rendering.
- `internal/version`: linker-injected version metadata.
- `tests/e2e`: compiled-binary integration tests using `os/exec`.

## Migration Phases

### Phase 0: Baseline Lock

Goal: preserve current behavior before writing the Go implementation.

Tasks:

- Add Python golden fixtures for current Markdown output, or manually create equivalent expected fixtures from current tests.
- Document current CLI flags and exit code expectations.
- Add examples for invalid URL, missing input file, duplicate IDs, strict failure, and output directory creation.
- Decide whether root-command export compatibility is required for the first Go release. Recommended: yes.

Exit criteria:

- Current behavior is captured in test data.
- A maintainer can compare Go output against expected Python behavior without network access.

### Phase 1: Go Skeleton

Goal: introduce Go module, CLI shell, version command, and local quality tooling.

Tasks:

- Create `go.mod` using a stable Go version from local policy.
- Add `cmd/yt-transcript-md/main.go`.
- Add `internal/version` with linker-injected fields: `Version`, `Commit`, `Date`, `GoVersion`.
- Add Cobra root command with `export`, `version`, and `completion`.
- Add Makefile targets modeled after `mcp-config-tui`: `tidy`, `tidy-check`, `mod-verify`, `fmt`, `vet`, `lint`, `test`, `coverage-check`, `build`, `verify`, `clean`.
- Add scripts under `scripts/` with `set -euo pipefail` and shared `scripts/lib/common.sh`.

Exit criteria:

- `go test ./...` passes.
- `make verify` runs module verification, tidy check, vet, lint, tests, build, and gitignore guard if adopted.
- `yt-transcript-md --help`, `yt-transcript-md export --help`, and `yt-transcript-md version` work.

### Phase 2: Parser and Markdown Port

Goal: port deterministic code first.

Tasks:

- Port video ID extraction and input splitting.
- Add table-driven parser tests matching current Python cases.
- Add fuzz test for `ExtractVideoID` and `ParseVideoInputs`.
- Port Markdown rendering.
- Add golden tests for plain transcripts, timestamped transcripts, failures, empty snippets, and paragraph wrapping.
- Inject clock/time source for deterministic tests.

Exit criteria:

- Parser tests cover all current accepted URL forms.
- Markdown golden tests pass with stable output.
- Fuzz seeds include valid URLs, invalid URLs, raw IDs, angle-bracket-wrapped links, empty strings, and mixed comma/newline input.

### Phase 3: App Orchestration

Goal: wire parsing, provider, rendering, filesystem writes, and exit semantics without real network dependency.

Tasks:

- Implement `internal/app.Export(ctx, options, provider, writer)`.
- Use a fake transcript provider in tests.
- Preserve partial success behavior.
- Preserve strict mode semantics.
- Validate mutually exclusive or missing input conditions.
- Validate retry count and delay values.
- Ensure output parent directory creation is tested.

Exit criteria:

- Unit tests prove success, partial failure, strict failure, missing input, missing file, output directory creation, and language parsing.
- No tests require YouTube network access.

### Phase 4: Transcript Provider Spike

Goal: choose and isolate a transcript fetching implementation.

Tasks:

- Evaluate `horiagug/youtube-transcript-api-go` and `rahadiangg/youtube-transcript-go` against a small manual matrix.
- Validate support for language priority and preserve formatting.
- Check whether the library exposes language name, language code, generated/manual status, start, duration, and text.
- Check failure modes for unavailable transcript, disabled captions, invalid video ID, and network timeout.
- Confirm dependency license compatibility.
- Add context timeout support around network calls.
- Add retry behavior in our code, not buried in CLI code.

Exit criteria:

- Provider choice is documented in `docs/transcript_provider_decision.md`.
- Provider is behind `internal/transcript.Provider`.
- Unit tests use fakes; network tests are opt-in with a build tag such as `integration`.

### Phase 5: Binary E2E Tests

Goal: test the installed CLI as users run it.

Pattern to copy from `mcp-config-tui`:

- Use `TestMain` to build the CLI once into a temp directory.
- Run test scenarios with `os/exec`.
- Compare generated Markdown against golden files.
- Sanitize nondeterministic timestamps or inject a test clock.

E2E cases:

- `--links` with one raw ID and fake provider.
- `--links` with duplicate URLs.
- `--input-file` with newline-separated links.
- `--timestamps`.
- `--languages en,si,hi`.
- Partial failure writes `Failed Videos` and exits 0.
- Strict failure exits 1.
- Invalid input exits 2.
- Missing input file exits 2.
- `--help` exits 0.
- `version` exits 0 and includes version metadata.

Exit criteria:

- `go test ./cmd/yt-transcript-md ./tests/e2e/...` passes.
- E2E tests do not hit the real network by default.
- Coverage gate is enforced at an agreed threshold. Recommended first gate: 70%.

### Phase 6: CI and Security

Goal: replace Python CI with Go quality gates.

CI jobs:

- Module hygiene: `go mod verify`, `make tidy-check`.
- Lint: `go vet ./...`, `golangci-lint run ./...`.
- Test and build: `go test ./...`, coverage gate, `go build`.
- Compatibility: at least `ubuntu-latest` and `macos-latest`.
- Vulnerability scan: `govulncheck ./...`.
- GoReleaser config check.

Recommended workflow conventions from `mcp-config-tui`:

- Pin important GitHub Actions by SHA for supply-chain stability.
- Use `actions/setup-go` with `go-version-file: go.mod`.
- Use `cache-dependency-path: go.sum`.
- Add CI concurrency cancellation for PR updates.

Exit criteria:

- PR CI blocks merge on tests, lint, tidy drift, and build failure.
- Release CI blocks release on tests, lint, vulnerability scan, and GoReleaser config errors.

### Phase 7: Release and Homebrew Publishing

Goal: publish GitHub releases, checksums, cross-platform binaries, and Homebrew formula updates.

Use GoReleaser v2 with a `.goreleaser.yml` modeled after `mcp-config-tui`.

Target artifacts:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`
- Optional: `windows/amd64` if Windows support is accepted.

Recommended `.goreleaser.yml` shape:

```yaml
version: 2

project_name: yt-transcript-md

before:
  hooks:
    - go mod tidy

builds:
  - id: yt-transcript-md
    main: ./cmd/yt-transcript-md
    binary: yt-transcript-md
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    mod_timestamp: "{{ .CommitTimestamp }}"
    ldflags:
      - >-
        -s -w
        -X github.com/nawodyaishan/yt-transcript-md/internal/version.Version={{ .Version }}
        -X github.com/nawodyaishan/yt-transcript-md/internal/version.Commit={{ .ShortCommit }}
        -X github.com/nawodyaishan/yt-transcript-md/internal/version.Date={{ .Date }}
        -X github.com/nawodyaishan/yt-transcript-md/internal/version.GoVersion={{ .Env.GOVERSION }}

archives:
  - id: default
    formats:
      - tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
      - "^chore:"

brews:
  - repository:
      owner: nawodyaishan
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/nawodyaishan/yt-transcript-md"
    description: "Export available YouTube transcripts to Markdown"
    license: MIT
    test: |
      system "#{bin}/yt-transcript-md", "--help"
      system "#{bin}/yt-transcript-md", "version"
    install: |
      bin.install "yt-transcript-md"
```

Homebrew release requirements:

- Create or reuse `github.com/nawodyaishan/homebrew-tap`.
- Add repository secret `HOMEBREW_TAP_TOKEN` with write access to the tap repo.
- Keep `GITHUB_TOKEN` for release creation in the source repo.
- Add `brew install nawodyaishan/tap/yt-transcript-md` to README after the first release.
- Add a release dry-run checklist before tagging.

Release workflow:

- Trigger on tags matching `v*`.
- Run pre-release lint, tests, vulnerability scan, and `goreleaser check`.
- Capture Go version into `GOVERSION`.
- Run `goreleaser release --clean`.

Tagging flow:

```bash
make verify
make tag V=v0.1.0 MSG="Release v0.1.0"
git push origin v0.1.0
```

Exit criteria:

- GitHub Release contains binaries, archives, and `checksums.txt`.
- Homebrew tap receives or updates `Formula/yt-transcript-md.rb`.
- `brew install nawodyaishan/tap/yt-transcript-md` works on macOS.
- Formula test passes with `brew test yt-transcript-md`.

### Phase 8: Python Removal

Goal: finish the migration without leaving two active implementations.

Tasks:

- Update README to Go install, source build, binary usage, and Homebrew usage.
- Remove `pyproject.toml`, `uv.lock`, `src/`, and Python tests after Go behavior is accepted.
- Replace Python Makefile targets with Go targets.
- Ensure old command examples still work through the Go binary.
- Keep `docs/tech_spec_v1.md` only if it is clearly marked historical, or replace it with a Go architecture spec.

Exit criteria:

- Repo has one supported implementation.
- README and CI match the Go implementation.
- `make verify` is the single local quality command.

## QA Strategy

### Unit tests

Required packages:

- `internal/input`
- `internal/markdown`
- `internal/transcript`
- `internal/app`
- `internal/cli`
- `internal/version`

Minimum unit test cases:

- URL extraction accepts current supported YouTube URL forms.
- Invalid non-YouTube URL fails.
- Invalid ID length fails.
- Input splitting handles commas, newlines, whitespace, and angle brackets.
- Dedupe preserves first occurrence order.
- Language parsing defaults to `en`.
- Timestamp formatting handles under one hour and over one hour.
- Markdown rendering normalizes whitespace.
- Markdown rendering includes failures.
- Strict mode stops after first failure and returns non-zero.
- Non-strict mode records failures and continues.
- Retry behavior retries expected count and respects context cancellation.

### Golden tests

Use golden files for Markdown and e2e CLI output.

Recommended golden cases:

- Plain single transcript.
- Timestamped transcript.
- Multiple videos.
- Mixed success and failure.
- Empty transcript.
- Paragraph wrapping.

### Fuzz tests

Add fuzz tests for parser functions.

Seed corpus:

- Valid raw ID.
- `https://youtu.be/dQw4w9WgXcQ`
- `https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=12`
- `https://www.youtube.com/shorts/dQw4w9WgXcQ`
- `<https://youtu.be/dQw4w9WgXcQ>`
- Empty string.
- Very long random string.
- Unicode-containing string.
- Comma and newline mixed input.

### E2E tests

Use compiled-binary tests with a fake provider mode.

Implementation options:

- Build test binary with a test-only provider via build tags.
- Use a local fixture server if the chosen provider supports injectable HTTP clients.
- Add an internal provider registry that tests can swap.

Avoid real YouTube network access in default CI. Real transcript checks should be opt-in:

```bash
go test -tags=integration ./internal/transcript/...
```

### Coverage gate

Initial gate:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Recommended thresholds:

- Phase 3: 60%
- Phase 5: 70%
- Phase 8: 80%

### Manual QA

Before first Go release:

- Run with one valid public video ID.
- Run with a video that has no transcript.
- Run with multiple languages.
- Run with `--strict`.
- Run with output path in a missing directory.
- Install generated archive locally and run `--help`, `version`, and one export.
- Install via Homebrew tap and run `brew test`.

## Quality Bar

Local command:

```bash
make verify
```

`make verify` should run:

- `go mod verify`
- `go mod tidy` drift check
- `gofmt` check
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./...`
- coverage gate
- `go build`

Recommended `.golangci.yml` categories:

- `errcheck`
- `govet`
- `ineffassign`
- `staticcheck`
- `unused`
- `gocritic`
- `misspell`
- `prealloc`
- `unconvert`
- `whitespace`

Keep lint strict enough to catch real defects, but avoid style-only churn during the migration.

## Risk Register

### YouTube transcript endpoint instability

Risk: YouTube transcript endpoints and blocking behavior can change.

Mitigation:

- Isolate provider behind an interface.
- Keep integration tests opt-in.
- Document known endpoint limitations.
- Keep error messages clear and actionable.

### Third-party Go transcript library maturity

Risk: available Go libraries are young and may lack stable behavior.

Mitigation:

- Perform a provider spike before committing.
- Keep provider swappable.
- Avoid leaking provider-specific types outside `internal/transcript`.

### Output regressions

Risk: Markdown changes break users who depend on current format.

Mitigation:

- Use golden tests.
- Keep compatibility fixtures from the Python version.
- Document intentional differences.

### Release token failure

Risk: GitHub Release succeeds but Homebrew tap update fails.

Mitigation:

- Add `goreleaser check` to CI.
- Verify `HOMEBREW_TAP_TOKEN` before first tag.
- Use a release candidate tag or dry-run before the first public release.

### Two implementations linger

Risk: Python and Go code both remain and drift.

Mitigation:

- Treat Python removal as an explicit phase.
- Do not publish the Go release as stable until README, CI, and package metadata point to Go.

## Implementation Order Checklist

- [ ] Add Go module and skeleton CLI.
- [ ] Add version package and linker flags.
- [ ] Port parser with table-driven tests and fuzz seeds.
- [ ] Port Markdown renderer with golden tests.
- [ ] Implement app orchestration with fake provider tests.
- [ ] Spike and choose transcript provider.
- [ ] Implement retrying YouTube provider behind interface.
- [ ] Add compiled-binary e2e tests.
- [ ] Add Go Makefile scripts modeled after `mcp-config-tui`.
- [ ] Add CI for tidy, lint, test, coverage, build, compatibility, and vulnerability scan.
- [ ] Add GoReleaser config.
- [ ] Add Homebrew tap publishing.
- [ ] Update README with Go and Homebrew install paths.
- [ ] Remove Python implementation after acceptance.

## Definition of Done

The migration is complete when:

- `yt-transcript-md` is a Go binary with no required Python runtime.
- Existing CLI examples continue to work.
- `make verify` passes locally.
- CI passes on Linux and macOS.
- Default tests do not depend on real YouTube network access.
- A tagged release publishes GitHub artifacts and checksums.
- Homebrew install works through `nawodyaishan/homebrew-tap`.
- README documents source build, binary release, and Homebrew install.
