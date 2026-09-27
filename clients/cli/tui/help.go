package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GetFullHelpGuideString builds and returns the formatted instruction manual string.
func GetFullHelpGuideString() string {
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

	b.WriteString(RenderConstellationLine(85))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(" %s\n", headerStyle.Render(T("help_full_title"))))
	b.WriteString(RenderConstellationLine(85))
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render(T("help_sec1_title")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   • %s : %s\n", keyStyle.Render("[Up/Down ↑/↓] or [k/j]"), T("help_sec1_updown")))
	b.WriteString(fmt.Sprintf("   • %s          : %s\n", keyStyle.Render("[Enter]"), T("help_sec1_enter")))
	b.WriteString(fmt.Sprintf("   • %s        : %s\n", keyStyle.Render("Direct Hotkeys"), T("help_sec1_hotkeys")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[W] / [1]"), T("help_sec1_key_w")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[T] / [2]"), T("help_sec1_key_t")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[S] / [3]"), T("help_sec1_key_s")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[H] / [4]"), T("help_sec1_key_h")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[C] / [5]"), T("help_sec1_key_c")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[L] / [6]"), T("help_sec1_key_l")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n", keyStyle.Render("[Q] / [7]"), T("help_sec1_key_q")))
	b.WriteString(fmt.Sprintf("     - %s : %s\n\n", keyStyle.Render("[E] / [8]"), T("help_sec1_key_e")))

	b.WriteString(sectionStyle.Render(T("help_sec2_title")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   • %s:\n     %s\n", descStyle.Render("[W] Web UI Workspace"), T("help_sec2_web_desc")))
	b.WriteString(fmt.Sprintf("   • %s:\n     %s\n", descStyle.Render("[T] Terminal UI (REPL)"), T("help_sec2_term_desc")))
	b.WriteString(fmt.Sprintf("   • %s:\n     %s\n", descStyle.Render("[S] Session History"), T("help_sec2_sessions_desc")))
	b.WriteString(fmt.Sprintf("   • %s:\n     %s\n", descStyle.Render("[C] Health Check"), T("help_sec2_health_desc")))
	b.WriteString(fmt.Sprintf("   • %s:\n\n     %s\n", descStyle.Render("[Q] Quick Setup Wizard"), T("help_sec2_setup_desc")))

	b.WriteString(sectionStyle.Render(T("help_sec3_title")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/help"), T("slash_help_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/chats"), T("slash_chats_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/resume <id>"), T("slash_resume_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/export [md|json]"), T("slash_export_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/fork [title]"), T("slash_fork_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/search <query>"), T("slash_search_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/anomalies"), T("slash_anomalies_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/skills"), T("slash_skills_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/doctor"), T("slash_doctor_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/cache [stats|clean]"), T("slash_cache_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/back"), T("slash_back_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/reset"), T("slash_reset_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/graph"), T("slash_graph_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/clear"), T("slash_clear_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/web"), T("slash_web_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/sessions"), T("slash_sessions_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/health"), T("slash_health_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n", cmdStyle.Render("/lang"), T("slash_lang_desc")))
	b.WriteString(fmt.Sprintf("   • %-20s : %s\n\n", cmdStyle.Render("/exit, quit"), T("slash_exit_desc")))

	b.WriteString(sectionStyle.Render(T("help_sec4_title")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   • %s : Launch interactive REPL directly (aliases: repl, chat).\n", mutedStyle.Render("niskava terminal [flags]")))
	b.WriteString(fmt.Sprintf("   • %s : Run audit & open live REPL pre-focused on ticker.\n", mutedStyle.Render("niskava investigate <TICKER> -i")))
	b.WriteString(fmt.Sprintf("   • %s : Single-line headless 7-stage investigation.\n", mutedStyle.Render("niskava investigate <TICKER> --days 30")))
	b.WriteString(fmt.Sprintf("   • %s : Launch background daemon server.\n", mutedStyle.Render("niskava serve --port 20128")))
	b.WriteString(fmt.Sprintf("   • %s : List recent chat & investigation sessions.\n", mutedStyle.Render("niskava sessions [list]")))
	b.WriteString(fmt.Sprintf("   • %s : Delete session permanently from SQLite.\n", mutedStyle.Render("niskava sessions delete <SESSION_ID>")))
	b.WriteString(fmt.Sprintf("   • %s : Search past chat history across sessions.\n", mutedStyle.Render("niskava sessions search <KEYWORD>")))
	b.WriteString(fmt.Sprintf("   • %s : Export session report to Markdown or JSON.\n", mutedStyle.Render("niskava sessions export <SESSION_ID> --format md")))
	b.WriteString(fmt.Sprintf("   • %s : Launch interactive setup wizard directly.\n", mutedStyle.Render("niskava setup")))
	b.WriteString(fmt.Sprintf("   • %s : Run system health & diagnostics check.\n", mutedStyle.Render("niskava doctor")))
	b.WriteString(fmt.Sprintf("   • %s : Launch Telegram bot worker.\n", mutedStyle.Render("niskava telegram")))
	b.WriteString(fmt.Sprintf("   • %s : Export Knowledge Graph HTML.\n", mutedStyle.Render("niskava graph --open")))
	b.WriteString(fmt.Sprintf("   • %s : Generate shell autocompletion script.\n\n", mutedStyle.Render("niskava completion [bash|zsh|fish|powershell]")))

	b.WriteString(RenderConstellationLine(85))
	b.WriteString("\n")

	return b.String()
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
		if !m.Ready {
			m.Viewport = viewport.New(msg.Width, msg.Height-3)
			m.Viewport.SetContent(GetFullHelpGuideString())
			m.Ready = true
		} else {
			m.Viewport.Width = msg.Width
			m.Viewport.Height = msg.Height - 3
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m HelpViewerModel) View() string {
	if !m.Ready {
		return "\n  Initializing help viewer...\n"
	}
	footer := lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("[↑/↓/k/j/PgUp/PgDn Scroll  •  Esc Return to Menu]")
	return fmt.Sprintf("%s\n\n  %s", m.Viewport.View(), footer)
}

// PrintFullHelpGuide displays the interactive scrollable help guide.
func PrintFullHelpGuide() {
	p := tea.NewProgram(HelpViewerModel{}, tea.WithAltScreen())
	_, _ = p.Run()
}
