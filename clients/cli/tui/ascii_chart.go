// Package tui provides interactive terminal interface tools for Niskava Agent.
package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// formatLargeNumber formats large volume or metric numbers cleanly (e.g. 125M, 24M, 1.5B).
func formatLargeNumber(val float64) string {
	absVal := math.Abs(val)
	if absVal >= 1_000_000_000 {
		return fmt.Sprintf("%.2fB", val/1_000_000_000)
	}
	if absVal >= 1_000_000 {
		return fmt.Sprintf("%.2fM", val/1_000_000)
	}
	if absVal >= 1_000 {
		return fmt.Sprintf("%.2fK", val/1_000)
	}
	return fmt.Sprintf("%.2f", val)
}

// RenderASCIIAnomalyChart renders an aesthetic Bloomberg/Binance terminal style ASCII volume & price chart.
func RenderASCIIAnomalyChart(ticker string, anomalies []ipc.Event, days int, overrideWidth ...int) string {
	if days <= 0 {
		days = 30
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

	tier := GetBreakpointTier(w)
	maxSlots := 10
	switch tier {
	case TierCompact:
		maxSlots = 4
	case TierNarrow:
		maxSlots = 8
	case TierNormal:
		maxSlots = 10
	case TierWide:
		maxSlots = 16
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(0, 1).
		MarginTop(1)

	var b strings.Builder

	displayTicker := strings.ToUpper(ticker)
	if displayTicker == "" {
		tickerSet := make(map[string]bool)
		var tickerList []string
		for _, a := range anomalies {
			t := strings.ToUpper(strings.TrimSpace(a.Ticker))
			if t != "" && !tickerSet[t] {
				tickerSet[t] = true
				tickerList = append(tickerList, t)
			}
		}
		if len(tickerList) > 0 {
			displayTicker = strings.Join(tickerList, ", ")
		} else {
			displayTicker = "MARKET"
		}
	}

	titleStr := fmt.Sprintf("📊 %s — Quantitative Anomaly Radar (%d-Day Observation Window)", displayTicker, days)
	if lipgloss.Width(titleStr) > boxW-2 {
		titleStr = fmt.Sprintf("📊 %s — Anomaly Radar (%dD)", displayTicker, days)
	}
	if lipgloss.Width(titleStr) > boxW-2 {
		titleStr = fmt.Sprintf("📊 %s", displayTicker)
	}
	if lipgloss.Width(titleStr) > boxW-2 {
		titleStr = Truncate(titleStr, max(4, boxW-2))
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(titleStr))
	b.WriteString("\n")

	divLine := strings.Repeat("─", max(4, boxW-2))
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(divLine))
	b.WriteString("\n")

	// Deduplicate anomalies by Ticker + Date + MetricType to prevent identical repetitive cards
	seenKeys := make(map[string]bool)
	var uniqueAnomalies []ipc.Event
	for _, a := range anomalies {
		key := fmt.Sprintf("%s|%s|%s", a.Ticker, a.AnomalyDate, a.MetricType)
		if !seenKeys[key] {
			seenKeys[key] = true
			uniqueAnomalies = append(uniqueAnomalies, a)
		}
	}
	anomalies = uniqueAnomalies

	if len(anomalies) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(T("slash_anomalies_empty")))
		b.WriteString("\n")
		return boxStyle.Render(b.String())
	}

	for idx, a := range anomalies {
		zScore := a.ZScore
		priceChange := a.PriceChangePct
		dateStr := a.AnomalyDate
		if dateStr == "" {
			dateStr = "Recent Session"
		}

		zScoreBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(fmt.Sprintf("%+.2fσ", zScore))
		if math.Abs(zScore) >= 3.0 {
			zScoreBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("%+.2fσ", zScore))
		}

		priceBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(fmt.Sprintf("%+.2f%%", priceChange))
		if priceChange < 0 {
			priceBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("%+.2f%%", priceChange))
		}

		// Calculate distinct visual bar lengths dynamically
		volFill := min(maxSlots, max(1, int(math.Abs(zScore)/3.5*float64(maxSlots))))
		priceFill := min(maxSlots, max(1, int(math.Abs(priceChange)/1.5*float64(maxSlots)/10.0)))
		if priceFill < 1 {
			priceFill = 1
		}

		volBarStyle := lipgloss.NewStyle().Foreground(ColorAccent)
		priceBarStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
		priceIcon := "▲"
		if priceChange < 0 {
			priceBarStyle = lipgloss.NewStyle().Foreground(ColorDanger)
			priceIcon = "▼"
		}
		emptyBarStyle := lipgloss.NewStyle().Foreground(ColorMuted)

		volBarChars := volBarStyle.Render(strings.Repeat("▰", volFill)) + emptyBarStyle.Render(strings.Repeat("░", maxSlots-volFill))
		priceBarChars := priceBarStyle.Render(strings.Repeat(priceIcon, priceFill)) + emptyBarStyle.Render(strings.Repeat("░", maxSlots-priceFill))

		metricLabel := a.MetricType
		if metricLabel == "" {
			metricLabel = "VOLUME_AND_PRICE_SURGE"
		}

		// Multiplier ratio calculation (e.g. 125M / 24M = 5.21x)
		surgeRatio := ""
		if a.BaselineValue > 0 {
			multiplier := a.MetricValue / a.BaselineValue
			surgeRatio = fmt.Sprintf("%.2fx surge", multiplier)
		} else {
			surgeRatio = fmt.Sprintf("%.2fσ z-score", zScore)
		}

		cardTicker := strings.ToUpper(strings.TrimSpace(a.Ticker))
		if cardTicker == "" {
			cardTicker = strings.ToUpper(strings.TrimSpace(ticker))
		}
		if cardTicker == "" {
			cardTicker = "N/A"
		}

		cardHeader := fmt.Sprintf("  #%d [%s] %s | Anomaly: %s [SUPPORTED]", idx+1, dateStr, cardTicker, metricLabel)
		if lipgloss.Width(cardHeader) > boxW-2 {
			cardHeader = fmt.Sprintf("  #%d [%s] %s | %s", idx+1, dateStr, cardTicker, metricLabel)
		}
		if lipgloss.Width(cardHeader) > boxW-2 {
			cardHeader = fmt.Sprintf("  #%d [%s] %s", idx+1, dateStr, cardTicker)
		}
		if lipgloss.Width(cardHeader) > boxW-2 {
			cardHeader = Truncate(cardHeader, max(4, boxW-2))
		}

		b.WriteString(cardHeader)
		b.WriteString("\n\n")

		if w >= 55 {
			b.WriteString(fmt.Sprintf("     • Volume Surge (Vz)  : %s  [%s]\n", zScoreBadge, volBarChars))
			b.WriteString(fmt.Sprintf("       ├─ Observed Volume : %s shares\n", formatLargeNumber(a.MetricValue)))
			b.WriteString(fmt.Sprintf("       ├─ MA20 Baseline   : %s shares (%s)\n", formatLargeNumber(a.BaselineValue), surgeRatio))
			b.WriteString(fmt.Sprintf("       └─ Statistical Z   : %s (Threshold: 2.50σ)\n\n", zScoreBadge))

			b.WriteString(fmt.Sprintf("     • Price Action (Rt)  : %s  [%s]\n", priceBadge, priceBarChars))
			b.WriteString(fmt.Sprintf("       ├─ Daily Breakout  : %s\n", priceBadge))
			b.WriteString(fmt.Sprintf("       └─ Sector Diverg.  : %s vs Subsector Median\n", priceBadge))
		} else if w >= 35 {
			b.WriteString(fmt.Sprintf("  • Vol Surge (Vz) : %s [%s]\n", zScoreBadge, volBarChars))
			b.WriteString(fmt.Sprintf("    ├─ Volume   : %s shares\n", formatLargeNumber(a.MetricValue)))
			b.WriteString(fmt.Sprintf("    ├─ Baseline : %s (%s)\n", formatLargeNumber(a.BaselineValue), surgeRatio))
			b.WriteString(fmt.Sprintf("    └─ Z-Score  : %s\n\n", zScoreBadge))

			b.WriteString(fmt.Sprintf("  • Price (Rt)    : %s [%s]\n", priceBadge, priceBarChars))
			b.WriteString(fmt.Sprintf("    ├─ Breakout : %s\n", priceBadge))
			b.WriteString(fmt.Sprintf("    └─ Diverg.  : %s vs Sector\n", priceBadge))
		} else {
			// Calculator mini screen (< 35 cols)
			b.WriteString(fmt.Sprintf("  Vol (Vz) : %s\n", zScoreBadge))
			b.WriteString(fmt.Sprintf("    Obs: %s\n", formatLargeNumber(a.MetricValue)))
			b.WriteString(fmt.Sprintf("    MA20: %s (%s)\n", formatLargeNumber(a.BaselineValue), surgeRatio))
			b.WriteString(fmt.Sprintf("  Price (Rt): %s\n", priceBadge))
		}

		if a.Description != "" {
			descW := max(6, boxW-16)
			wrappedDesc := wrapText(a.Description, descW)
			descLines := strings.Split(wrappedDesc, "\n")
			b.WriteString("\n     ↳ Context: ")
			for iL, dLine := range descLines {
				if iL == 0 {
					b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(dLine) + "\n")
				} else {
					b.WriteString("                " + lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(dLine) + "\n")
				}
			}
		}

		if idx < len(anomalies)-1 {
			b.WriteString("\n  ")
			dottedLen := max(2, (boxW-6)/2)
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("· ", dottedLen)))
			b.WriteString("\n\n")
		}
	}

	return boxStyle.Render(b.String())
}

// AnomalyViewerModel is an interactive AltScreen Bubbletea model for quantitative anomaly radar with mouse scroll support.
type AnomalyViewerModel struct {
	Ticker    string
	Anomalies []ipc.Event
	Days      int
	Viewport  viewport.Model
	Ready     bool
}

func (m AnomalyViewerModel) Init() tea.Cmd {
	return nil
}

func (m AnomalyViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.Viewport.SetContent(RenderASCIIAnomalyChart(m.Ticker, m.Anomalies, m.Days, w))
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m AnomalyViewerModel) View() string {
	if !m.Ready {
		return "\n  Initializing anomaly radar...\n\033[J"
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

// ShowAnomalyViewerScreen displays the interactive AltScreen anomaly radar chart with mouse scroll support.
func ShowAnomalyViewerScreen(ticker string, anomalies []ipc.Event, days int) {
	p := tea.NewProgram(AnomalyViewerModel{Ticker: ticker, Anomalies: anomalies, Days: days}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, _ = p.Run()
}
