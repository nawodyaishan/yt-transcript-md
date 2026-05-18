package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yt-transcript-md",
	Short: "Copy a YouTube link, run one command, get a Markdown transcript",
	Long: `Clipboard-first YouTube transcript capture.

Copy a YouTube link or video ID, run yt-transcript-md with no arguments, and
the command saves transcripts.md while also copying the generated Markdown back
to your clipboard. Use flags or the export command for explicit file and batch
workflows.`,
	Example: `  # Default workflow: read clipboard, save transcripts.md, copy Markdown back
  yt-transcript-md

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
