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

// Lipgloss Color Palette for NISKAVA-HUD (Binance Dark Financial OSINT Aesthetic)
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
	b.WriteString(statusDotStyle.Render("● ● ● ●") + "  " + lipgloss.NewStyle().Foreground(ColorMuted).Render(T("hud_system_online")))
	b.WriteString("\n\n")

	// 4. Motto Tagline
	b.WriteString(hudSubtitleStyle.Render("I think, therefore I process.  —  \"Don't just answer questions. Investigate them.\"") + "\n\n")

	// 5. Upper Constellation Line
	b.WriteString(RenderConstellationLine(85) + "\n\n")

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
			b.WriteString(labelStyle.Render(lblStr) + valueHighlightStyle.Render(valStr) + "\n")
		} else {
			b.WriteString(labelStyle.Render(lblStr) + valueStyle.Render(valStr) + "\n")
		}
	}

	// 7. Lower Constellation Line
	b.WriteString("\n" + RenderConstellationLine(85) + "\n")

	return b.String()
}
