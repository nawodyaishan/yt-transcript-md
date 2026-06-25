package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/nawodyaishan/yt-transcript-md/internal/history"
)

type historyTUISelector struct{}

func (historyTUISelector) Select(candidates []history.Candidate) ([]history.Candidate, error) {
	model := newHistoryPickerModel(candidates)
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return nil, err
	}
	finalModel, ok := result.(historyPickerModel)
	if !ok {
		return nil, fmt.Errorf("history picker returned unexpected model")
	}
	if finalModel.canceled {
		return nil, app.ErrClipboardSelectionCanceled
	}
	selected := finalModel.selectedCandidates()
	if len(selected) == 0 {
		return nil, fmt.Errorf("no videos selected")
	}
	return selected, nil
}

type historyPickerModel struct {
	candidates []history.Candidate
	cursor     int
	selected   map[int]bool
	search     textinput.Model
	firstNMode bool
	message    string
	canceled   bool
	done       bool
}

func newHistoryPickerModel(candidates []history.Candidate) historyPickerModel {
	search := textinput.New()
	search.Placeholder = "search videos"
	search.CharLimit = 80
	search.Width = 40
	return historyPickerModel{
		candidates: candidates,
		selected:   make(map[int]bool),
		search:     search,
	}
}

func (m historyPickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m historyPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.firstNMode {
			return m.updateFirstN(key)
		}
		if m.search.Focused() {
			switch key.String() {
			case "ctrl+c":
				m.canceled = true
				m.done = true
				return m, tea.Quit
			case "esc", "enter":
				m.search.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(key)
			if m.cursor >= len(m.visibleIndexes()) {
				m.cursor = 0
			}
			return m, cmd
		}
		switch key.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			m.done = true
			return m, tea.Quit
		case "enter":
			if len(m.selected) == 0 {
				m.message = "Select at least one video, or press q to cancel."
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
			return m, nil
		case "down", "j":
			m.moveCursor(1)
			return m, nil
		case " ":
			visible := m.visibleIndexes()
			if len(visible) == 0 {
				return m, nil
			}
			index := visible[m.cursor]
			m.selected[index] = !m.selected[index]
			if !m.selected[index] {
				delete(m.selected, index)
			}
			return m, nil
		case "a":
			for _, index := range m.visibleIndexes() {
				m.selected[index] = true
			}
			return m, nil
		case "n":
			m.firstNMode = true
			m.search.SetValue("")
			m.search.Placeholder = "first N visible videos"
			m.search.Focus()
			m.message = "Enter how many visible videos to select."
			return m, textinput.Blink
		case "/":
			m.search.Focus()
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.cursor >= len(m.visibleIndexes()) {
		m.cursor = 0
	}
	return m, cmd
}

func (m historyPickerModel) updateFirstN(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c", "esc", "q":
		m.firstNMode = false
		m.search.SetValue("")
		m.search.Blur()
		m.message = ""
		return m, nil
	case "enter":
		count, err := strconv.Atoi(strings.TrimSpace(m.search.Value()))
		visible := m.visibleIndexes()
		if err != nil || count < 1 || count > len(visible) {
			m.message = fmt.Sprintf("Enter a number from 1 to %d.", len(visible))
			return m, nil
		}
		for _, index := range visible[:count] {
			m.selected[index] = true
		}
		m.firstNMode = false
		m.search.SetValue("")
		m.search.Blur()
		m.message = fmt.Sprintf("Selected first %d visible video(s).", count)
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(key)
	return m, cmd
}

func (m *historyPickerModel) moveCursor(delta int) {
	visible := m.visibleIndexes()
	if len(visible) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = len(visible) - 1
	}
	if m.cursor >= len(visible) {
		m.cursor = 0
	}
}

func (m historyPickerModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString("Select YouTube videos from clipboard history\n\n")
	if m.firstNMode {
		b.WriteString(m.search.View())
		b.WriteString("\n\n")
	} else {
		b.WriteString("Search: ")
		b.WriteString(m.search.View())
		b.WriteString("\n\n")
	}
	visible := m.visibleIndexes()
	if len(visible) == 0 {
		b.WriteString("No videos match the current search.\n")
	} else {
		for row, index := range visible {
			candidate := m.candidates[index]
			cursor := " "
			if row == m.cursor {
				cursor = ">"
			}
			check := " "
			if m.selected[index] {
				check = "x"
			}
			_, _ = fmt.Fprintf(&b, "%s [%s] %s  %s  %s\n", cursor, check, candidate.Video.VideoID, candidate.Source, candidate.Preview)
		}
	}
	b.WriteString("\nspace select  / search  a all visible  n first N  enter process  q cancel\n")
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(m.message)
		b.WriteString("\n")
	}
	return b.String()
}

func (m historyPickerModel) visibleIndexes() []int {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	if m.firstNMode {
		query = ""
	}
	var indexes []int
	for i, candidate := range m.candidates {
		text := strings.ToLower(candidate.Video.VideoID + " " + candidate.Video.Original + " " + candidate.Source + " " + candidate.Preview)
		if query == "" || strings.Contains(text, query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m historyPickerModel) selectedCandidates() []history.Candidate {
	var result []history.Candidate
	for i, candidate := range m.candidates {
		if m.selected[i] {
			result = append(result, candidate)
		}
	}
	return result
}
