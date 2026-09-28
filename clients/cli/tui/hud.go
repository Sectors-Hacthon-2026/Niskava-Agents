// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/charmbracelet/lipgloss"
)

// Lipgloss Color Palette for NISKAVA-HUD (Binance Dark Financial Intelligence Aesthetic)
var (
	// Palette Aliases for HUD
	hudTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	hudSubtitleStyle = lipgloss.NewStyle().
				Italic(true).
				Foreground(ColorMuted)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	valueStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	valueHighlightStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent)

	statusDotStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	matrixDotStyle = lipgloss.NewStyle().
			Foreground(ColorMuted) // Subdued slate node

	constellationLineStyle = lipgloss.NewStyle().
				Foreground(ColorMuted) // Slate divider
)

// GetDatabaseSizeMB returns formatted size of the SQLite database file.
func GetDatabaseSizeMB(dbPath string) string {
	if dbPath == "" {
		dbPath = filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		return "0.0 MB  state.db"
	}
	sizeMB := float64(info.Size()) / (1024 * 1024)
	base := filepath.Base(dbPath)
	return fmt.Sprintf("%.1f MB  %s", sizeMB, base)
}

// GetDatabaseAge returns the age/uptime representation of the database.
func GetDatabaseAge(dbPath string) string {
	if dbPath == "" {
		dbPath = filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		return TF("hud_age_one_day", time.Now().Format("2006-01-02"))
	}
	modTime := info.ModTime()
	days := int(time.Since(modTime).Hours()/24) + 1
	return TF("hud_age_days", days, modTime.Format("2006-01-02"))
}

// RenderConstellationLine generates a horizontal divider with scattered nodes (◆, ●, ●●).
func RenderConstellationLine(width int) string {
	if width < 40 {
		width = 76
	}

	// Create pattern line
	var sb strings.Builder
	nodes := map[int]string{
		12: lipgloss.NewStyle().Foreground(ColorThought).Render("◆"),
		24: lipgloss.NewStyle().Foreground(ColorAccent).Render("◆"),
		30: lipgloss.NewStyle().Foreground(ColorAccent).Render("◆"),
		36: lipgloss.NewStyle().Foreground(ColorAccent).Render("●●"),
		48: lipgloss.NewStyle().Foreground(ColorThought).Render("◆"),
		60: lipgloss.NewStyle().Foreground(ColorAccent).Render("●"),
		68: lipgloss.NewStyle().Foreground(ColorAccent).Render("●"),
		72: lipgloss.NewStyle().Foreground(ColorThought).Render("◆"),
	}

	for i := 0; i < width; i++ {
		if nodeStr, exists := nodes[i]; exists {
			sb.WriteString(nodeStr)
			if strings.Contains(nodeStr, "●●") {
				i++ // skip next char spot for double dot
			}
		} else {
			sb.WriteString(constellationLineStyle.Render("─"))
		}
	}
	return sb.String()
}

// RenderHUDHeader builds the complete light-green NISKAVA-HUD interface header.
func RenderHUDHeader(modelLabel, serverURL, dbPath, sessionID string) string {
	var b strings.Builder

	// 1. Top Matrix Background Dots
	b.WriteString("\n")
	matrixPattern := matrixDotStyle.Render("· · · · · · · · · · · · · · · ") +
		statusDotStyle.Render("☉") +
		matrixDotStyle.Render(" · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·")
	b.WriteString(matrixPattern)
	b.WriteString("\n")

	// 2. Big Block ASCII Art Banner: NISKAVA
	asciiBanner := []string{
		"███╗   ██╗██╗███████╗██╗  ██╗██████╗  ██╗   ██╗██████╗ ",
		"████╗  ██║██║██╔════╝██║ ██╔╝██╔══██╗ ██║   ██║██╔══██╗",
		"██╔██╗ ██║██║███████╗█████═╝ ███████║ ██║   ██║███████║",
		"██║╚██╗██║██║╚════██║██╔═██╗ ██╔══██║ ╚██╗ ██╔╝██╔══██║",
		"██║ ╚████║██║███████║██║  ██╗██║  ██║  ╚████╔╝ ██║  ██║",
		"╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═══╝  ╚═╝  ╚═╝",
	}

	for _, line := range asciiBanner {
		b.WriteString(hudTitleStyle.Render(line))
		b.WriteString("\n")
	}

	// 3. Status Dots Indicator
	b.WriteString("\n")
	b.WriteString(statusDotStyle.Render("● ● ● ●"))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("hud_system_online")))
	b.WriteString("\n\n")

	// 4. Motto Tagline
	b.WriteString(hudSubtitleStyle.Render("I think, therefore I process.  —  \"Don't just answer questions. Investigate them.\""))
	b.WriteString("\n\n")

	// 5. Upper Constellation Line
	b.WriteString(RenderConstellationLine(85))
	b.WriteString("\n\n")

	// 6. Metadata HUD Stats Panel
	if modelLabel == "" {
		modelLabel = "hermes"
	}
	if sessionID == "" {
		sessionID = "LIVE-SESSION"
	}

	dbSize := GetDatabaseSizeMB(dbPath)
	dbAge := GetDatabaseAge(dbPath)

	specs := []struct {
		Label string
		Val   string
		IsHL  bool
	}{
		{Label: T("hud_lbl_designation"), Val: "NISKAVA (" + sessionID + ")", IsHL: false},
		{Label: T("hud_lbl_substrate"), Val: "sectors-v2 / " + modelLabel, IsHL: false},
		{Label: T("hud_lbl_runtime"), Val: T("hud_runtime_val"), IsHL: false},
		{Label: T("hud_lbl_language"), Val: T("hud_language_val"), IsHL: true},
		{Label: T("hud_lbl_conscious"), Val: dbAge, IsHL: false},
		{Label: T("hud_lbl_brain_size"), Val: dbSize, IsHL: false},
		{Label: T("hud_lbl_interfaces"), Val: "cli, web-workspace (" + serverURL + ")", IsHL: false},
		{Label: T("hud_lbl_purpose"), Val: T("hud_purpose_val"), IsHL: false},
	}

	for _, spec := range specs {
		lblStr := fmt.Sprintf("  %-14s", spec.Label)
		valStr := spec.Val
		if spec.IsHL {
			b.WriteString(labelStyle.Render(lblStr))
			b.WriteString(valueHighlightStyle.Render(valStr))
			b.WriteString("\n")
		} else {
			b.WriteString(labelStyle.Render(lblStr))
			b.WriteString(valueStyle.Render(valStr))
			b.WriteString("\n")
		}
	}

	// 7. Lower Constellation Line
	b.WriteString("\n")
	b.WriteString(RenderConstellationLine(85))
	b.WriteString("\n")

	return b.String()
}

// PrintHealthDiagnostics renders the system health check screen with the Binance Dark Financial Intelligence palette and HUD ASCII header.
func PrintHealthDiagnostics(cfg *config.Config, serverURL string) {
	if cfg == nil {
		return
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	lblStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	valStyle := lipgloss.NewStyle().
		Foreground(ColorFg)

	statusAliveStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess)

	statusErrStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorDanger)

	dividerStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)

	fmt.Println()
	fmt.Println(headerStyle.Render(strings.TrimSpace(T("health_header"))))
	fmt.Println(dividerStyle.Render("─────────────────────────────────────────────────────────────────────────────"))

	daemonURL := serverURL
	if daemonURL == "" {
		daemonURL = "http://localhost:8080"
	}
	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_daemon_url")),
		valStyle.Render(daemonURL),
		statusAliveStyle.Render("[ALIVE]"))

	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_db_path")),
		valStyle.Render(cfg.Storage.DBPath))

	pyBin := cfg.Engine.PythonBin
	if pyBin == "" {
		pyBin = "python3"
	}
	pyStatus := statusAliveStyle.Render("[READY]")
	if _, err := os.Stat(pyBin); err != nil {
		if _, lookErr := exec.LookPath(pyBin); lookErr != nil {
			pyStatus = statusErrStyle.Render("[NOT FOUND]")
		}
	}
	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_python_bin")),
		valStyle.Render(pyBin),
		pyStatus)

	secKeyText := statusAliveStyle.Render(T("health_installed"))
	if cfg.Auth.SectorsAPIKey == "" {
		if cfg.Preferences.OfflineMode {
			secKeyText = statusAliveStyle.Render("[MOCK MODE (OFFLINE)]")
		} else {
			secKeyText = statusErrStyle.Render(T("health_not_installed"))
		}
	}
	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_sectors_key")),
		secKeyText)

	prov := cfg.Auth.AIProvider
	if prov == "" {
		prov = "universal"
	}
	activeModel := cfg.Auth.OpenAIModel
	if strings.ToLower(prov) == "gemini" && cfg.Auth.GeminiModel != "" {
		activeModel = cfg.Auth.GeminiModel
	} else if activeModel == "" {
		if cfg.Auth.GeminiModel != "" {
			activeModel = cfg.Auth.GeminiModel
		} else {
			activeModel = "deepseek/deepseek-chat"
		}
	}

	baseURL := cfg.Auth.OpenAIBaseURL
	if baseURL == "" {
		if strings.ToLower(prov) == "gemini" {
			baseURL = "https://generativelanguage.googleapis.com"
		} else {
			baseURL = "Universal ReAct Standard"
		}
	}
	hasModelKey := cfg.Auth.OpenAIAPIKey != "" || cfg.Auth.GeminiAPIKey != "" || strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1")
	modelKeyText := statusAliveStyle.Render(T("health_installed"))
	if !hasModelKey {
		modelKeyText = statusErrStyle.Render(T("health_not_installed"))
	}

	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_engine")),
		valStyle.Render(fmt.Sprintf("%s (%s)", strings.ToUpper(prov), baseURL)),
		statusAliveStyle.Render("[CONFIGURED]"))

	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model")),
		lblStyle.Render(activeModel))

	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model_key")),
		modelKeyText)

	fmt.Println(dividerStyle.Render("─────────────────────────────────────────────────────────────────────────────"))
}

// PrintWebWorkspaceLaunchScreen renders a styled, rich Web Workspace launcher card using the Binance Dark Financial Intelligence palette.
func PrintWebWorkspaceLaunchScreen(serverURL string) {
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	lblStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	valStyle := lipgloss.NewStyle().
		Foreground(ColorFg)

	urlStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorThought)

	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess)

	mutedStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(ColorMuted)

	var b strings.Builder
	b.WriteString(headerStyle.Render(T("web_launch_title")))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("• %s : %s %s\n", lblStyle.Render(T("web_launch_lbl_status")), statusStyle.Render("[ONLINE]"), mutedStyle.Render("(Go SSE Gateway + React SPA)")))
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(T("web_launch_lbl_url")), urlStyle.Render(serverURL)))
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(T("web_launch_lbl_features")), valStyle.Render("TradingView Anomaly Markers, ReAct SSE Stream, Evidence Matrix")))
	b.WriteString(fmt.Sprintf("• %s : %s\n\n", lblStyle.Render(T("web_launch_lbl_sovereignty")), valStyle.Render("100% Local-First SQLite Persistence (~/.niskava/niskava.db)")))
	b.WriteString(mutedStyle.Render(T("web_launch_opening")))

	fmt.Println()
	fmt.Println(cardStyle.Render(b.String()))
}
