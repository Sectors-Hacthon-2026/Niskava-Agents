package tui

import (
	"fmt"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
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
func RenderConfigurationDashboard(c *config.Config) string {
	if c == nil {
		c = config.DefaultConfig()
	}

	dashboardTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBg).
		Background(ColorAccent).
		Padding(0, 1)

	dashboardCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(78).
		Padding(0, 1).
		Foreground(ColorFg)

	var sb strings.Builder
	sb.WriteString(dashboardTitleStyle.Render("CURRENT ACTIVE CONFIGURATION") + "\n\n")

	provider := c.Auth.AIProvider
	if provider == "" {
		provider = "openai"
	}

	activeModel := c.Auth.OpenAIModel
	if strings.EqualFold(provider, "gemini") {
		activeModel = c.Auth.GeminiModel
	}

	sectorsStatus := "Live IDX API (" + MaskAPIKey(c.Auth.SectorsAPIKey) + ")"
	if c.Preferences.OfflineMode || strings.TrimSpace(c.Auth.SectorsAPIKey) == "" {
		sectorsStatus = "Offline Mock Mode (Law 5 Credit Conservation)"
	}

	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "AI Provider", lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(provider)))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Active Model", activeModel))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Endpoint BaseURL", c.Auth.OpenAIBaseURL))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Gemini API Key", MaskAPIKey(c.Auth.GeminiAPIKey)))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "OpenAI/Router Key", MaskAPIKey(c.Auth.OpenAIAPIKey)))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Sectors Financial", sectorsStatus))
	sb.WriteString(fmt.Sprintf("  • %-16s: %.0fs\n", "LLM Timeout", c.Preferences.LLMTimeoutSecs))
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Python Engine", c.Engine.PythonBin))

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
	sb.WriteString(fmt.Sprintf("  • %-16s: %s\n", "Telegram Bot", telegramStatus))

	return dashboardCardStyle.Render(sb.String())
}
