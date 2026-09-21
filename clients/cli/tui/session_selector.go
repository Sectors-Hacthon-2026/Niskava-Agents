// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"fmt"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionSelectorModel is an interactive Bubbletea menu to pick and resume past chat sessions.
type SessionSelectorModel struct {
	Sessions        []db.ChatSession
	Cursor          int
	SelectedSession *db.ChatSession
	Canceled        bool
}

var (
	sessionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22C55E")).
			Padding(1, 2).
			Foreground(lipgloss.Color("#F8FAFC"))

	sessionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF87"))

	sessionActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00FF87"))

	sessionCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#4ADE80"))

	sessionMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#64748B"))
)

// NewSessionSelectorModel creates an interactive selector model for chat sessions.
func NewSessionSelectorModel(sessions []db.ChatSession) SessionSelectorModel {
	return SessionSelectorModel{
		Sessions: sessions,
		Cursor:   0,
	}
}

func (m SessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			m.Canceled = true
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else if len(m.Sessions) > 0 {
				m.Cursor = len(m.Sessions) - 1
			}

		case "down", "j":
			if m.Cursor < len(m.Sessions)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "enter":
			if len(m.Sessions) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Sessions) {
				selected := m.Sessions[m.Cursor]
				m.SelectedSession = &selected
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SessionSelectorModel) View() string {
	var b strings.Builder

	title := T("session_selector_title")
	b.WriteString(sessionTitleStyle.Render(title) + "\n\n")

	if len(m.Sessions) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render(T("session_selector_empty")) + "\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
	}

	for i, s := range m.Sessions {
		dateStr := s.UpdatedAt
		if len(dateStr) > 16 {
			dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
		}

		preview := s.LastMessagePreview
		if len(preview) > 35 {
			preview = preview[:32] + "..."
		}
		if preview == "" {
			preview = "-"
		}

		pinBadge := ""
		if s.IsPinned {
			pinBadge = " 📌"
		}

		lineTitle := fmt.Sprintf("%-26s %s (%d msgs) [%s]", s.ID, s.Title+pinBadge, s.MessageCount, dateStr)
		previewLine := fmt.Sprintf("    ↳ %s", preview)

		if i == m.Cursor {
			b.WriteString(sessionCursorStyle.Render("▶ ") + sessionActiveStyle.Render(lineTitle) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A7F3D0")).Render(previewLine) + "\n")
		} else {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0")).Render(lineTitle) + "\n")
			b.WriteString(sessionMetaStyle.Render(previewLine) + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true).Render(T("session_selector_hint")))

	return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
}
