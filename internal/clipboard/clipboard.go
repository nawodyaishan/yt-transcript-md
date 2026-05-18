package clipboard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// System reads from and writes to the operating system clipboard.
type System struct{}

type commandSpec struct {
	name string
	args []string
}

var errUnsupported = errors.New("no supported clipboard command found")

// ReadAll returns the current text content from the system clipboard.
func (System) ReadAll() (string, error) {
	spec, err := pasteCommand()
	if err != nil {
		return "", err
	}

	out, err := exec.Command(spec.name, spec.args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w", spec.name, err)
	}

	return string(out), nil
}

// WriteAll replaces the system clipboard text content.
func (System) WriteAll(text string) error {
	spec, err := copyCommand()
	if err != nil {
		return err
	}

	cmd := exec.Command(spec.name, spec.args...)
	cmd.Stdin = strings.NewReader(text)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s failed: %w: %s", spec.name, err, strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("%s failed: %w", spec.name, err)
	}

	return nil
}

func pasteCommand() (commandSpec, error) {
	return firstAvailable(pasteCandidates())
}

func copyCommand() (commandSpec, error) {
	return firstAvailable(copyCandidates())
}

func firstAvailable(candidates []commandSpec) (commandSpec, error) {
	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate.name); err == nil {
			return candidate, nil
		}
	}

	return commandSpec{}, errUnsupported
}

func pasteCandidates() []commandSpec {
	switch runtime.GOOS {
	case "darwin":
		return []commandSpec{{name: "pbpaste"}}
	case "windows":
		return []commandSpec{{name: "powershell.exe", args: []string{"-NoProfile", "-Command", "Get-Clipboard -Raw"}}}
	default:
		return unixPasteCandidates()
	}
}

func copyCandidates() []commandSpec {
	switch runtime.GOOS {
	case "darwin":
		return []commandSpec{{name: "pbcopy"}}
	case "windows":
		return []commandSpec{{name: "clip.exe"}}
	default:
		return unixCopyCandidates()
	}
}

func unixPasteCandidates() []commandSpec {
	candidates := make([]commandSpec, 0, 4)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		candidates = append(candidates, commandSpec{name: "wl-paste", args: []string{"--no-newline"}})
	}

	return append(candidates,
		commandSpec{name: "xclip", args: []string{"-out", "-selection", "clipboard"}},
		commandSpec{name: "xsel", args: []string{"--output", "--clipboard"}},
		commandSpec{name: "termux-clipboard-get"},
	)
}

func unixCopyCandidates() []commandSpec {
	candidates := make([]commandSpec, 0, 4)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		candidates = append(candidates, commandSpec{name: "wl-copy"})
	}

	return append(candidates,
		commandSpec{name: "xclip", args: []string{"-in", "-selection", "clipboard"}},
		commandSpec{name: "xsel", args: []string{"--input", "--clipboard"}},
		commandSpec{name: "termux-clipboard-set"},
	)
}
