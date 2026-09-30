// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LangSelectorModel is a Bubbletea sub-model for selecting active language preference.
type LangSelectorModel struct {
	Languages []LanguageInfo
	Cursor    int
	Selected  string
	Canceled  bool
	Width     int
	Height    int
}

var (
	langTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	activeBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorAccent).
				Padding(0, 1)

	langCursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)
)

// NewLangSelectorModel creates a new language selection sub-menu model.
func NewLangSelectorModel() LangSelectorModel {
	langs := GetSupportedLanguages()
	cursor := 0
	for i, l := range langs {
		if l.Code == ActiveLanguage {
			cursor = i
			break
		}
	}

	return LangSelectorModel{
		Languages: langs,
		Cursor:    cursor,
		Width:     GetTermWidth(),
		Height:    GetTermHeight(),
	}
}

func (m LangSelectorModel) Init() tea.Cmd {
	return nil
}

func (m LangSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			m.Canceled = true
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				m.Cursor = len(m.Languages) - 1
			}

		case "down", "j":
			if m.Cursor < len(m.Languages)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "1":
			if len(m.Languages) >= 1 {
				m.Cursor = 0
				m.Selected = m.Languages[0].Code
				return m, tea.Quit
			}

		case "2":
			if len(m.Languages) >= 2 {
				m.Cursor = 1
				m.Selected = m.Languages[1].Code
				return m, tea.Quit
			}

		case "enter":
			if m.Cursor >= 0 && m.Cursor < len(m.Languages) {
				m.Selected = m.Languages[m.Cursor].Code
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m LangSelectorModel) View() string {
	var b strings.Builder

	termW := m.Width
	if termW <= 0 {
		termW = GetTermWidth()
	}
	boxW := termW - 4
	if boxW < 36 {
		boxW = 36
	}

	langBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(1, 2).
		Foreground(ColorFg)

	title := T("lang_selector_title")
	b.WriteString(langTitleStyle.Render(title))
	b.WriteString("\n\n")

	for i, l := range m.Languages {
		shortcut := fmt.Sprintf("[%d]", i+1)
		codeBadge := fmt.Sprintf("[%s]", strings.ToUpper(l.Code))
		lineStr := fmt.Sprintf("%s %s %s (%s)", shortcut, codeBadge, l.NativeName, l.Name)

		if l.Code == ActiveLanguage {
			lineStr += " " + AtomicBadge(T("lang_active_badge"), activeBadgeStyle)
		}

		if i == m.Cursor {
			b.WriteString(langCursorStyle.Render("▶ "))
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorFg).Render(lineStr))
			b.WriteString("\n")
		} else {
			b.WriteString("  ")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(lineStr))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(T("lang_selector_hint")))

	return "\n" + langBoxStyle.Render(b.String()) + "\n"
}
