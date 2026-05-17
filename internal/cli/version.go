package cli

import (
	"fmt"

	"github.com/nawodyaishan/yt-transcript-md/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("yt-transcript-md %s\n", version.Version)
		fmt.Printf("commit: %s\n", version.Commit)
		fmt.Printf("build date: %s\n", version.Date)
		fmt.Printf("go version: %s\n", version.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
