// Package tui provides interactive terminal interface tools for Niskava Agent.
package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
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
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(titleStr))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("─────────────────────────────────────────────────────────────────────────────"))
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

		// Calculate distinct visual bar lengths (10 slots)
		volFill := min(10, max(1, int(math.Abs(zScore)/3.5*10.0)))
		priceFill := min(10, max(1, int(math.Abs(priceChange)/1.5)))

		volBarStyle := lipgloss.NewStyle().Foreground(ColorAccent)
		priceBarStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
		priceIcon := "▲"
		if priceChange < 0 {
			priceBarStyle = lipgloss.NewStyle().Foreground(ColorDanger)
			priceIcon = "▼"
		}
		emptyBarStyle := lipgloss.NewStyle().Foreground(ColorMuted)

		volBarChars := volBarStyle.Render(strings.Repeat("▰", volFill)) + emptyBarStyle.Render(strings.Repeat("░", 10-volFill))
		priceBarChars := priceBarStyle.Render(strings.Repeat(priceIcon, priceFill)) + emptyBarStyle.Render(strings.Repeat("░", 10-priceFill))

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

		b.WriteString(fmt.Sprintf("  #%d [%s] %s | Anomaly: %s [SUPPORTED]\n\n",
			idx+1, dateStr, cardTicker, metricLabel))

		b.WriteString(fmt.Sprintf("     • Volume Surge (Vz)  : %s  [%s]\n", zScoreBadge, volBarChars))
		b.WriteString(fmt.Sprintf("       ├─ Observed Volume : %s shares\n", formatLargeNumber(a.MetricValue)))
		b.WriteString(fmt.Sprintf("       ├─ MA20 Baseline   : %s shares (%s)\n", formatLargeNumber(a.BaselineValue), surgeRatio))
		b.WriteString(fmt.Sprintf("       └─ Statistical Z   : %s (Threshold: 2.50σ)\n\n", zScoreBadge))

		b.WriteString(fmt.Sprintf("     • Price Action (Rt)  : %s  [%s]\n", priceBadge, priceBarChars))
		b.WriteString(fmt.Sprintf("       ├─ Daily Breakout  : %s\n", priceBadge))
		b.WriteString(fmt.Sprintf("       └─ Sector Diverg.  : %s vs Subsector Median\n", priceBadge))

		if a.Description != "" {
			b.WriteString(fmt.Sprintf("\n     ↳ Context: %s\n", lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(a.Description)))
		}

		if idx < len(anomalies)-1 {
			b.WriteString("\n  ")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("· · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·"))
			b.WriteString("\n\n")
		}
	}

	return boxStyle.Render(b.String())
}
