package history

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nawodyaishan/yt-transcript-md/internal/input"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

const (
	SourceAuto     = "auto"
	SourceCurrent  = "current"
	SourceCopyQ    = "copyq"
	SourceCliphist = "cliphist"
	SourceGPaste   = "gpaste"

	DefaultLimit = 50
)

type Options struct {
	Source    string
	Limit     int
	NoHistory bool
}

type Entry struct {
	Provider string
	ID       string
	Text     string
	Preview  string
	Rank     int
}

type Candidate struct {
	Video         models.VideoInput
	Source        string
	SourceEntryID string
	SourceRank    int
	Preview       string
}

type Provider interface {
	Name() string
	Available(ctx context.Context) error
	Entries(ctx context.Context, limit int) ([]Entry, error)
}

type CommandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct {
	Timeout time.Duration
}

func (r ExecRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return exec.CommandContext(runCtx, name, args...).Output()
}

func DefaultProviders(runner CommandRunner) []Provider {
	if runner == nil {
		runner = ExecRunner{}
	}
	return []Provider{
		NewCopyQProvider(runner),
		NewCliphistProvider(runner),
		NewGPasteProvider(runner),
	}
}

func ValidateOptions(opts Options) (Options, error) {
	if opts.Source == "" {
		opts.Source = SourceAuto
	}
	if opts.Limit == 0 {
		opts.Limit = DefaultLimit
	}
	if opts.Limit < 0 {
		return opts, fmt.Errorf("history limit must be greater than zero")
	}
	switch opts.Source {
	case SourceAuto, SourceCurrent, SourceCopyQ, SourceCliphist, SourceGPaste:
		return opts, nil
	default:
		return opts, fmt.Errorf("invalid history source %q; use auto, current, copyq, cliphist, or gpaste", opts.Source)
	}
}

func CollectCandidates(ctx context.Context, currentText string, providers []Provider, opts Options) ([]Candidate, []error, error) {
	opts, err := ValidateOptions(opts)
	if err != nil {
		return nil, nil, err
	}

	var warnings []error
	var candidates []Candidate
	seen := make(map[string]bool)

	addTextCandidates := func(source string, entryID string, rank int, text string) {
		videos, err := input.ExtractClipboardVideoInputs(text)
		if err != nil {
			return
		}
		preview := sanitizePreview(text)
		for _, video := range videos {
			if seen[video.VideoID] {
				continue
			}
			seen[video.VideoID] = true
			candidates = append(candidates, Candidate{
				Video:         video,
				Source:        source,
				SourceEntryID: entryID,
				SourceRank:    rank,
				Preview:       preview,
			})
		}
	}

	addTextCandidates(SourceCurrent, "current", 0, currentText)

	if opts.NoHistory || opts.Source == SourceCurrent {
		return candidates, warnings, nil
	}

	selectedProviders := providersForSource(providers, opts.Source)
	if opts.Source != SourceAuto && len(selectedProviders) == 0 {
		return candidates, warnings, fmt.Errorf("history source %q is not configured", opts.Source)
	}

	for _, provider := range selectedProviders {
		if err := provider.Available(ctx); err != nil {
			if opts.Source == SourceAuto {
				warnings = append(warnings, fmt.Errorf("%s unavailable: %w", provider.Name(), err))
				continue
			}
			return candidates, warnings, fmt.Errorf("%s unavailable: %w", provider.Name(), err)
		}
		entries, err := provider.Entries(ctx, opts.Limit)
		if err != nil {
			if opts.Source == SourceAuto {
				warnings = append(warnings, fmt.Errorf("%s failed: %w", provider.Name(), err))
				continue
			}
			return candidates, warnings, fmt.Errorf("%s failed: %w", provider.Name(), err)
		}
		for _, entry := range entries {
			addTextCandidates(provider.Name(), entry.ID, entry.Rank, entry.Text)
		}
	}

	return candidates, warnings, nil
}

func providersForSource(providers []Provider, source string) []Provider {
	if source == SourceAuto {
		return providers
	}
	var selected []Provider
	for _, provider := range providers {
		if provider.Name() == source {
			selected = append(selected, provider)
		}
	}
	return selected
}

func sanitizePreview(text string) string {
	preview := strings.Join(strings.Fields(text), " ")
	if len(preview) > 120 {
		preview = preview[:117] + "..."
	}
	return preview
}

type commandProvider struct {
	name    string
	binary  string
	runner  CommandRunner
	entries func(context.Context, CommandRunner, int) ([]Entry, error)
}

func (p commandProvider) Name() string {
	return p.name
}

func (p commandProvider) Available(ctx context.Context) error {
	if p.runner == nil {
		return errors.New("command runner is not configured")
	}
	_, err := p.runner.LookPath(p.binary)
	return err
}

func (p commandProvider) Entries(ctx context.Context, limit int) ([]Entry, error) {
	return p.entries(ctx, p.runner, limit)
}

func NewCopyQProvider(runner CommandRunner) Provider {
	return commandProvider{
		name:   SourceCopyQ,
		binary: "copyq",
		runner: runner,
		entries: func(ctx context.Context, runner CommandRunner, limit int) ([]Entry, error) {
			var entries []Entry
			for i := 0; i < limit; i++ {
				out, err := runner.Run(ctx, "copyq", "read", strconv.Itoa(i))
				if err != nil {
					if i == 0 {
						return nil, err
					}
					break
				}
				text := strings.TrimRight(string(out), "\n")
				if strings.TrimSpace(text) == "" {
					continue
				}
				entries = append(entries, Entry{
					Provider: SourceCopyQ,
					ID:       strconv.Itoa(i),
					Text:     text,
					Preview:  sanitizePreview(text),
					Rank:     i,
				})
			}
			return entries, nil
		},
	}
}

func NewCliphistProvider(runner CommandRunner) Provider {
	return commandProvider{
		name:   SourceCliphist,
		binary: "cliphist",
		runner: runner,
		entries: func(ctx context.Context, runner CommandRunner, limit int) ([]Entry, error) {
			out, err := runner.Run(ctx, "cliphist", "list")
			if err != nil {
				return nil, err
			}
			lines := nonEmptyLines(string(out), limit)
			var entries []Entry
			for i, line := range lines {
				decoded, err := runner.Run(ctx, "cliphist", "decode", line)
				if err != nil {
					continue
				}
				text := strings.TrimRight(string(decoded), "\n")
				if strings.TrimSpace(text) == "" {
					continue
				}
				entries = append(entries, Entry{
					Provider: SourceCliphist,
					ID:       cliphistID(line),
					Text:     text,
					Preview:  sanitizePreview(text),
					Rank:     i,
				})
			}
			return entries, nil
		},
	}
}

func NewGPasteProvider(runner CommandRunner) Provider {
	return commandProvider{
		name:   SourceGPaste,
		binary: "gpaste-client",
		runner: runner,
		entries: func(ctx context.Context, runner CommandRunner, limit int) ([]Entry, error) {
			out, err := runner.Run(ctx, "gpaste-client", "history", "--oneline")
			if err != nil {
				return nil, err
			}
			lines := nonEmptyLines(string(out), limit)
			var entries []Entry
			for i, line := range lines {
				id := gpasteID(line)
				if id == "" {
					id = strconv.Itoa(i)
				}
				decoded, err := runner.Run(ctx, "gpaste-client", "get", id, "--raw")
				text := ""
				if err == nil {
					text = strings.TrimRight(string(decoded), "\n")
				}
				if strings.TrimSpace(text) == "" {
					text = gpastePreviewText(line)
				}
				if strings.TrimSpace(text) == "" {
					continue
				}
				entries = append(entries, Entry{
					Provider: SourceGPaste,
					ID:       id,
					Text:     text,
					Preview:  sanitizePreview(text),
					Rank:     i,
				})
			}
			return entries, nil
		},
	}
}

func nonEmptyLines(raw string, limit int) []string {
	var result []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func cliphistID(line string) string {
	parts := strings.SplitN(line, "\t", 2)
	return strings.TrimSpace(parts[0])
}

func gpasteID(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	id := strings.TrimSuffix(fields[0], ":")
	return id
}

func gpastePreviewText(line string) string {
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return ""
	}
	return strings.Join(fields[1:], " ")
}
