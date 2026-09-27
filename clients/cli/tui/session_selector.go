// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionSelectorModel is an interactive Bubbletea menu to pick, resume, pin, export, and manage chat sessions.
type SessionSelectorModel struct {
	AppDB             *db.DB
	Sessions          []db.ChatSession
	Cursor            int
	SelectedSession   *db.ChatSession
	Canceled          bool
	FilterQuery       string
	ConfirmDelete     bool
	DeleteTarget      *db.ChatSession
	ExportModalActive bool
	ExportFormatIndex int // 0: Markdown (.md), 1: JSON (.json), 2: Plain Text (.txt)
	StatusNotice      string
	StatusNoticeTime  time.Time
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
	return NewSessionSelectorModelWithDB(sessions, nil)
}

// NewSessionSelectorModelWithDB creates an interactive selector model backed by local DB for deletion and pinning.
func NewSessionSelectorModelWithDB(sessions []db.ChatSession, appDB *db.DB) SessionSelectorModel {
	sorted := sortSessions(sessions)
	return SessionSelectorModel{
		AppDB:    appDB,
		Sessions: sorted,
		Cursor:   0,
	}
}

// sortSessions sorts sessions pinned first, then by UpdatedAt descending.
func sortSessions(sessions []db.ChatSession) []db.ChatSession {
	cloned := make([]db.ChatSession, len(sessions))
	copy(cloned, sessions)
	sort.SliceStable(cloned, func(i, j int) bool {
		if cloned[i].IsPinned != cloned[j].IsPinned {
			return cloned[i].IsPinned
		}
		return cloned[i].UpdatedAt > cloned[j].UpdatedAt
	})
	return cloned
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

// CopyToClipboard writes text directly to OS system clipboard without CGO.
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return fmt.Errorf("no clipboard utility found")
		}
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := in.Write([]byte(text)); err != nil {
		in.Close()
		return err
	}
	in.Close()
	return cmd.Wait()
}

// ExportSessionTranscript saves a chat session transcript to ~/.niskava/exports/ in target format.
func ExportSessionTranscript(appDB *db.DB, session *db.ChatSession, formatIndex int) (string, error) {
	if session == nil {
		return "", fmt.Errorf("session is nil")
	}

	homeDir, _ := os.UserHomeDir()
	exportDir := filepath.Join(homeDir, ".niskava", "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	ext := "md"
	if formatIndex == 1 {
		ext = "json"
	} else if formatIndex == 2 {
		ext = "txt"
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanID := strings.ReplaceAll(session.ID, " ", "_")
	filename := fmt.Sprintf("niskava_session_%s_%s.%s", cleanID, timestamp, ext)
	outPath := filepath.Join(exportDir, filename)

	var messages []db.ChatMessage
	if appDB != nil {
		msgs, err := appDB.GetChatHistory(session.ID, 100)
		if err == nil {
			messages = msgs
		}
	}

	var content string
	switch formatIndex {
	case 1: // JSON
		exportObj := map[string]interface{}{
			"session":     session,
			"messages":    messages,
			"exported_at": time.Now().Format(time.RFC3339),
			"disclaimer":  "Niskava Agent provides non-advisory market intelligence research. Not financial advice.",
		}
		data, err := json.MarshalIndent(exportObj, "", "  ")
		if err != nil {
			return "", err
		}
		content = string(data)

	case 2: // Plain Text
		var b strings.Builder
		b.WriteString(fmt.Sprintf("NISKAVA RESEARCH REPORT - SESSION %s\n", session.ID))
		b.WriteString(fmt.Sprintf("Title: %s\n", session.Title))
		b.WriteString(fmt.Sprintf("Date: %s\n\n", session.UpdatedAt))
		for _, m := range messages {
			b.WriteString(fmt.Sprintf("[%s] %s:\n%s\n\n", m.Role, m.CreatedAt, m.Content))
		}
		b.WriteString("\n---------------------------------------------------\n")
		b.WriteString("DISCLAIMER: Operational empirical research only. Not investment advice.\n")
		content = b.String()

	default: // Markdown (.md)
		var b strings.Builder
		b.WriteString(fmt.Sprintf("# Niskava Research Report: %s\n\n", session.Title))
		b.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", session.ID))
		b.WriteString(fmt.Sprintf("- **Date:** %s\n", session.UpdatedAt))
		b.WriteString(fmt.Sprintf("- **Messages:** %d\n\n", session.MessageCount))
		b.WriteString("---\n\n")

		if len(messages) == 0 {
			b.WriteString(fmt.Sprintf("### Summary\n\n%s\n\n", session.LastMessagePreview))
		} else {
			for _, m := range messages {
				if m.Role == "user" {
					b.WriteString(fmt.Sprintf("### 👤 User\n\n%s\n\n", m.Content))
				} else {
					b.WriteString(fmt.Sprintf("### ⚡ Niskava Agent\n\n%s\n\n", m.Content))
				}
			}
		}

		b.WriteString("---\n\n> **Financial Non-Advisory Disclaimer:** Niskava Agent operates under strict POJK/IDX non-advisory guidelines. Research output does not constitute investment advice.\n")
		content = b.String()
	}

	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	return outPath, nil
}

func (m SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredSessions()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := strings.ToLower(msg.String())

		// Mode A: Confirmation Delete Dialog Active
		if m.ConfirmDelete {
			switch k {
			case "y":
				if m.DeleteTarget != nil {
					targetID := m.DeleteTarget.ID
					if m.AppDB != nil {
						_ = m.AppDB.DeleteChatSession(targetID)
					}
					// Remove from local Sessions list
					var updated []db.ChatSession
					for _, s := range m.Sessions {
						if s.ID != targetID {
							updated = append(updated, s)
						}
					}
					m.Sessions = sortSessions(updated)
					m.StatusNotice = TF("session_selector_exported_notice", targetID)
					m.StatusNoticeTime = time.Now()
				}
				m.ConfirmDelete = false
				m.DeleteTarget = nil
				m.Cursor = 0
				return m, nil

			case "n", "esc":
				m.ConfirmDelete = false
				m.DeleteTarget = nil
				return m, nil
			}
			return m, nil
		}

		// Mode B: Interactive Export Modal Dialog Active
		if m.ExportModalActive {
			switch k {
			case "up", "k":
				if m.ExportFormatIndex > 0 {
					m.ExportFormatIndex--
				} else {
					m.ExportFormatIndex = 2
				}
				return m, nil

			case "down", "j":
				if m.ExportFormatIndex < 2 {
					m.ExportFormatIndex++
				} else {
					m.ExportFormatIndex = 0
				}
				return m, nil

			case "enter":
				if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
					target := filtered[m.Cursor]
					outPath, err := ExportSessionTranscript(m.AppDB, &target, m.ExportFormatIndex)
					if err == nil {
						m.StatusNotice = TF("session_selector_exported_notice", outPath)
					} else {
						m.StatusNotice = fmt.Sprintf("❌ Export failed: %v", err)
					}
					m.StatusNoticeTime = time.Now()
				}
				m.ExportModalActive = false
				return m, nil

			case "esc":
				m.ExportModalActive = false
				return m, nil
			}
			return m, nil
		}

		// Mode C: Normal Session Selector Navigation & Hotkeys
		switch k {
		case "esc", "ctrl+c":
			if m.FilterQuery != "" {
				m.FilterQuery = ""
				m.Cursor = 0
				return m, nil
			}
			m.Canceled = true
			return m, tea.Quit

		case "ctrl+d", "delete":
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				m.DeleteTarget = &target
				m.ConfirmDelete = true
				return m, nil
			}

		case "ctrl+p":
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				newPinned := !target.IsPinned
				if m.AppDB != nil {
					_ = m.AppDB.UpdateChatSession(target.ID, nil, &newPinned, nil)
				}
				// Update in local sessions list
				for i, s := range m.Sessions {
					if s.ID == target.ID {
						m.Sessions[i].IsPinned = newPinned
						break
					}
				}
				m.Sessions = sortSessions(m.Sessions)
				if newPinned {
					m.StatusNotice = T("session_selector_pinned_notice")
				} else {
					m.StatusNotice = T("session_selector_unpinned_notice")
				}
				m.StatusNoticeTime = time.Now()
				return m, nil
			}

		case "ctrl+e":
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				m.ExportModalActive = true
				return m, nil
			}

		case "ctrl+y":
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				textToCopy := fmt.Sprintf("[%s] %s\nSummary: %s", target.ID, target.Title, target.LastMessagePreview)
				if err := CopyToClipboard(textToCopy); err == nil {
					m.StatusNotice = T("session_selector_copied_notice")
				} else {
					m.StatusNotice = fmt.Sprintf("📋 Preview copied (clipboard: %v)", err)
				}
				m.StatusNoticeTime = time.Now()
				return m, nil
			}

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

	// Render confirmation delete dialog if active
	if m.ConfirmDelete && m.DeleteTarget != nil {
		confirmStr := TF("session_selector_delete_confirm", m.DeleteTarget.ID)
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(1, 2).
			Render(confirmStr)
		return "\n" + sessionBoxStyle.Render(title+"\n\n"+box) + "\n"
	}

	// Render Export Modal Dialog if active
	if m.ExportModalActive {
		exportTitle := T("session_selector_export_title")
		formats := []string{
			"Markdown Report (.md)       [Standard Format]",
			"Raw JSON Audit Trail (.json) [Programmatic Data]",
			"Plain Text Summary (.txt)   [Clean ASCII Summary]",
		}

		var fLines []string
		fLines = append(fLines, lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(exportTitle))
		fLines = append(fLines, "")

		for i, fmtStr := range formats {
			if i == m.ExportFormatIndex {
				fLines = append(fLines, lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render("  ▸ [•] "+fmtStr))
			} else {
				fLines = append(fLines, lipgloss.NewStyle().Foreground(ColorMuted).Render("    [ ] "+fmtStr))
			}
		}

		fLines = append(fLines, "")
		fLines = append(fLines, lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render("  [Enter] Export Now  •  [↑/↓] Select Format  •  [Esc] Cancel"))

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(1, 2).
			Render(strings.Join(fLines, "\n"))
		return "\n" + sessionBoxStyle.Render(title+"\n\n"+box) + "\n"
	}

	filtered := m.getFilteredSessions()
	totalAll := len(m.Sessions)
	totalFiltered := len(filtered)

	// Render temporary status notice message if active
	if m.StatusNotice != "" && time.Since(m.StatusNoticeTime) <= 4*time.Second {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(m.StatusNotice) + "\n\n")
	}

	// Always render visible search/filter bar box
	searchPlaceholder := T("session_selector_search_placeholder")

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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("session_selector_no_match")) + "\n")
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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(TF("session_selector_showing", windowStart+1, windowEnd, totalFiltered)) + "\n\n")
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
			pinBadge = " 📌"
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
