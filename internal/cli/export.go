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
	Short: "Advanced file and batch transcript export",
	Long: `Advanced file and batch transcript export.

Fetch transcripts from explicit links, video IDs, or an input file and
write the generated Markdown to a chosen output path.

For the simplest workflow, copy a YouTube link and run yt-transcript-md with no
arguments. This export command is for explicit file destinations, batch input,
and automation.`,
	Example: `  yt-transcript-md export --links "https://youtu.be/dQw4w9WgXcQ" --out notes.md
  yt-transcript-md export --input-file links.txt --out transcripts.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExport()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	addExportFlags(exportCmd)
	addExportFlags(rootCmd)

	// Keep flag-based root export for compatibility, and use the clipboard
	// workflow only for a true no-flag invocation.
	originalRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If a subcommand was specified, Execute will handle it.
		if cmd.Flags().Changed("links") || cmd.Flags().Changed("input-file") {
			return runExport()
		}
		if cmd.Flags().NFlag() == 0 && len(args) == 0 {
			return runClipboardExport()
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
	metadataProvider := getMetadataProvider()

	return app.Export(ctx, exportOptionsFromFlags(), provider, metadataProvider, os.Stdout)
}

func runClipboardExport() error {
	ctx := context.Background()
	provider := getProvider()
	metadataProvider := getMetadataProvider()
	clipboard := getClipboard()

	return app.ExportClipboard(ctx, exportOptionsFromFlags(), clipboard, provider, metadataProvider, os.Stdout)
}

func exportOptionsFromFlags() app.ExportOptions {
	return app.ExportOptions{
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
}
