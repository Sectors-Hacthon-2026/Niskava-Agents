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
	FilterQuery     string
}

var (
	sessionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Width(76).
			Padding(0, 1).
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

func (m SessionSelectorModel) getFilteredSessions() []db.ChatSession {
	q := strings.TrimSpace(strings.ToLower(m.FilterQuery))
	if q == "" {
		return m.Sessions
	}
	var filtered []db.ChatSession
	for _, s := range m.Sessions {
		if strings.Contains(strings.ToLower(s.ID), q) ||
			strings.Contains(strings.ToLower(s.Title), q) ||
			strings.Contains(strings.ToLower(s.LastMessagePreview), q) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (m SessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredSessions()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			if m.FilterQuery != "" {
				m.FilterQuery = ""
				m.Cursor = 0
				return m, nil
			}
			m.Canceled = true
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else if len(filtered) > 0 {
				m.Cursor = len(filtered) - 1
			}

		case "down", "j":
			if m.Cursor < len(filtered)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "backspace":
			if len(m.FilterQuery) > 0 {
				m.FilterQuery = m.FilterQuery[:len(m.FilterQuery)-1]
				m.Cursor = 0
			}

		case "enter":
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				selected := filtered[m.Cursor]
				m.SelectedSession = &selected
			}
			return m, tea.Quit

		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				m.FilterQuery += string(msg.Runes)
				m.Cursor = 0
			}
		}
	}
	return m, nil
}

// SanitizePreviewText strips newlines, markdown tokens, and wide emojis to ensure 100% predictable 1-to-1 ASCII display width.
func SanitizePreviewText(raw string) string {
	s := strings.ReplaceAll(raw, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	s = strings.ReplaceAll(s, "###", "")
	s = strings.ReplaceAll(s, "##", "")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "❌", "")
	s = strings.ReplaceAll(s, "⚠️", "")
	s = strings.ReplaceAll(s, "🚨", "")
	s = strings.ReplaceAll(s, "⚡", "")
	s = strings.ReplaceAll(s, "👤", "")

	var b strings.Builder
	for _, r := range s {
		if (r >= 32 && r <= 126) || (r >= 160 && r <= 255) {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune(' ')
		}
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func (m SessionSelectorModel) View() string {
	var b strings.Builder

	title := T("session_selector_title")
	b.WriteString(sessionTitleStyle.Render(title) + "\n\n")

	filtered := m.getFilteredSessions()
	totalAll := len(m.Sessions)
	totalFiltered := len(filtered)

	// Always render visible search/filter bar box
	searchPlaceholder := "Ketik kode emiten/kata kunci untuk memfilter..."
	if ActiveLanguage == "en" {
		searchPlaceholder = "Type ticker or keyword to filter sessions..."
	}

	searchVal := m.FilterQuery
	if searchVal == "" {
		searchVal = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(searchPlaceholder)
	} else {
		searchVal = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(searchVal)
	}

	searchBar := fmt.Sprintf("🔍 Filter: [ %s ] (%d/%d)", searchVal, totalFiltered, totalAll)
	b.WriteString(searchBar + "\n\n")

	if totalAll == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("session_selector_empty")) + "\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
	}

	if totalFiltered == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  (Tidak ada sesi yang cocok dengan pencarian)") + "\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n"
	}

	// Sliding Viewport Window (max 5 sessions visible simultaneously)
	maxVisible := 5
	windowStart := 0
	if m.Cursor >= maxVisible {
		windowStart = m.Cursor - maxVisible + 1
	}
	windowEnd := windowStart + maxVisible
	if windowEnd > totalFiltered {
		windowEnd = totalFiltered
		windowStart = max(0, windowEnd-maxVisible)
	}

	if totalFiltered > maxVisible {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("--- Showing %d-%d of %d sessions ---", windowStart+1, windowEnd, totalFiltered)) + "\n\n")
	}

	for i := windowStart; i < windowEnd; i++ {
		s := filtered[i]
		dateStr := s.UpdatedAt
		if len(dateStr) > 16 {
			dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
		}

		// Clean and sanitize preview string (strip newlines, markdown, and emojis to prevent line breaking inside card)
		preview := SanitizePreviewText(s.LastMessagePreview)
		previewRunes := []rune(preview)
		if len(previewRunes) > 40 {
			preview = string(previewRunes[:37]) + "..."
		}
		if preview == "" {
			preview = "-"
		}

		pinBadge := ""
		if s.IsPinned {
			pinBadge = " [PINNED]"
		}

		rawTitle := SanitizePreviewText(s.Title)
		title := rawTitle + pinBadge
		titleRunes := []rune(title)
		if len(titleRunes) > 16 {
			title = string(titleRunes[:13]) + "..."
		}

		lineTitle := fmt.Sprintf("%-18s %-16s (%d msgs) [%s]", s.ID, title, s.MessageCount, dateStr)
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
