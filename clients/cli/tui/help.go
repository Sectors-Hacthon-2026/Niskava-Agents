package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// PrintFullHelpGuide renders an interactive instruction manual for all surfaces, options, and commands.
func PrintFullHelpGuide() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF87")).
		Background(lipgloss.Color("#052E16")).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ADE80"))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FACC15"))

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#38BDF8"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8FAFC"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	fmt.Println()
	fmt.Println(RenderConstellationLine(85))
	fmt.Printf(" %s\n", headerStyle.Render(T("help_full_title")))
	fmt.Println(RenderConstellationLine(85))

	fmt.Println()
	fmt.Println(sectionStyle.Render(T("help_sec1_title")))
	fmt.Printf("   • %s : %s\n", keyStyle.Render("[Up/Down ↑/↓] or [k/j]"), T("help_sec1_updown"))
	fmt.Printf("   • %s          : %s\n", keyStyle.Render("[Enter]"), T("help_sec1_enter"))
	fmt.Printf("   • %s        : %s\n", keyStyle.Render("Direct Hotkeys"), T("help_sec1_hotkeys"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[W] / [1]"), T("help_sec1_key_w"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[T] / [2]"), T("help_sec1_key_t"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[S] / [3]"), T("help_sec1_key_s"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[H] / [4]"), T("help_sec1_key_h"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[C] / [5]"), T("help_sec1_key_c"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[L] / [6]"), T("help_sec1_key_l"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[Q] / [7]"), T("help_sec1_key_q"))
	fmt.Printf("     - %s : %s\n", keyStyle.Render("[E] / [8]"), T("help_sec1_key_e"))

	fmt.Println()
	fmt.Println(sectionStyle.Render(T("help_sec2_title")))
	fmt.Printf("   • %s:\n     %s\n", descStyle.Render("[W] Web UI Workspace"), T("help_sec2_web_desc"))
	fmt.Printf("   • %s:\n     %s\n", descStyle.Render("[T] Terminal UI (REPL)"), T("help_sec2_term_desc"))
	fmt.Printf("   • %s:\n     %s\n", descStyle.Render("[S] Session History"), T("help_sec2_sessions_desc"))
	fmt.Printf("   • %s:\n     %s\n", descStyle.Render("[C] Health Check"), T("help_sec2_health_desc"))
	fmt.Printf("   • %s:\n     %s\n", descStyle.Render("[Q] Quick Setup Wizard"), T("help_sec2_setup_desc"))

	fmt.Println()
	fmt.Println(sectionStyle.Render(T("help_sec3_title")))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/help"), T("slash_help_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/reset"), T("slash_reset_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/graph"), T("slash_graph_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/clear"), T("slash_clear_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/web"), T("slash_web_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/sessions"), T("slash_sessions_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/health"), T("slash_health_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/lang"), T("slash_lang_desc"))
	fmt.Printf("   • %-16s : %s\n", cmdStyle.Render("/exit, quit"), T("slash_exit_desc"))

	fmt.Println()
	fmt.Println(sectionStyle.Render(T("help_sec4_title")))
	fmt.Printf("   • %s : Single-line headless investigation.\n", mutedStyle.Render("niskava investigate <TICKER> --days 30"))
	fmt.Printf("   • %s : Launch background daemon server.\n", mutedStyle.Render("niskava serve --port 20128"))
	fmt.Printf("   • %s : Display session history from CLI.\n", mutedStyle.Render("niskava sessions"))
	fmt.Printf("   • %s : Launch interactive setup wizard directly.\n", mutedStyle.Render("niskava setup"))

	fmt.Println()
	fmt.Println(RenderConstellationLine(85))
}
