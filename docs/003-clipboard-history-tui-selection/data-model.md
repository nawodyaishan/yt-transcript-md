# Data Model: Clipboard History TUI Selection

## Existing Models

### VideoInput

Current video identity:

- `Original string`
- `VideoID string`

No field change is required. Selected candidates should ultimately become `VideoInput` values before transcript fetching.

## New Models

### HistorySource

Represents where clipboard text came from.

- `current`
- `copyq`
- `cliphist`
- `gpaste`

### HistoryEntry

Represents one text-like clipboard-history entry from a provider.

- `Provider string`
- `ID string`
- `Text string`
- `Preview string`
- `Rank int`

Rules:

- `Text` is used only for local extraction and should not be logged.
- `Preview` must be sanitized and truncated.
- `Rank` preserves provider order.

### VideoCandidate

Represents a deduped selectable YouTube video found in current clipboard or history.

- `Video models.VideoInput`
- `Source string`
- `SourceEntryID string`
- `SourceRank int`
- `Preview string`

Rules:

- Deduplicate by `Video.VideoID`.
- Keep the first candidate found.
- Current clipboard candidates rank before provider history candidates.
- Do not fetch metadata to populate this model.

### HistoryOptions

Runtime options for history scanning.

- `Source string`: `auto`, `current`, `copyq`, `cliphist`, or `gpaste`
- `Limit int`
- `NoHistory bool`
- `Interactive bool`

Rules:

- `Limit` must be positive.
- `NoHistory` disables providers and uses current clipboard only.
- `Interactive=false` must not open the TUI.

### TUISelection

Represents the user's selection from the TUI.

- `Selected []VideoCandidate`
- `Canceled bool`

Rules:

- Empty selection without cancellation is invalid.
- Cancellation exits before network and output side effects.

## Existing Models Not Changed

### TranscriptDocument

No change planned. Only selected videos become transcript documents.

### VideoMetadata

No change planned. Metadata is fetched only after selection.

### FailedVideo

No change planned. History provider failures are command warnings/errors, not failed video entries.
