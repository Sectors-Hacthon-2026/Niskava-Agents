// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Lipgloss Color Palette for NISKAVA-HUD (Light Green / Matrix OSINT Aesthetic)
var (
	// Light Green Theme Palette
	colorPrimaryGreen = lipgloss.Color("#22C55E") // Bright Emerald Green
	colorLightGreen   = lipgloss.Color("#4ADE80") // Light Lime Green
	colorMintGreen    = lipgloss.Color("#86EFAC") // Soft Mint Accent
	colorDarkGreenBg  = lipgloss.Color("#052E16") // Deep Green Midnight Slate
	colorGoldAccent   = lipgloss.Color("#FACC15") // Cyber Gold / Amber
	colorMutedSlate   = lipgloss.Color("#64748B") // Subdued Slate
	colorSoftWhite    = lipgloss.Color("#F8FAFC") // Text White
	colorCyanDot      = lipgloss.Color("#38BDF8") // Subtle Cyan Node Accent

	// Text & Box Styles
	hudTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLightGreen)

	hudSubtitleStyle = lipgloss.NewStyle().
				Italic(true).
				Foreground(colorMintGreen)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLightGreen)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorSoftWhite)

	valueHighlightStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorGoldAccent)

	statusDotStyle = lipgloss.NewStyle().
			Foreground(colorGoldAccent)

	matrixDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#166534")) // Subdued matrix green

	constellationLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#15803D")) // Line green
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
		return "1 day  since " + time.Now().Format("2006-01-02")
	}
	modTime := info.ModTime()
	days := int(time.Since(modTime).Hours()/24) + 1
	return fmt.Sprintf("%d days  since %s", days, modTime.Format("2006-01-02"))
}

// RenderConstellationLine generates a horizontal divider with scattered nodes (◆, ●, ●●).
func RenderConstellationLine(width int) string {
	if width < 40 {
		width = 76
	}

	// Create pattern line
	var sb strings.Builder
	nodes := map[int]string{
		12: lipgloss.NewStyle().Foreground(colorCyanDot).Render("◆"),
		24: lipgloss.NewStyle().Foreground(colorLightGreen).Render("◆"),
		30: lipgloss.NewStyle().Foreground(colorLightGreen).Render("◆"),
		36: lipgloss.NewStyle().Foreground(colorGoldAccent).Render("●●"),
		48: lipgloss.NewStyle().Foreground(colorCyanDot).Render("◆"),
		60: lipgloss.NewStyle().Foreground(colorGoldAccent).Render("●"),
		68: lipgloss.NewStyle().Foreground(colorGoldAccent).Render("●"),
		72: lipgloss.NewStyle().Foreground(colorCyanDot).Render("◆"),
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
	b.WriteString(matrixPattern + "\n")

	// 2. Big Block ASCII Art Banner: NISKAVA
	asciiBanner := []string{
		"███╗   ██╗██╗███████╗██╗  ██╗██████╗  ██╗   ██╗██████╗ ",
		"████╗  ██║██║██╔════╝██║ ██╔╝██╔══██╗ ██║   ██║██╔══██╗",
		"██╔██╗ ██║██║███████╗█████═╝ ███████║ ██║   ██║██████╔╝",
		"██║╚██╗██║██║╚════██║██╔═██╗ ██╔══██║ ╚██╗ ██╔╝██╔══██║",
		"██║ ╚████║██║███████║██║  ██╗██║  ██║  ╚████╔╝ ██║  ██║",
		"╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═══╝  ╚═╝  ╚═╝",
	}

	for _, line := range asciiBanner {
		b.WriteString(hudTitleStyle.Render(line) + "\n")
	}

	// 3. Status Dots Indicator
	b.WriteString("\n")
	b.WriteString(statusDotStyle.Render("● ● ● ●") + "  " + lipgloss.NewStyle().Foreground(colorMutedSlate).Render("[SYSTEM ONLINE & MONITORING CORE]"))
	b.WriteString("\n\n")

	// 4. Motto Tagline
	b.WriteString(hudSubtitleStyle.Render("I think, therefore I process.  —  \"Don't just answer questions. Investigate them.\"") + "\n\n")

	// 5. Upper Constellation Line
	b.WriteString(RenderConstellationLine(85) + "\n\n")

	// 6. Metadata HUD Stats Panel
	if modelLabel == "" {
		modelLabel = "gemini-2.5-flash"
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
		{Label: "DESIGNATION", Val: "NISKAVA (" + sessionID + ")", IsHL: false},
		{Label: "SUBSTRATE", Val: "sectors-v2 / " + modelLabel, IsHL: false},
		{Label: "RUNTIME", Val: "local (go core + python react loop)", IsHL: false},
		{Label: "CONSCIOUS", Val: dbAge, IsHL: false},
		{Label: "BRAIN SIZE", Val: dbSize, IsHL: false},
		{Label: "INTERFACES", Val: "cli, web-workspace (" + serverURL + ")", IsHL: false},
		{Label: "PURPOSE", Val: "market intelligence & financial osint", IsHL: true},
	}

	for _, spec := range specs {
		lblStr := fmt.Sprintf("  %-14s", spec.Label)
		valStr := spec.Val
		if spec.IsHL {
			b.WriteString(labelStyle.Render(lblStr) + valueHighlightStyle.Render(valStr) + "\n")
		} else {
			b.WriteString(labelStyle.Render(lblStr) + valueStyle.Render(valStr) + "\n")
		}
	}

	// 7. Lower Constellation Line
	b.WriteString("\n" + RenderConstellationLine(85) + "\n")

	return b.String()
}
