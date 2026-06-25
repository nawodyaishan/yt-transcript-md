package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

var errClipboardSelectionExplicitInput = errors.New("--clipboard-selection can only be used with the default clipboard workflow")

type fixedClipboardSelector struct {
	selection app.ClipboardSelection
}

func (s fixedClipboardSelector) Select(videos []models.VideoInput) (app.ClipboardSelection, error) {
	return s.selection, nil
}

func (s fixedClipboardSelector) FixedSelection() app.ClipboardSelection {
	return s.selection
}

type terminalClipboardSelector struct {
	reader *bufio.Reader
	writer io.Writer
}

func getClipboardSelector(raw string) (app.ClipboardSelector, error) {
	if strings.TrimSpace(raw) != "" {
		selection, err := parseClipboardSelectionFlag(raw)
		if err != nil {
			return nil, err
		}
		return fixedClipboardSelector{selection: selection}, nil
	}

	if isInteractive(os.Stdin) && isInteractive(os.Stdout) {
		return terminalClipboardSelector{
			reader: bufio.NewReader(os.Stdin),
			writer: os.Stdout,
		}, nil
	}

	return nil, nil
}

func parseClipboardSelectionFlag(raw string) (app.ClipboardSelection, error) {
	value := strings.TrimSpace(raw)
	switch {
	case value == string(app.ClipboardSelectionAll):
		return app.ClipboardSelection{Mode: app.ClipboardSelectionAll}, nil
	case strings.HasPrefix(value, string(app.ClipboardSelectionOne)+":"):
		index, err := parsePositiveSuffix(value, string(app.ClipboardSelectionOne)+":", "video index")
		if err != nil {
			return app.ClipboardSelection{}, err
		}
		return app.ClipboardSelection{Mode: app.ClipboardSelectionOne, Index: index}, nil
	case strings.HasPrefix(value, string(app.ClipboardSelectionRecent)+":"):
		count, err := parsePositiveSuffix(value, string(app.ClipboardSelectionRecent)+":", "recent count")
		if err != nil {
			return app.ClipboardSelection{}, err
		}
		return app.ClipboardSelection{Mode: app.ClipboardSelectionRecent, Count: count}, nil
	default:
		return app.ClipboardSelection{}, fmt.Errorf("invalid --clipboard-selection value %q; use all, one:<index>, or recent:<count>", raw)
	}
}

func parsePositiveSuffix(value string, prefix string, label string) (int, error) {
	raw := strings.TrimPrefix(value, prefix)
	number, err := strconv.Atoi(raw)
	if err != nil || number < 1 {
		return 0, fmt.Errorf("invalid %s in --clipboard-selection value %q", label, value)
	}
	return number, nil
}

func (s terminalClipboardSelector) Select(videos []models.VideoInput) (app.ClipboardSelection, error) {
	_, _ = fmt.Fprintf(s.writer, "Detected %d unique YouTube videos in clipboard.\n", len(videos))
	for {
		_, _ = fmt.Fprint(s.writer, "Process [o]ne, [a]ll, [r]ecent N from clipboard order, or [c]ancel? ")
		answer, err := s.readLine()
		if err != nil {
			return app.ClipboardSelection{}, app.ErrClipboardSelectionCanceled
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "a", "all":
			return app.ClipboardSelection{Mode: app.ClipboardSelectionAll}, nil
		case "o", "one":
			return s.selectOne(videos)
		case "r", "recent":
			return s.selectRecent(len(videos))
		case "c", "cancel", "q", "quit":
			return app.ClipboardSelection{Mode: app.ClipboardSelectionCancel}, nil
		default:
			_, _ = fmt.Fprintln(s.writer, "Enter one, all, recent, or cancel.")
		}
	}
}

func (s terminalClipboardSelector) selectOne(videos []models.VideoInput) (app.ClipboardSelection, error) {
	for i, video := range videos {
		_, _ = fmt.Fprintf(s.writer, "%d. %s (%s)\n", i+1, video.VideoID, video.Original)
	}
	for {
		_, _ = fmt.Fprintf(s.writer, "Video number [1-%d] or cancel? ", len(videos))
		answer, err := s.readLine()
		if err != nil {
			return app.ClipboardSelection{}, app.ErrClipboardSelectionCanceled
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "c" || answer == "cancel" || answer == "q" || answer == "quit" {
			return app.ClipboardSelection{Mode: app.ClipboardSelectionCancel}, nil
		}
		index, err := strconv.Atoi(answer)
		if err == nil && index >= 1 && index <= len(videos) {
			return app.ClipboardSelection{Mode: app.ClipboardSelectionOne, Index: index}, nil
		}
		_, _ = fmt.Fprintf(s.writer, "Enter a number from 1 to %d, or cancel.\n", len(videos))
	}
}

func (s terminalClipboardSelector) selectRecent(videoCount int) (app.ClipboardSelection, error) {
	for {
		_, _ = fmt.Fprintf(s.writer, "How many videos from clipboard order [1-%d], or cancel? ", videoCount)
		answer, err := s.readLine()
		if err != nil {
			return app.ClipboardSelection{}, app.ErrClipboardSelectionCanceled
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "c" || answer == "cancel" || answer == "q" || answer == "quit" {
			return app.ClipboardSelection{Mode: app.ClipboardSelectionCancel}, nil
		}
		count, err := strconv.Atoi(answer)
		if err == nil && count >= 1 && count <= videoCount {
			return app.ClipboardSelection{Mode: app.ClipboardSelectionRecent, Count: count}, nil
		}
		_, _ = fmt.Fprintf(s.writer, "Enter a number from 1 to %d, or cancel.\n", videoCount)
	}
}

func (s terminalClipboardSelector) readLine() (string, error) {
	line, err := s.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if err != nil && strings.TrimSpace(line) == "" {
		return "", err
	}
	return line, nil
}

func isInteractive(file *os.File) bool {
	if file == nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}
