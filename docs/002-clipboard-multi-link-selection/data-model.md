# Data Model: Clipboard Multi-Link Selection Prompt

## Existing Models

### VideoInput

Current input identity:

- `Original string`: the user-provided URL or video ID after splitting and trimming
- `VideoID string`: canonical 11-character YouTube video ID

No field change is planned. Selection operates on parsed and deduplicated `VideoInput` values.

## Input Extraction

Clipboard mode adds an extraction step before canonical parsing.

Rules:

- Extract valid YouTube URL candidates and raw 11-character video ID candidates from arbitrary clipboard text.
- Preserve first-seen order from the clipboard.
- Strip surrounding punctuation or markup that is not part of the URL or ID.
- Pass extracted candidates through the canonical parser and deduper before selection.
- Return the existing no-valid-input error if no valid YouTube videos are extracted.
- Do not apply prose extraction to explicit `--links`, `--input-file`, or `export` inputs in this slice.

## New Models

### ClipboardSelectionMode

Represents how the default clipboard workflow should reduce a multi-video clipboard input.

Values:

- `all`: process every parsed unique video
- `one`: process one indexed video
- `recent`: process the first N unique videos in parsed clipboard order
- `cancel`: exit before network or output side effects

### ClipboardSelection

Represents the resolved selection decision.

- `Mode ClipboardSelectionMode`
- `Index int`: one-based video index for `one`
- `Count int`: selected count for `recent`

Rules:

- `Index` is required only for `one`.
- `Count` is required only for `recent`.
- `all` and `cancel` ignore `Index` and `Count`.
- Validation must reject out-of-range indexes, zero or negative counts, and counts greater than the detected video count.

### ClipboardSelector

Interface for resolving a selection in default clipboard mode.

- `Select(videos []models.VideoInput) (ClipboardSelection, error)`

Rules:

- The selector receives already parsed and deduplicated videos.
- The selector must not fetch metadata or transcripts.
- Tests may use a fake selector.
- Production CLI may use a terminal selector or a selector derived from a non-interactive flag.

### ClipboardSelectionFlag

Root-only CLI input for resolving multi-video clipboard selection in non-interactive mode.

Accepted values:

- `all`
- `one:<index>`
- `recent:<count>`

Rules:

- `index` is one-based and must refer to a detected unique video.
- `count` must be positive and no greater than the detected unique video count.
- Invalid values fail before network or output side effects.
- The flag applies only to default clipboard mode.

## Existing Models Not Changed

### TranscriptDocument

No change planned. Only selected videos become transcript documents.

### VideoMetadata

No change planned. Metadata is fetched only after selection.

### FailedVideo

No change planned. Prompt cancellation and invalid selection are command errors, not failed video entries.
