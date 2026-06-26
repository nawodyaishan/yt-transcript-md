package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yt-transcript-md [URL ...]",
	Args:  cobra.ArbitraryArgs,
	Short: "Copy a YouTube link, run one command, get a Markdown transcript",
	Long: `Clipboard-first YouTube transcript capture.

Copy a YouTube link or video ID, run yt-transcript-md with no arguments, and
the command saves transcripts.md while also copying the generated Markdown back
to your clipboard. In an interactive terminal, the default workflow can also
scan supported clipboard-history managers for recently copied YouTube links and
show a searchable multi-select TUI before fetching transcripts.

Pass one or more YouTube links as positional arguments to fetch transcripts
without touching the clipboard. Use --history-source and --history-limit to
control clipboard-history scanning. Use --clipboard-selection for
non-interactive clipboard runs, or use flags and the export command for
explicit file and batch workflows.`,
	Example: `  # Default workflow: read clipboard, save transcripts.md, copy Markdown back
  yt-transcript-md

  # Fetch transcripts from links passed directly as arguments
  yt-transcript-md https://youtu.be/dQw4w9WgXcQ https://youtu.be/jNQXAC9IVRw

  # Non-interactive clipboard batch selection
  yt-transcript-md --clipboard-selection recent:2

  # Scan CopyQ history and choose from recent copied YouTube links
  yt-transcript-md --history-source copyq --history-limit 25

  # Advanced workflow: write a specific link to a chosen file
  yt-transcript-md --links "https://youtu.be/dQw4w9WgXcQ" --out notes.md

  # Batch workflow with explicit export command
  yt-transcript-md export --input-file links.txt --out transcripts.md`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root command flags will be defined here or in export.go if they share logic.
}
