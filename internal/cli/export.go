package cli

import (
	"context"
	"os"

	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/spf13/cobra"
)

type exportOptions struct {
	links              string
	inputFile          string
	out                string
	languages          string
	timestamps         bool
	preserveFormatting bool
	retries            int
	retryDelaySeconds  float64
	strict             bool
}

var opts exportOptions

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Fetch transcripts and write a Markdown file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExport()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	addExportFlags(exportCmd)
	addExportFlags(rootCmd)

	// Set the root command to run export by default if no subcommand is provided
	// and flags are set.
	originalRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If a subcommand was specified, Execute will handle it.
		// If no subcommand, check if we should run export.
		if cmd.Flags().Changed("links") || cmd.Flags().Changed("input-file") {
			return runExport()
		}
		if originalRunE != nil {
			return originalRunE(cmd, args)
		}
		return cmd.Help()
	}
}

func addExportFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.links, "links", "l", "", "Comma-separated YouTube links or video IDs")
	cmd.Flags().StringVarP(&opts.inputFile, "input-file", "f", "", "Text file containing links")
	cmd.Flags().StringVarP(&opts.out, "out", "o", "transcripts.md", "Output Markdown file path")
	cmd.Flags().StringVar(&opts.languages, "languages", "en", "Comma-separated language priority list")
	cmd.Flags().BoolVar(&opts.timestamps, "timestamps", false, "Include per-snippet timestamps")
	cmd.Flags().BoolVar(&opts.preserveFormatting, "preserve-formatting", false, "Preserve YouTube transcript HTML formatting")
	cmd.Flags().IntVar(&opts.retries, "retries", 1, "Number of retries per failed transcript fetch")
	cmd.Flags().Float64Var(&opts.retryDelaySeconds, "retry-delay-seconds", 1.5, "Base retry delay in seconds")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Exit with a non-zero code if any video fails")
}

func runExport() error {
	ctx := context.Background()
	provider := getProvider()

	exportOpts := app.ExportOptions{
		Links:              opts.links,
		InputFile:          opts.inputFile,
		Out:                opts.out,
		Languages:          opts.languages,
		Timestamps:         opts.timestamps,
		PreserveFormatting: opts.preserveFormatting,
		Retries:            opts.retries,
		RetryDelaySeconds:  opts.retryDelaySeconds,
		Strict:             opts.strict,
	}

	return app.Export(ctx, exportOpts, provider, os.Stdout)
}
