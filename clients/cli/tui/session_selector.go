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
			BorderForeground(ColorAccent).
			Width(76).
			Padding(1, 2).
			Foreground(ColorFg)

	sessionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent)

	sessionActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent)

	sessionCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent)

	sessionMetaStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
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

	total := len(m.Sessions)
	if total == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("session_selector_empty")) + "\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
	}

	// Sliding Viewport Window (max 5 sessions visible simultaneously)
	maxVisible := 5
	windowStart := 0
	if m.Cursor >= maxVisible {
		windowStart = m.Cursor - maxVisible + 1
	}
	windowEnd := windowStart + maxVisible
	if windowEnd > total {
		windowEnd = total
		windowStart = max(0, windowEnd-maxVisible)
	}

	if total > maxVisible {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("--- Showing %d-%d of %d sessions ---", windowStart+1, windowEnd, total)) + "\n\n")
	}

	for i := windowStart; i < windowEnd; i++ {
		s := m.Sessions[i]
		dateStr := s.UpdatedAt
		if len(dateStr) > 16 {
			dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
		}

		preview := s.LastMessagePreview
		if len(preview) > 58 {
			preview = preview[:55] + "..."
		}
		if preview == "" {
			preview = "-"
		}

		pinBadge := ""
		if s.IsPinned {
			pinBadge = " [PINNED]"
		}

		title := s.Title + pinBadge
		if len(title) > 18 {
			title = title[:15] + "..."
		}

		lineTitle := fmt.Sprintf("%-19s %-18s (%d msgs) [%s]", s.ID, title, s.MessageCount, dateStr)
		previewLine := fmt.Sprintf("    ↳ %s", preview)

		if i == m.Cursor {
			b.WriteString(sessionCursorStyle.Render("> ") + sessionActiveStyle.Render(lineTitle) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorAccent).Render(previewLine) + "\n")
		} else {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(ColorFg).Render(lineTitle) + "\n")
			b.WriteString(sessionMetaStyle.Render(previewLine) + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(T("session_selector_hint")))

	return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
}
