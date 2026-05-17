package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yt-transcript-md",
	Short: "Export YouTube transcripts to Markdown",
	Long:  `Export available YouTube transcripts from links or video IDs into a single Markdown file.`,
	// We'll add the export logic to the root command later for backward compatibility.
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
