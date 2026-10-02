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
	tea "github.com/charmbracelet/bubbletea"
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
func RenderConstellationLine(overrideWidth ...int) string {
	w := GetTermWidth()
	if len(overrideWidth) > 0 && overrideWidth[0] > 0 {
		w = overrideWidth[0]
	}
	if w < 15 {
		w = 15
	}

	n15 := int(float64(w) * 0.15)
	n30 := int(float64(w) * 0.30)
	n50 := int(float64(w) * 0.50)
	n75 := int(float64(w) * 0.75)
	n90 := int(float64(w) * 0.90)

	nodes := map[int]string{
		n15: lipgloss.NewStyle().Foreground(ColorThought).Render("◆"),
		n30: lipgloss.NewStyle().Foreground(ColorAccent).Render("◆"),
		n50: lipgloss.NewStyle().Foreground(ColorAccent).Render("●●"),
		n75: lipgloss.NewStyle().Foreground(ColorThought).Render("◆"),
		n90: lipgloss.NewStyle().Foreground(ColorAccent).Render("●"),
	}

	var sb strings.Builder
	for i := 0; i < w; i++ {
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
	w := GetTermWidth()

	// 1. Top Matrix Background Dots
	dotW := w - 4
	if dotW < 10 {
		dotW = 10
	}
	b.WriteString("\n")
	dotCount := (dotW - 4) / 2
	if dotCount < 5 {
		dotCount = 5
	}
	leftDots := dotCount / 3
	rightDots := dotCount - leftDots - 1

	matrixPattern := matrixDotStyle.Render(strings.Repeat("·", leftDots)) +
		statusDotStyle.Render("☉") +
		matrixDotStyle.Render(" "+strings.Repeat("· ", rightDots))
	b.WriteString(matrixPattern)
	b.WriteString("\n")

	// 2. Responsive ASCII Art Banner
	b.WriteString(RenderResponsiveASCIIHeader(w, hudTitleStyle))

	// 3. Status Dots Indicator
	b.WriteString("\n")
	b.WriteString(statusDotStyle.Render("● ● ● ●"))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("hud_system_online")))
	b.WriteString("\n\n")

	// 4. Motto Tagline
	tagline := "I think, therefore I process. — \"Don't just answer questions. Investigate them.\""
	if w < 75 {
		tagline = "\"Don't just answer questions. Investigate them.\""
	}
	if w < 52 {
		tagline = "\"Don't just answer questions.\nInvestigate them.\""
	}
	b.WriteString(hudSubtitleStyle.Render(tagline))
	b.WriteString("\n\n")

	// 5. Upper Constellation Line
	divW := w - 2
	if divW < 15 {
		divW = 15
	}
	b.WriteString(RenderConstellationLine(divW))
	b.WriteString("\n\n")

	// 6. Metadata HUD Stats Panel
	if modelLabel == "" {
		modelLabel = "niskava"
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
		{Label: T("hud_lbl_interfaces"), Val: "cli, web (" + serverURL + ")", IsHL: false},
		{Label: T("hud_lbl_purpose"), Val: T("hud_purpose_val"), IsHL: false},
	}

	lblWidth := 14
	if w < 50 {
		lblWidth = 11
	}
	maxValWidth := w - (lblWidth + 4)
	if maxValWidth < 10 {
		maxValWidth = 10
	}

	for _, spec := range specs {
		lblStr := fmt.Sprintf("  %-*s", lblWidth, Truncate(spec.Label, lblWidth))
		valStr := Truncate(spec.Val, maxValWidth)
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
	b.WriteString(RenderConstellationLine(divW))
	b.WriteString("\n")

	return b.String()
}

// PrintHealthDiagnostics renders the system health check screen with the Binance Dark Financial Intelligence palette and HUD ASCII header.
func PrintHealthDiagnostics(cfg *config.Config, serverURL string) {
	if cfg == nil {
		return
	}

	w := GetTermWidth()

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

	fmt.Println()
	fmt.Println(headerStyle.Render(strings.TrimSpace(T("health_header"))))
	fmt.Println(Sep(2, w))

	daemonURL := serverURL
	if daemonURL == "" {
		daemonURL = "http://localhost:8080"
	}
	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_daemon_url")),
		valStyle.Render(daemonURL),
		AtomicBadge("[ALIVE]", statusAliveStyle))

	dbPathDisp := TruncateMiddle(cfg.Storage.DBPath, w-24)
	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_db_path")),
		valStyle.Render(dbPathDisp))

	pyBin := cfg.Engine.PythonBin
	if pyBin == "" {
		pyBin = "python3"
	}
	pyStatus := AtomicBadge("[READY]", statusAliveStyle)
	if _, err := os.Stat(pyBin); err != nil {
		if _, lookErr := exec.LookPath(pyBin); lookErr != nil {
			pyStatus = AtomicBadge("[NOT FOUND]", statusErrStyle)
		}
	}
	pyBinDisp := TruncateMiddle(pyBin, w-24)
	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_python_bin")),
		valStyle.Render(pyBinDisp),
		pyStatus)

	if strings.Contains(pyBin, "WindowsApps") {
		hintStyle := lipgloss.NewStyle().Foreground(ColorWarning).Italic(true)
		fmt.Println(hintStyle.Render("  💡 Hint: WindowsApps is a Microsoft Store stub. If engine fails, install Python directly from python.org."))
	}

	secKeyText := AtomicBadge(T("health_installed"), statusAliveStyle)
	if cfg.Auth.SectorsAPIKey == "" {
		secKeyText = AtomicBadge(T("health_not_installed"), statusErrStyle)
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
	modelKeyText := AtomicBadge(T("health_installed"), statusAliveStyle)
	if !hasModelKey {
		modelKeyText = AtomicBadge(T("health_not_installed"), statusErrStyle)
	}

	fmt.Printf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_engine")),
		valStyle.Render(fmt.Sprintf("%s (%s)", strings.ToUpper(prov), Truncate(baseURL, 25))),
		AtomicBadge("[CONFIGURED]", statusAliveStyle))

	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model")),
		lblStyle.Render(activeModel))

	fmt.Printf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model_key")),
		modelKeyText)

	fmt.Println(Sep(2, w))
}

// HealthViewerModel is an interactive Bubbletea AltScreen model for system health diagnostics.
type HealthViewerModel struct {
	CFG       *config.Config
	ServerURL string
	Width     int
	Height    int
}

func (m HealthViewerModel) Init() tea.Cmd {
	return nil
}

func (m HealthViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "q", "space", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m HealthViewerModel) View() string {
	w := m.Width
	if w <= 0 {
		w = GetTermWidth()
	}
	boxW := w - 4
	if boxW > w-2 {
		boxW = w - 2
	}
	if boxW < 16 {
		boxW = max(10, w-2)
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

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(1, 2).
		MarginTop(1)

	var b strings.Builder
	b.WriteString(headerStyle.Render(strings.TrimSpace(T("health_header"))))
	b.WriteString("\n\n")

	daemonURL := m.ServerURL
	if daemonURL == "" {
		daemonURL = "http://localhost:8080"
	}
	b.WriteString(fmt.Sprintf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_daemon_url")),
		valStyle.Render(daemonURL),
		AtomicBadge("[ALIVE]", statusAliveStyle)))

	dbPath := m.CFG.Storage.DBPath
	dbPathDisp := TruncateMiddle(dbPath, boxW-24)
	b.WriteString(fmt.Sprintf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_db_path")),
		valStyle.Render(dbPathDisp)))

	pyBin := m.CFG.Engine.PythonBin
	if pyBin == "" {
		pyBin = "python3"
	}
	pyStatus := AtomicBadge("[READY]", statusAliveStyle)
	if _, err := os.Stat(pyBin); err != nil {
		if _, lookErr := exec.LookPath(pyBin); lookErr != nil {
			pyStatus = AtomicBadge("[NOT FOUND]", statusErrStyle)
		}
	}
	pyBinDisp := TruncateMiddle(pyBin, boxW-24)
	b.WriteString(fmt.Sprintf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_python_bin")),
		valStyle.Render(pyBinDisp),
		pyStatus))

	secKeyText := AtomicBadge(T("health_installed"), statusAliveStyle)
	if m.CFG.Auth.SectorsAPIKey == "" {
		secKeyText = AtomicBadge(T("health_not_installed"), statusErrStyle)
	}
	b.WriteString(fmt.Sprintf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_sectors_key")),
		secKeyText))

	prov := m.CFG.Auth.AIProvider
	if prov == "" {
		prov = "universal"
	}
	activeModel := m.CFG.Auth.OpenAIModel
	if strings.ToLower(prov) == "gemini" && m.CFG.Auth.GeminiModel != "" {
		activeModel = m.CFG.Auth.GeminiModel
	} else if activeModel == "" {
		if m.CFG.Auth.GeminiModel != "" {
			activeModel = m.CFG.Auth.GeminiModel
		} else {
			activeModel = "deepseek/deepseek-chat"
		}
	}

	baseURL := m.CFG.Auth.OpenAIBaseURL
	if baseURL == "" {
		if strings.ToLower(prov) == "gemini" {
			baseURL = "https://generativelanguage.googleapis.com"
		} else {
			baseURL = "Universal ReAct Standard"
		}
	}
	hasModelKey := m.CFG.Auth.OpenAIAPIKey != "" || m.CFG.Auth.GeminiAPIKey != "" || strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1")
	modelKeyText := AtomicBadge(T("health_installed"), statusAliveStyle)
	if !hasModelKey {
		modelKeyText = AtomicBadge(T("health_not_installed"), statusErrStyle)
	}

	b.WriteString(fmt.Sprintf("• %s: %s %s\n",
		lblStyle.Render(T("health_lbl_engine")),
		valStyle.Render(fmt.Sprintf("%s (%s)", strings.ToUpper(prov), Truncate(baseURL, 25))),
		AtomicBadge("[CONFIGURED]", statusAliveStyle)))

	b.WriteString(fmt.Sprintf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model")),
		lblStyle.Render(activeModel)))

	b.WriteString(fmt.Sprintf("• %s: %s\n",
		lblStyle.Render(T("health_lbl_model_key")),
		modelKeyText))

	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("[Press Esc or Enter to Return to Menu]"))

	return "\n  " + cardStyle.Render(b.String()) + "\n\033[J"
}

// ShowHealthDiagnosticsScreen displays interactive AltScreen health diagnostics card that live-resizes on window resize.
func ShowHealthDiagnosticsScreen(cfg *config.Config, serverURL string) {
	p := tea.NewProgram(HealthViewerModel{CFG: cfg, ServerURL: serverURL}, tea.WithAltScreen())
	_, _ = p.Run()
}

// PrintWebWorkspaceLaunchScreen renders a styled, rich Web Workspace launcher card using the Binance Dark Financial Intelligence palette.
func PrintWebWorkspaceLaunchScreen(serverURL string) {
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	w := GetTermWidth()
	boxW := w - 4
	if boxW < 30 {
		boxW = 30
	}
	contentW := boxW - 6
	if contentW < 18 {
		contentW = 18
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(0, 1).
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

	titleText := T("web_launch_title")
	if lipgloss.Width(titleText) > contentW {
		titleText = "🌐 NISKAVA WEB WORKSPACE"
	}
	if lipgloss.Width(titleText) > contentW {
		titleText = "NISKAVA WEB"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(headerStyle.Render(titleText))
	b.WriteString("\n\n")

	// 1. Status
	statusLabel := T("web_launch_lbl_status")
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(statusLabel), statusStyle.Render("[ONLINE]")))
	if contentW >= 32 {
		b.WriteString(fmt.Sprintf("  %s\n", mutedStyle.Render("(Go SSE Gateway + React SPA)")))
	}

	// 2. URL
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(T("web_launch_lbl_url")), urlStyle.Render(serverURL)))

	// 3. Features
	b.WriteString(fmt.Sprintf("• %s :\n", lblStyle.Render(T("web_launch_lbl_features"))))
	featLines := wrapText("TradingView Anomaly Markers, ReAct SSE Stream, Evidence Matrix", contentW-2)
	for _, fl := range strings.Split(featLines, "\n") {
		b.WriteString(fmt.Sprintf("  %s\n", valStyle.Render(fl)))
	}

	// 4. Sovereignty
	b.WriteString(fmt.Sprintf("• %s :\n", lblStyle.Render(T("web_launch_lbl_sovereignty"))))
	sovLines := wrapText("100% Local-First SQLite Persistence (~/.niskava/niskava.db)", contentW-2)
	for _, sl := range strings.Split(sovLines, "\n") {
		b.WriteString(fmt.Sprintf("  %s\n", valStyle.Render(sl)))
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(Truncate(T("web_launch_opening"), contentW)))

	fmt.Println()
	fmt.Println(cardStyle.Render(b.String()))
}

// WebWorkspaceViewerModel is an interactive Bubbletea AltScreen model for the Web Workspace card.
type WebWorkspaceViewerModel struct {
	ServerURL string
	Width     int
	Height    int
}

func (m WebWorkspaceViewerModel) Init() tea.Cmd {
	return nil
}

func (m WebWorkspaceViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "q", "space", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m WebWorkspaceViewerModel) View() string {
	w := m.Width
	if w <= 0 {
		w = GetTermWidth()
	}
	boxW := w - 4
	if boxW > w-2 {
		boxW = w - 2
	}
	if boxW < 16 {
		boxW = max(10, w-2)
	}
	contentW := boxW - 6
	if contentW < 14 {
		contentW = max(8, boxW-4)
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(1, 2).
		MarginTop(1)

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

	titleText := T("web_launch_title")
	if lipgloss.Width(titleText) > contentW {
		titleText = "🌐 NISKAVA WEB WORKSPACE"
	}
	if lipgloss.Width(titleText) > contentW {
		titleText = "NISKAVA WEB"
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(titleText))
	b.WriteString("\n\n")

	// 1. Status
	statusLabel := T("web_launch_lbl_status")
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(statusLabel), statusStyle.Render("[ONLINE]")))
	if contentW >= 32 {
		b.WriteString(fmt.Sprintf("  %s\n", mutedStyle.Render("(Go SSE Gateway + React SPA)")))
	}

	// 2. URL
	b.WriteString(fmt.Sprintf("• %s : %s\n", lblStyle.Render(T("web_launch_lbl_url")), urlStyle.Render(m.ServerURL)))

	// 3. Features
	b.WriteString(fmt.Sprintf("• %s :\n", lblStyle.Render(T("web_launch_lbl_features"))))
	featLines := wrapText("TradingView Anomaly Markers, ReAct SSE Stream, Evidence Matrix", contentW-2)
	for _, fl := range strings.Split(featLines, "\n") {
		b.WriteString(fmt.Sprintf("  %s\n", valStyle.Render(fl)))
	}

	// 4. Sovereignty
	b.WriteString(fmt.Sprintf("• %s :\n", lblStyle.Render(T("web_launch_lbl_sovereignty"))))
	sovLines := wrapText("100% Local-First SQLite Persistence (~/.niskava/niskava.db)", contentW-2)
	for _, sl := range strings.Split(sovLines, "\n") {
		b.WriteString(fmt.Sprintf("  %s\n", valStyle.Render(sl)))
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(Truncate(T("web_launch_opening"), contentW)))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("[Press Esc or Enter to Return to Menu]"))

	return "\n  " + cardStyle.Render(b.String()) + "\n\033[J"
}

// ShowWebWorkspaceLaunchScreen displays the interactive AltScreen Web Workspace card that live-resizes on window resize.
func ShowWebWorkspaceLaunchScreen(serverURL string) {
	p := tea.NewProgram(WebWorkspaceViewerModel{ServerURL: serverURL}, tea.WithAltScreen())
	_, _ = p.Run()
}

// RenderResponsiveASCIIHeader returns an ASCII banner scaled appropriately for terminal width.
// Width >= 48: 6-line double-line block ASCII art banner (47 cols wide, scaled for standard terminals).
// Width 36 to 47: 2-line mini block ASCII art banner (28 cols wide, compact and non-wrapping).
// Width 30 to 35: Compact styled pill title badge + subtitle.
// Width < 30: Minimal pill badge.
func RenderResponsiveASCIIHeader(width int, style lipgloss.Style) string {
	var b strings.Builder

	if width >= 48 {
		asciiBanner := []string{
			"███╗ ██╗██╗█████╗██╗  ██╗██████╗██╗  ██╗██████╗",
			"████╗██║██║██╔══╝██║ ██╔╝██╔═██╗██║  ██║██╔═██╗",
			"██╔████║██║█████╗█████═╝ ██████║██║  ██║██████║",
			"██║╚███║██║╚══██║██╔═██╗ ██╔═██║╚██╗██╔╝██╔═██║",
			"██║ ╚██║██║█████║██║  ██╗██║ ██║ ╚████╔╝██║ ██║",
			"╚═╝  ╚═╝╚═╝╚════╝╚═╝  ╚═╝╚═╝ ╚═╝  ╚═══╝ ╚═╝ ╚═╝",
		}
		for _, line := range asciiBanner {
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	} else if width >= 36 {
		asciiCompact := []string{
			"█▄ █ █ █▀▀ █▄▀ █▀█ █   █ █▀█",
			"█ ▀█ █ ▄▄█ █ █ █▀█  ▀█▀  █▀█",
		}
		for _, line := range asciiCompact {
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	} else if width >= 30 {
		compactTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
			Padding(0, 1).
			Render("NISKAVA AGENT")
		b.WriteString("\n")
		b.WriteString(compactTitle)
		b.WriteString("  ")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorThought).Italic(true).Render("IDX Market Intelligence"))
		b.WriteString("\n")
	} else {
		compactTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
			Padding(0, 1).
			Render("NISKAVA")
		b.WriteString("\n")
		b.WriteString(compactTitle)
		b.WriteString("\n")
	}

	return b.String()
}
