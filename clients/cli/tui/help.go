package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// padRight pads a string with spaces until its visual display width (measured by lipgloss.Width) reaches targetWidth.
func padRight(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// GetFullHelpGuideString builds and returns the formatted instruction manual string.
func GetFullHelpGuideString(overrideWidth ...int) string {
	w := GetTermWidth()
	if len(overrideWidth) > 0 && overrideWidth[0] > 0 {
		w = overrideWidth[0]
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorThought)

	descStyle := lipgloss.NewStyle().
		Foreground(ColorFg)

	mutedStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)

	var b strings.Builder

	// Header Title Banner
	titleText := T("help_full_title")
	b.WriteString(RenderConstellationLine(w))
	b.WriteString("\n")
	if lipgloss.Width(titleText)+4 > w && w >= 30 {
		wrappedTitleStyle := headerStyle.Width(w - 4).Align(lipgloss.Center)
		b.WriteString(fmt.Sprintf(" %s\n", wrappedTitleStyle.Render(titleText)))
	} else {
		b.WriteString(fmt.Sprintf(" %s\n", headerStyle.Render(titleText)))
	}
	b.WriteString(RenderConstellationLine(w))
	b.WriteString("\n\n")

	// Section 1: Main Menu Navigation Guide
	b.WriteString(sectionStyle.Render(T("help_sec1_title")))
	b.WriteString("\n")

	sec1Items := []struct {
		Key  string
		Desc string
	}{
		{"[Up/Down ↑/↓] or [k/j]", T("help_sec1_updown")},
		{"[Enter]", T("help_sec1_enter")},
		{"Direct Hotkeys", T("help_sec1_hotkeys")},
	}

	keyColWidth := 22
	for _, item := range sec1Items {
		paddedKey := padRight(item.Key, keyColWidth)
		prefix := fmt.Sprintf("   • %s : ", keyStyle.Render(paddedKey))
		b.WriteString(formatHelpItem(prefix, item.Desc, w-2))
	}

	hotkeys := []struct {
		Key  string
		Desc string
	}{
		{"[W] / [1]", T("help_sec1_key_w")},
		{"[T] / [2]", T("help_sec1_key_t")},
		{"[S] / [3]", T("help_sec1_key_s")},
		{"[H] / [4]", T("help_sec1_key_h")},
		{"[C] / [5]", T("help_sec1_key_c")},
		{"[L] / [6]", T("help_sec1_key_l")},
		{"[Q] / [7]", T("help_sec1_key_q")},
		{"[E] / [8]", T("help_sec1_key_e")},
	}

	hkColWidth := 11
	for _, hk := range hotkeys {
		paddedHk := padRight(hk.Key, hkColWidth)
		prefix := fmt.Sprintf("     - %s : ", keyStyle.Render(paddedHk))
		b.WriteString(formatHelpItem(prefix, hk.Desc, w-2))
	}
	b.WriteString("\n")

	// Section 2: Operational Surfaces & Feature Instructions
	b.WriteString(sectionStyle.Render(T("help_sec2_title")))
	b.WriteString("\n")

	sec2Items := []struct {
		Title string
		Desc  string
	}{
		{"[W] Web UI Workspace", T("help_sec2_web_desc")},
		{"[T] Terminal UI (REPL)", T("help_sec2_term_desc")},
		{"[S] Session History", T("help_sec2_sessions_desc")},
		{"[C] Health Check", T("help_sec2_health_desc")},
		{"[Q] Quick Setup Wizard", T("help_sec2_setup_desc")},
	}

	descWidth := w - 7
	if descWidth < 15 {
		descWidth = 15
	}

	for _, item := range sec2Items {
		b.WriteString(fmt.Sprintf("   • %s:\n", descStyle.Render(item.Title)))
		paras := strings.Split(item.Desc, "\n")
		for _, para := range paras {
			trimmed := strings.TrimSpace(para)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "- ") {
				b.WriteString(formatHelpItem("     - ", strings.TrimPrefix(trimmed, "- "), w-2))
			} else {
				wrapped := wrapText(trimmed, descWidth)
				for _, line := range strings.Split(wrapped, "\n") {
					b.WriteString("     " + line + "\n")
				}
			}
		}
		b.WriteString("\n")
	}

	// Section 3: Slash Commands
	b.WriteString(sectionStyle.Render(T("help_sec3_title")))
	b.WriteString("\n")

	slashCmds := []struct {
		Cmd  string
		Desc string
	}{
		{"/help", T("slash_help_desc")},
		{"/chats", T("slash_chats_desc")},
		{"/compact", T("slash_compact_desc")},
		{"/find <kw>", T("slash_find_desc")},
		{"/copy", T("slash_copy_desc")},
		{"/resume <id>", T("slash_resume_desc")},
		{"/export [md|json]", T("slash_export_desc")},
		{"/fork [title]", T("slash_fork_desc")},
		{"/search <query>", T("slash_search_desc")},
		{"/anomalies", T("slash_anomalies_desc")},
		{"/skills", T("slash_skills_desc")},
		{"/doctor", T("slash_doctor_desc")},
		{"/cache [stats|clean]", T("slash_cache_desc")},
		{"/back", T("slash_back_desc")},
		{"/reset", T("slash_reset_desc")},
		{"/graph", T("slash_graph_desc")},
		{"/clear", T("slash_clear_desc")},
		{"/web", T("slash_web_desc")},
		{"/sessions", T("slash_sessions_desc")},
		{"/health", T("slash_health_desc")},
		{"/lang [en|id]", T("slash_lang_desc")},
		{"/timeout [arg]", T("slash_timeout_desc")},
		{"/exit, quit", T("slash_exit_desc")},
	}

	if w >= 85 {
		scColWidth := 21
		for _, sc := range slashCmds {
			paddedCmd := padRight(sc.Cmd, scColWidth)
			prefix := fmt.Sprintf("   • %s : ", cmdStyle.Render(paddedCmd))
			b.WriteString(formatHelpItem(prefix, sc.Desc, w-2))
		}
	} else {
		for _, sc := range slashCmds {
			b.WriteString(fmt.Sprintf("   • %s:\n", cmdStyle.Render(sc.Cmd)))
			wrapped := wrapText(sc.Desc, descWidth)
			for _, line := range strings.Split(wrapped, "\n") {
				b.WriteString("     " + line + "\n")
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Section 4: Direct CLI Commands
	b.WriteString(sectionStyle.Render(T("help_sec4_title")))
	b.WriteString("\n")

	cliCmds := []struct {
		Cmd  string
		Desc string
	}{
		{"niskava terminal [flags]", "Launch interactive REPL directly (aliases: repl, chat)."},
		{"niskava investigate <TICKER> -i", "Run audit & open live REPL pre-focused on ticker."},
		{"niskava investigate <TICKER> --days 30", "Single-line headless 7-stage investigation."},
		{"niskava serve --port 20128", "Launch background daemon server."},
		{"niskava sessions [list]", "List recent chat & investigation sessions."},
		{"niskava sessions delete <SESSION_ID>", "Delete session permanently from SQLite."},
		{"niskava sessions search <KEYWORD>", "Search past chat history across sessions."},
		{"niskava sessions export <ID> --format md", "Export session report to Markdown or JSON."},
		{"niskava setup", "Launch interactive setup wizard directly."},
		{"niskava doctor", "Run system health & diagnostics check."},
		{"niskava telegram", "Launch Telegram bot worker."},
		{"niskava graph --open", "Export Knowledge Graph HTML."},
		{"niskava completion [shell]", "Generate shell autocompletion script."},
	}

	for _, item := range cliCmds {
		b.WriteString(fmt.Sprintf("   • %s:\n", mutedStyle.Render(item.Cmd)))
		wrapped := wrapText(item.Desc, descWidth)
		for _, line := range strings.Split(wrapped, "\n") {
			b.WriteString("     " + line + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString(RenderConstellationLine(w))
	b.WriteString("\n")

	return b.String()
}

// formatHelpItem formats a key-value or bullet item so that the description text
// wraps cleanly to subsequent lines, indented under the description column.
func formatHelpItem(prefix, desc string, totalWidth int) string {
	if totalWidth < 30 {
		totalWidth = 30
	}

	prefixLen := lipgloss.Width(prefix)
	avail := totalWidth - prefixLen

	if avail < 15 {
		indent := "     "
		descWidth := totalWidth - 5
		if descWidth < 15 {
			descWidth = 15
		}
		wrappedDesc := wrapText(desc, descWidth)
		descLines := strings.Split(wrappedDesc, "\n")
		var sb strings.Builder
		sb.WriteString(prefix)
		sb.WriteString("\n")
		for _, line := range descLines {
			sb.WriteString(indent)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		return sb.String()
	}

	wrappedDesc := wrapText(desc, avail)
	descLines := strings.Split(wrappedDesc, "\n")
	indent := strings.Repeat(" ", prefixLen)

	var sb strings.Builder
	for i, line := range descLines {
		if i == 0 {
			sb.WriteString(prefix)
			sb.WriteString(line)
			sb.WriteString("\n")
		} else {
			sb.WriteString(indent)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// wrapText breaks a string into lines that do not exceed width characters.
func wrapText(s string, width int) string {
	if width <= 0 || len(s) == 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if lipgloss.Width(currentLine+" "+word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return strings.Join(lines, "\n")
}

// HelpViewerModel is a Bubbletea model for interactive scrollable help viewing.
type HelpViewerModel struct {
	Viewport viewport.Model
	Ready    bool
}

func (m HelpViewerModel) Init() tea.Cmd {
	return nil
}

func (m HelpViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		w := msg.Width
		h := msg.Height - 3
		if !m.Ready {
			m.Viewport = viewport.New(w, h)
			m.Ready = true
		} else {
			m.Viewport.Width = w
			m.Viewport.Height = h
		}
		m.Viewport.SetContent(GetFullHelpGuideString(w))
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m HelpViewerModel) View() string {
	if !m.Ready {
		return "\n  Initializing help viewer...\n\033[J"
	}
	w := m.Viewport.Width
	if w <= 0 {
		w = GetTermWidth()
	}
	footerText := "[↑/↓/k/j/PgUp/PgDn Scroll  •  Esc Return to Menu]"
	if w < 55 {
		footerText = "[↑/↓ Scroll  •  Esc Return to Menu]"
	}
	if w < 40 {
		footerText = "[↑/↓ Scroll • Esc Menu]"
	}
	footer := lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(footerText)
	return fmt.Sprintf("%s\n\n  %s", m.Viewport.View(), footer) + "\033[J"
}

// PrintFullHelpGuide displays the interactive scrollable help guide.
func PrintFullHelpGuide() {
	p := tea.NewProgram(HelpViewerModel{}, tea.WithAltScreen())
	_, _ = p.Run()
}
