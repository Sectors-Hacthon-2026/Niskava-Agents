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
	fmt.Printf(" %s\n", headerStyle.Render("NISKAVA AGENT — SYSTEM INSTRUCTION & USAGE GUIDE"))
	fmt.Println(RenderConstellationLine(85))

	fmt.Println()
	fmt.Println(sectionStyle.Render("1. MAIN MENU NAVIGATION GUIDE (LAUNCHER):"))
	fmt.Printf("   • %s : Move cursor up/down between menu choices.\n", keyStyle.Render("[Up/Down ↑/↓] or [k/j]"))
	fmt.Printf("   • %s          : Execute selected menu option.\n", keyStyle.Render("[Enter]"))
	fmt.Printf("   • %s        : Press direct shortcut keys for fast execution:\n", keyStyle.Render("Direct Hotkeys"))
	fmt.Printf("     - %s : Open Web Workspace in default browser (REST & SSE Visual Stream)\n", keyStyle.Render("[W] / [1]"))
	fmt.Printf("     - %s : Launch Interactive Live Terminal UI (REPL & Anomaly Reasoning)\n", keyStyle.Render("[T] / [2]"))
	fmt.Printf("     - %s : View Investigation Session History & Audit Trail (SQLite)\n", keyStyle.Render("[S] / [3]"))
	fmt.Printf("     - %s : Display this System Guide & Instruction Manual\n", keyStyle.Render("[H] / [4]"))
	fmt.Printf("     - %s : Check Daemon Server, DB Connection & AI Provider Health\n", keyStyle.Render("[C] / [5]"))
	fmt.Printf("     - %s : Launch Quick Setup Wizard (.env configuration)\n", keyStyle.Render("[Q] / [6]"))
	fmt.Printf("     - %s : Stop daemon server & exit Niskava Agent\n", keyStyle.Render("[E] / [7]"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("2. OPERATIONAL SURFACES & FEATURE INSTRUCTIONS:"))
	fmt.Printf("   • %s:\n     Visual research dashboard based on React SPA in browser. Displays TradingView/Recharts\n     candlesticks, quantitative anomaly markers (MA20/Z-score), 3-tier evidence matrix,\n     and local memory graph visualizer.\n", descStyle.Render("[W] Web UI Workspace"))
	fmt.Printf("   • %s:\n     Interactive terminal ReAct inner-monologue.\n     - Type a 4-5 letter stock ticker (e.g. ANTM, BBCA, BUMI) for auto 30-day investigation.\n     - Ask free-form market research questions (e.g. 'Why did ANTM surge sharply yesterday?').\n", descStyle.Render("[T] Terminal UI (REPL)"))
	fmt.Printf("   • %s:\n     Reads local SQLite tables (~/.niskava/niskava.db) to inspect past audit trails & evidence.\n", descStyle.Render("[S] Session History"))
	fmt.Printf("   • %s:\n     Displays status of background daemon, Sectors v2 API key, and LLM providers (Gemini/OpenAI).\n", descStyle.Render("[C] Health Check"))
	fmt.Printf("   • %s:\n     Interactive setup wizard to automatically configure your .env file.\n", descStyle.Render("[Q] Quick Setup Wizard"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("3. SLASH COMMANDS IN REPL SESSIONS:"))
	fmt.Printf("   • %-16s : Show help command overlay.\n", cmdStyle.Render("/help"))
	fmt.Printf("   • %-16s : Start new conversation session & clear memory graph.\n", cmdStyle.Render("/reset"))
	fmt.Printf("   • %-16s : Clear terminal screen & redraw HUD banner.\n", cmdStyle.Render("/clear"))
	fmt.Printf("   • %-16s : Open Web Workspace directly in browser.\n", cmdStyle.Render("/web"))
	fmt.Printf("   • %-16s : Inspect saved investigation history from SQLite.\n", cmdStyle.Render("/sessions"))
	fmt.Printf("   • %-16s : Check health status of AI providers & database.\n", cmdStyle.Render("/health"))
	fmt.Printf("   • %-16s : Exit REPL session back to main menu.\n", cmdStyle.Render("/exit, quit"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("4. DIRECT COMMAND LINE CLI COMMANDS:"))
	fmt.Printf("   • %s : Single-line headless investigation.\n", mutedStyle.Render("niskava investigate <TICKER> --days 30"))
	fmt.Printf("   • %s : Launch background daemon server.\n", mutedStyle.Render("niskava serve --port 8080"))
	fmt.Printf("   • %s : Display session history from CLI.\n", mutedStyle.Render("niskava sessions"))
	fmt.Printf("   • %s : Launch interactive setup wizard directly.\n", mutedStyle.Render("niskava setup"))

	fmt.Println()
	fmt.Println(RenderConstellationLine(85))
}
