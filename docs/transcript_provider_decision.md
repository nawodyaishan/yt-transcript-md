# Transcript Provider Decision

## Selected Provider
`github.com/horiagug/youtube-transcript-api-go`

## Rationale
After evaluating several Go ports of the `youtube-transcript-api`, `horiagug/youtube-transcript-api-go` was selected for the following reasons:

- **Feature Completeness**: It supports the most critical features required by `yt-transcript-md`:
    - Language priority (fetching the first available language from a list).
    - `preserve_formatting` (preserving HTML tags in transcripts).
    - Metadata: Returns language name, language code, and `IsGenerated` status.
- **Maturity**: It includes a CLI and has a well-structured package layout.
- **Compatibility**: Its API most closely matches the requirements of the original Python implementation.

## Comparison Table

| Feature | `horiagug` | `hightemp` | `rahadiangg` |
| :--- | :--- | :--- | :--- |
| Language Priority | Yes | Yes | Yes |
| Preserve Formatting | Yes | No | No |
| `IsGenerated` Metadata | Yes | No | Yes |
| Context Support | No | No | No |
| Maturity | Medium | Medium | Low |

## Implementation Notes
- **Context Support**: Since the library does not natively support `context.Context`, we will use a custom `http.Client` with a timeout or wrap the calls to ensure we don't hang indefinitely.
- **Retry Logic**: We will implement our own retry logic in `internal/transcript/youtube_provider.go` to maintain control over the behavior and parameters (as per the migration plan).
- **Interface Isolation**: The provider will remain behind the `transcript.Provider` interface to allow for future swaps if a better library emerges.
