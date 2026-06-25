package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/nawodyaishan/yt-transcript-md/internal/history"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
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
	clipboardSelection string
	historySource      string
	historyLimit       int
	noHistory          bool
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
	rootCmd.Flags().StringVar(&opts.clipboardSelection, "clipboard-selection", "", "Resolve multi-link clipboard input without prompting: all, one:<index>, or recent:<count>")
	rootCmd.Flags().StringVar(&opts.historySource, "history-source", "auto", "Clipboard history source: auto, current, copyq, cliphist, or gpaste")
	rootCmd.Flags().IntVar(&opts.historyLimit, "history-limit", history.DefaultLimit, "Maximum clipboard history entries to scan per provider")
	rootCmd.Flags().BoolVar(&opts.noHistory, "no-history", false, "Disable clipboard history scanning and use only the current clipboard")

	// Keep flag-based root export for compatibility, and use the clipboard
	// workflow only for a true no-flag invocation.
	originalRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If a subcommand was specified, Execute will handle it.
		if cmd.Flags().Changed("links") || cmd.Flags().Changed("input-file") {
			if err := rejectClipboardOnlyFlags(cmd); err != nil {
				return err
			}
			return runExport()
		}
		if isClipboardInvocation(cmd) && len(args) == 0 {
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
	selector, err := getClipboardSelector(opts.clipboardSelection)
	if err != nil {
		return err
	}

	historyOpts, err := historyOptionsFromFlags()
	if err != nil {
		return err
	}
	if !isInteractive(os.Stdin) || !isInteractive(os.Stdout) {
		if opts.clipboardSelection == "" && !historyFlagsChanged() {
			historyOpts.NoHistory = true
		}
	}

	rawInput, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %w", err)
	}

	candidates, warnings, err := history.CollectCandidates(ctx, rawInput, getHistoryProviders(), historyOpts)
	for _, warning := range warnings {
		_, _ = fmt.Fprintf(os.Stdout, "[WARN] Clipboard history provider warning: %v\n", warning)
	}
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("input error: no valid YouTube links or video IDs provided")
	}

	selectedVideos, err := selectHistoryCandidates(candidates, selector)
	if err != nil {
		return err
	}

	return app.ExportClipboardVideos(ctx, exportOptionsFromFlags(), clipboard, selectedVideos, provider, metadataProvider, os.Stdout)
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

func rejectClipboardOnlyFlags(cmd *cobra.Command) error {
	switch {
	case cmd.Flags().Changed("clipboard-selection"):
		return errClipboardSelectionExplicitInput
	case cmd.Flags().Changed("history-source"), cmd.Flags().Changed("history-limit"), cmd.Flags().Changed("no-history"):
		return fmt.Errorf("--history-source, --history-limit, and --no-history can only be used with the default clipboard workflow")
	default:
		return nil
	}
}

func isClipboardInvocation(cmd *cobra.Command) bool {
	return cmd.Flags().NFlag() == 0 ||
		cmd.Flags().Changed("clipboard-selection") ||
		cmd.Flags().Changed("history-source") ||
		cmd.Flags().Changed("history-limit") ||
		cmd.Flags().Changed("no-history")
}

func historyFlagsChanged() bool {
	return opts.historySource != "" && opts.historySource != history.SourceAuto ||
		opts.historyLimit != history.DefaultLimit ||
		opts.noHistory
}

func historyOptionsFromFlags() (history.Options, error) {
	return history.ValidateOptions(history.Options{
		Source:    opts.historySource,
		Limit:     opts.historyLimit,
		NoHistory: opts.noHistory,
	})
}

func selectHistoryCandidates(candidates []history.Candidate, selector app.ClipboardSelector) ([]models.VideoInput, error) {
	videos := candidateVideos(candidates)
	if len(candidates) <= 1 {
		if fixedSelector, ok := selector.(app.FixedClipboardSelector); ok {
			return app.ApplyClipboardSelection(videos, fixedSelector.FixedSelection())
		}
		return videos, nil
	}
	if fixedSelector, ok := selector.(app.FixedClipboardSelector); ok {
		return app.ApplyClipboardSelection(videos, fixedSelector.FixedSelection())
	}
	if selector == nil {
		return nil, fmt.Errorf("clipboard selection is required for %d detected videos", len(candidates))
	}
	if _, ok := selector.(terminalClipboardSelector); ok {
		selected, err := historyTUISelector{}.Select(candidates)
		if err != nil {
			return nil, err
		}
		return candidateVideos(selected), nil
	}
	selection, err := selector.Select(videos)
	if err != nil {
		return nil, err
	}
	return app.ApplyClipboardSelection(videos, selection)
}

func candidateVideos(candidates []history.Candidate) []models.VideoInput {
	videos := make([]models.VideoInput, 0, len(candidates))
	for _, candidate := range candidates {
		videos = append(videos, candidate.Video)
	}
	return videos
}
