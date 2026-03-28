package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEEEEE"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)
)

type configItem struct {
	name        string
	description string
	value       string
}

type ConfigModel struct {
	items    []configItem
	cursor   int
	editing  bool
	input    string
	quitting bool
}

func NewConfigModel() ConfigModel {
	return ConfigModel{
		items: []configItem{
			{name: "API Endpoint", description: "Corridor API URL", value: "https://api.corridor.security"},
			{name: "API Key", description: "Your API key", value: "********"},
			{name: "Organization", description: "Organization ID", value: "default"},
			{name: "Output Format", description: "json, yaml, or table", value: "table"},
		},
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditMode(msg)
		}
		return m.handleNavMode(msg)
	}
	return m, nil
}

func (m ConfigModel) handleNavMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter", "e":
		m.editing = true
		m.input = m.items[m.cursor].value
	}
	return m, nil
}

func (m ConfigModel) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.items[m.cursor].value = m.input
		m.editing = false
	case "esc":
		m.editing = false
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m ConfigModel) View() string {
	if m.quitting {
		return ""
	}

	s := titleStyle.Render("⚙  Corridor Configuration") + "\n\n"

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		value := item.value
		if m.editing && i == m.cursor {
			value = m.input + "▌"
		}

		s += fmt.Sprintf("%s%s\n", cursor, style.Render(item.name))
		s += fmt.Sprintf("   %s: %s\n\n", normalStyle.Render(item.description), value)
	}

	if m.editing {
		s += helpStyle.Render("enter: save • esc: cancel")
	} else {
		s += helpStyle.Render("↑/↓: navigate • enter: edit • q: quit")
	}

	return s
}
