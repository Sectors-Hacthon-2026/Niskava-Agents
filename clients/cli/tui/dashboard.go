package tui

import (
	"fmt"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MaskAPIKey returns a masked preview of a secret key preserving head and tail.
func MaskAPIKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "(not configured)"
	}
	if len(trimmed) <= 8 {
		return "••••"
	}
	headLen := 4
	if strings.HasPrefix(trimmed, "sk-") {
		headLen = 5
	} else if strings.HasPrefix(trimmed, "AIzaSy") {
		headLen = 6
	}
	tailLen := 4
	if len(trimmed) <= headLen+tailLen {
		return trimmed[:headLen] + "••••"
	}
	return trimmed[:headLen] + "••••" + trimmed[len(trimmed)-tailLen:]
}

// RenderConfigurationDashboard builds an aesthetic Lipgloss card summarizing the current active configuration.
func RenderConfigurationDashboard(c *config.Config, overrideWidth ...int) string {
	if c == nil {
		c = config.DefaultConfig()
	}

	w := GetTermWidth()
	if len(overrideWidth) > 0 && overrideWidth[0] > 0 {
		w = overrideWidth[0]
	}
	boxW := w - 4
	if boxW > w-2 {
		boxW = w - 2
	}
	if boxW < 16 {
		boxW = max(10, w-2)
	}
	if boxW > 120 {
		boxW = 120
	}

	lblWidth := 16
	if w < 50 {
		lblWidth = 10
	}
	if w < 35 {
		lblWidth = 8
	}

	contentW := max(4, boxW-(lblWidth+6))

	dashboardTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	dashboardCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(0, 1).
		Foreground(ColorFg)

	var sb strings.Builder
	titleText := "CURRENT ACTIVE CONFIGURATION"
	if boxW < 40 {
		titleText = "CONFIGURATION"
	}
	sb.WriteString(dashboardTitleStyle.Render(titleText) + "\n\n")

	provider := c.Auth.AIProvider
	if provider == "" {
		provider = "openai"
	}

	activeModel := c.Auth.OpenAIModel
	if strings.EqualFold(provider, "gemini") {
		activeModel = c.Auth.GeminiModel
	}

	sectorsStatus := "Live IDX API (" + MaskAPIKey(c.Auth.SectorsAPIKey) + ")"
	if strings.TrimSpace(c.Auth.SectorsAPIKey) == "" {
		sectorsStatus = "Not Configured (Required)"
	}

	fmtFmt := fmt.Sprintf("  • %%-%ds: %%s\n", lblWidth)

	sb.WriteString(fmt.Sprintf(fmtFmt, "AI Provider", lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(Truncate(provider, contentW))))
	sb.WriteString(fmt.Sprintf(fmtFmt, "Active Model", Truncate(activeModel, contentW)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "Endpoint BaseURL", Truncate(c.Auth.OpenAIBaseURL, contentW)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "Gemini API Key", MaskAPIKey(c.Auth.GeminiAPIKey)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "OpenAI/Router Key", MaskAPIKey(c.Auth.OpenAIAPIKey)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "Sectors Financial", Truncate(sectorsStatus, contentW)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "LLM Timeout", fmt.Sprintf("%.0fs", c.Preferences.LLMTimeoutSecs)))
	sb.WriteString(fmt.Sprintf(fmtFmt, "Python Engine", TruncateMiddle(c.Engine.PythonBin, contentW)))

	var telegramStatus string
	if strings.TrimSpace(c.Telegram.BotToken) != "" {
		statusBadge := "Disabled"
		if c.Telegram.Enabled {
			statusBadge = "Enabled"
		}
		userCountStr := "Open access"
		if len(c.Telegram.AllowedUsers) > 0 {
			userCountStr = fmt.Sprintf("%d users", len(c.Telegram.AllowedUsers))
		}
		telegramStatus = fmt.Sprintf("%s (%s, %s)", MaskAPIKey(c.Telegram.BotToken), statusBadge, userCountStr)
	} else {
		telegramStatus = "Not configured (Optional)"
	}
	sb.WriteString(fmt.Sprintf(fmtFmt, "Telegram Bot", Truncate(telegramStatus, contentW)))

	return dashboardCardStyle.Render(sb.String())
}

// ConfigViewerModel is an interactive AltScreen Bubbletea model for configuration dashboard with mouse scroll support.
type ConfigViewerModel struct {
	CFG      *config.Config
	Viewport viewport.Model
	Ready    bool
}

func (m ConfigViewerModel) Init() tea.Cmd {
	return nil
}

func (m ConfigViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.Viewport.SetContent(RenderConfigurationDashboard(m.CFG, w))
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m ConfigViewerModel) View() string {
	if !m.Ready {
		return "\n  Initializing configuration...\n\033[J"
	}
	w := m.Viewport.Width
	if w <= 0 {
		w = GetTermWidth()
	}
	footerText := "[↑/↓/k/j/Mouse Scroll  •  Esc Return to Menu]"
	if w < 55 {
		footerText = "[↑/↓ Scroll  •  Esc Return]"
	}
	footer := lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(footerText)
	return fmt.Sprintf("%s\n\n  %s", m.Viewport.View(), footer) + "\033[J"
}

// ShowConfigurationScreen launches interactive AltScreen configuration card that scales live on window resize.
func ShowConfigurationScreen(cfg *config.Config) {
	p := tea.NewProgram(ConfigViewerModel{CFG: cfg}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, _ = p.Run()
}
