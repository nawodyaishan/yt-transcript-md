# Data Model: Clipboard-First Rebrand, Default Save, Rich Metadata, and Colored Feedback

## Existing Models

### VideoInput

Current input identity:

- `Original string`: the user-provided URL or video ID after splitting and trimming
- `VideoID string`: canonical 11-character YouTube video ID

No change planned.

### TranscriptSnippet

Current transcript snippet:

- `Text string`
- `Start float64`
- `Duration float64`

No change planned.

### TranscriptDocument

Current transcript document:

- `Video VideoInput`
- `Language string`
- `LanguageCode string`
- `IsGenerated bool`
- `Snippets []TranscriptSnippet`

Planned addition:

- `Metadata VideoMetadata`

Metadata uses the zero value when unavailable, so existing transcript rendering and tests can remain simple.

## New Models

### VideoMetadata

Represents optional public metadata for a video.

- `Title string`
- `AuthorName string`
- `AuthorURL string`
- `ProviderName string`
- `ProviderURL string`
- `ThumbnailURL string`
- `ThumbnailWidth int`
- `ThumbnailHeight int`
- `CacheAgeSeconds int`

Rules:

- All fields are optional.
- Empty metadata must not fail transcript rendering.
- Do not store or render oEmbed `html`.
- Do not store credentials, cookies, or private user data.

## FailedVideo

Current failure model:

- `Original string`
- `Reason string`

No change planned. Metadata warnings do not become `FailedVideo` entries unless transcript fetching also fails.
