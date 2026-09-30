// Package tui provides interactive terminal interface tools for Niskava Agent.
package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/charmbracelet/lipgloss"
)

// RenderASCIIAnomalyChart renders an aesthetic Bloomberg/Binance terminal style ASCII volume & price chart.
func RenderASCIIAnomalyChart(ticker string, anomalies []ipc.Event, days int) string {
	if days <= 0 {
		days = 30
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(0, 1).
		MarginTop(1)

	var b strings.Builder

	titleStr := fmt.Sprintf("📊 %s — Quant Anomaly & Volatility Visualizer (%d-Day Observation Window)", strings.ToUpper(ticker), days)
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(titleStr))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("─────────────────────────────────────────────────────────────────────────────"))
	b.WriteString("\n")

	if len(anomalies) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render(T("slash_anomalies_empty")))
		b.WriteString("\n")
		return boxStyle.Render(b.String())
	}

	// Bar chart sparkline character buckets
	bars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	volBars := []string{"░", "▒", "▓", "█"}

	for idx, a := range anomalies {
		zScore := a.ZScore
		priceChange := a.PriceChangePct
		dateStr := a.AnomalyDate
		if dateStr == "" {
			dateStr = "Recent Session"
		}

		// Calculate sparkline visual bar heights
		volLevel := min(3, max(0, int(math.Abs(zScore)/1.0)))
		priceLevel := min(7, max(0, int((priceChange+10.0)/2.5)))

		volIcon := volBars[volLevel]
		priceIcon := bars[priceLevel]

		zScoreBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(fmt.Sprintf("%.2fσ", zScore))
		if math.Abs(zScore) >= 3.0 {
			zScoreBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("%.2fσ", zScore))
		}

		priceBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(fmt.Sprintf("%+.2f%%", priceChange))
		if priceChange < 0 {
			priceBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("%+.2f%%", priceChange))
		}

		b.WriteString(fmt.Sprintf("  #%d [%s] Ticker: %s | Volume Z-Score: %s | Abnormal Return: %s\n",
			idx+1, dateStr, strings.ToUpper(a.Ticker), zScoreBadge, priceBadge))

		b.WriteString(fmt.Sprintf("     Price Bar  : [%s%s%s%s%s%s%s%s] (%s)\n",
			priceIcon, priceIcon, bars[min(7, priceLevel+1)], priceIcon, bars[max(0, priceLevel-1)], priceIcon, priceIcon, priceIcon, a.MetricType))

		b.WriteString(fmt.Sprintf("     Volume Bar : [%s%s%s%s%s%s%s%s] (Val: %.2f | Baseline: %.2f)\n",
			volIcon, volIcon, volBars[min(3, volLevel+1)], volIcon, volIcon, volIcon, volIcon, volIcon, a.MetricValue, a.BaselineValue))

		if a.Description != "" {
			b.WriteString(fmt.Sprintf("     ↳ %s\n", lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(a.Description)))
		}

		if idx < len(anomalies)-1 {
			b.WriteString("  ")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("· · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·"))
			b.WriteString("\n")
		}
	}

	return boxStyle.Render(b.String())
}
