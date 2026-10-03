// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InvestigationSelectorModel is an interactive Bubbletea menu to pick, inspect, delete, export, and manage investigation sessions.
type InvestigationSelectorModel struct {
	AppDB             *db.DB
	Investigations    []db.Investigation
	Cursor            int
	SelectedSession   *db.Investigation
	Canceled          bool
	FilterQuery       string
	ConfirmDelete     bool
	DeleteTarget      *db.Investigation
	ExportModalActive bool
	ExportFormatIndex int // 0: Markdown (.md), 1: JSON (.json), 2: Plain Text (.txt)
	StatusNotice      string
	StatusNoticeTime  time.Time
	Width             int
	Height            int
}

// sortInvestigations sorts investigations pinned first, then by StartedAt descending.
func sortInvestigations(investigations []db.Investigation) []db.Investigation {
	cloned := make([]db.Investigation, len(investigations))
	copy(cloned, investigations)
	sort.SliceStable(cloned, func(i, j int) bool {
		if cloned[i].IsPinned != cloned[j].IsPinned {
			return cloned[i].IsPinned
		}
		return cloned[i].StartedAt > cloned[j].StartedAt
	})
	return cloned
}

// NewInvestigationSelectorModel creates an interactive selector model for investigation sessions.
func NewInvestigationSelectorModel(investigations []db.Investigation) InvestigationSelectorModel {
	return NewInvestigationSelectorModelWithDB(investigations, nil)
}

// NewInvestigationSelectorModelWithDB creates an interactive selector model backed by local DB for deletion and export.
func NewInvestigationSelectorModelWithDB(investigations []db.Investigation, appDB *db.DB) InvestigationSelectorModel {
	sorted := sortInvestigations(investigations)
	return InvestigationSelectorModel{
		AppDB:          appDB,
		Investigations: sorted,
		Cursor:         0,
		Width:          GetTermWidth(),
		Height:         GetTermHeight(),
	}
}

func (m InvestigationSelectorModel) getFilteredInvestigations() []db.Investigation {
	q := strings.TrimSpace(strings.ToLower(m.FilterQuery))
	if q == "" {
		return m.Investigations
	}
	var filtered []db.Investigation
	for _, inv := range m.Investigations {
		summary := ""
		if inv.SummaryText != nil {
			summary = *inv.SummaryText
		}
		if strings.Contains(strings.ToLower(inv.ID), q) ||
			strings.Contains(strings.ToLower(inv.Ticker), q) ||
			strings.Contains(strings.ToLower(inv.Status), q) ||
			strings.Contains(strings.ToLower(summary), q) {
			filtered = append(filtered, inv)
		}
	}
	return filtered
}

func (m InvestigationSelectorModel) Init() tea.Cmd {
	return nil
}

// ExportInvestigationTranscript saves an investigation audit trail report to ~/.niskava/exports/ in target format.
func ExportInvestigationTranscript(appDB *db.DB, inv *db.Investigation, formatIndex int) (string, error) {
	if inv == nil {
		return "", fmt.Errorf("investigation is nil")
	}

	homeDir, _ := os.UserHomeDir()
	exportDir := filepath.Join(homeDir, ".niskava", "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	ext := "md"
	switch formatIndex {
	case 1:
		ext = "json"
	case 2:
		ext = "txt"
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanID := strings.ReplaceAll(inv.ID, " ", "_")
	filename := fmt.Sprintf("niskava_investigation_%s_%s.%s", cleanID, timestamp, ext)
	outPath := filepath.Join(exportDir, filename)

	var anomalies []db.Anomaly
	var findings []db.Finding
	if appDB != nil {
		anomalies, _ = appDB.GetAnomaliesByInvestigation(inv.ID)
		findings, _ = appDB.ListFindingsByInvestigation(inv.ID)
	}

	var content string
	switch formatIndex {
	case 1: // JSON
		exportObj := map[string]interface{}{
			"investigation": inv,
			"anomalies":     anomalies,
			"findings":      findings,
			"exported_at":   time.Now().Format(time.RFC3339),
			"disclaimer":    "Niskava Agent provides non-advisory market intelligence research. Not financial advice.",
		}
		data, err := json.MarshalIndent(exportObj, "", "  ")
		if err != nil {
			return "", err
		}
		content = string(data)

	case 2: // Plain Text
		var b strings.Builder
		b.WriteString(fmt.Sprintf("NISKAVA INVESTIGATION AUDIT TRAIL - %s (%s)\n", inv.ID, inv.Ticker))
		b.WriteString(fmt.Sprintf("Status: %s | Started: %s\n", inv.Status, inv.StartedAt))
		if inv.SummaryText != nil && *inv.SummaryText != "" {
			b.WriteString(fmt.Sprintf("Summary: %s\n\n", *inv.SummaryText))
		}
		if len(anomalies) > 0 {
			b.WriteString("ANOMALIES:\n")
			for idx, a := range anomalies {
				b.WriteString(fmt.Sprintf(" %d. %s [%s] Val: %.2f, Baseline: %.2f, Z: %.2fσ (%s)\n", idx+1, a.AnomalyDate, a.MetricType, a.MetricValue, a.BaselineValue, a.ZScore, a.Description))
			}
			b.WriteString("\n")
		}
		if len(findings) > 0 {
			b.WriteString("FINDINGS:\n")
			for idx, f := range findings {
				b.WriteString(fmt.Sprintf(" %d. [%s] %s (Conf: %.0f%%)\n    %s\n", idx+1, f.VerificationStatus, f.Title, f.ConfidenceScore*100, f.ClaimText))
			}
		}
		b.WriteString("\n---------------------------------------------------\n")
		b.WriteString("DISCLAIMER: Operational empirical research only. Not investment advice.\n")
		content = b.String()

	default: // Markdown (.md)
		var b strings.Builder
		b.WriteString(fmt.Sprintf("# Niskava Agent — Audit & Investigation Report (%s)\n\n", inv.Ticker))
		b.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", inv.ID))
		b.WriteString(fmt.Sprintf("- **Ticker:** `%s`\n", inv.Ticker))
		b.WriteString(fmt.Sprintf("- **Status:** `%s`\n", inv.Status))
		b.WriteString(fmt.Sprintf("- **Started At:** `%s`\n\n---\n\n", inv.StartedAt))

		if inv.SummaryText != nil && *inv.SummaryText != "" {
			b.WriteString(fmt.Sprintf("## ⚡ Executive Summary\n%s\n\n---\n\n", *inv.SummaryText))
		}

		if len(anomalies) > 0 {
			b.WriteString(fmt.Sprintf("## 📊 Quantitative Anomalies (%d Detected)\n\n", len(anomalies)))
			b.WriteString("| # | Date | Metric | Value | Baseline | Z-Score | Description |\n")
			b.WriteString("|---|---|---|---|---|---|---|\n")
			for idx, a := range anomalies {
				b.WriteString(fmt.Sprintf("| %d | %s | %s | %.2f | %.2f | %.2fσ | %s |\n",
					idx+1, a.AnomalyDate, a.MetricType, a.MetricValue, a.BaselineValue, a.ZScore, a.Description))
			}
			b.WriteString("\n---\n\n")
		}

		if len(findings) > 0 {
			b.WriteString(fmt.Sprintf("## 🔍 Verified Intelligence Findings (%d Emitted)\n\n", len(findings)))
			for idx, f := range findings {
				b.WriteString(fmt.Sprintf("### %d. [%s] %s (Confidence: %.0f%%)\n", idx+1, f.VerificationStatus, f.Title, f.ConfidenceScore*100))
				b.WriteString(fmt.Sprintf("%s\n\n", f.ClaimText))
			}
			b.WriteString("---\n\n")
		}

		b.WriteString("> **Financial Non-Advisory Disclaimer:** Niskava Agent operates under strict POJK/IDX non-advisory guidelines. Research output does not constitute investment advice.\n")
		content = b.String()
	}

	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	return outPath, nil
}

func (m InvestigationSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredInvestigations()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		k := strings.ToLower(msg.String())

		isCtrlC := msg.Type == tea.KeyCtrlC || k == "ctrl+c"
		isCtrlD := msg.Type == tea.KeyCtrlD || k == "ctrl+d" || isControlRune(msg, 4)
		isCtrlP := msg.Type == tea.KeyCtrlP || k == "ctrl+p" || isControlRune(msg, 16)
		isCtrlE := msg.Type == tea.KeyCtrlE || k == "ctrl+e" || isControlRune(msg, 5)
		isCtrlY := msg.Type == tea.KeyCtrlY || k == "ctrl+y" || isControlRune(msg, 25)
		isEsc := msg.Type == tea.KeyEsc || k == "esc"
		isDelete := msg.Type == tea.KeyDelete || k == "delete"

		// Mode A: Confirmation Delete Dialog Active
		if m.ConfirmDelete {
			if k == "y" {
				if m.DeleteTarget != nil {
					targetID := m.DeleteTarget.ID
					if m.AppDB != nil {
						_ = m.AppDB.DeleteInvestigation(targetID)
					}
					// Remove from local list
					var updated []db.Investigation
					for _, inv := range m.Investigations {
						if inv.ID != targetID {
							updated = append(updated, inv)
						}
					}
					m.Investigations = updated
					m.StatusNotice = TF("session_selector_exported_notice", targetID)
					m.StatusNoticeTime = time.Now()
				}
				m.ConfirmDelete = false
				m.DeleteTarget = nil
				m.Cursor = 0
				return m, nil
			} else if k == "n" || isEsc {
				m.ConfirmDelete = false
				m.DeleteTarget = nil
				return m, nil
			}
			return m, nil
		}

		// Mode B: Interactive Export Modal Dialog Active
		if m.ExportModalActive {
			if msg.Type == tea.KeyUp || k == "up" || k == "k" {
				if m.ExportFormatIndex > 0 {
					m.ExportFormatIndex--
				} else {
					m.ExportFormatIndex = 2
				}
				return m, nil
			} else if msg.Type == tea.KeyDown || k == "down" || k == "j" {
				if m.ExportFormatIndex < 2 {
					m.ExportFormatIndex++
				} else {
					m.ExportFormatIndex = 0
				}
				return m, nil
			} else if msg.Type == tea.KeyEnter || k == "enter" {
				if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
					target := filtered[m.Cursor]
					outPath, err := ExportInvestigationTranscript(m.AppDB, &target, m.ExportFormatIndex)
					if err == nil {
						m.StatusNotice = TF("session_selector_exported_notice", outPath)
					} else {
						m.StatusNotice = fmt.Sprintf("❌ Export failed: %v", err)
					}
					m.StatusNoticeTime = time.Now()
				}
				m.ExportModalActive = false
				return m, nil
			} else if isEsc {
				m.ExportModalActive = false
				return m, nil
			}
			return m, nil
		}

		// Mode C: Normal Session Selector Navigation & Hotkeys
		if isEsc || isCtrlC {
			if m.FilterQuery != "" {
				m.FilterQuery = ""
				m.Cursor = 0
				return m, nil
			}
			m.Canceled = true
			return m, tea.Quit
		}

		if isCtrlD || isDelete {
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				m.DeleteTarget = &target
				m.ConfirmDelete = true
				return m, nil
			}
			return m, nil
		}

		if isCtrlP {
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				newPinned := !target.IsPinned
				if m.AppDB != nil {
					_ = m.AppDB.UpdateInvestigationPin(target.ID, newPinned)
				}
				for i, inv := range m.Investigations {
					if inv.ID == target.ID {
						m.Investigations[i].IsPinned = newPinned
						break
					}
				}
				m.Investigations = sortInvestigations(m.Investigations)
				if newPinned {
					m.StatusNotice = T("session_selector_pinned_notice")
				} else {
					m.StatusNotice = T("session_selector_unpinned_notice")
				}
				m.StatusNoticeTime = time.Now()
				return m, nil
			}
			return m, nil
		}

		if isCtrlE {
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				m.ExportModalActive = true
				return m, nil
			}
			return m, nil
		}

		if isCtrlY {
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				target := filtered[m.Cursor]
				summary := "-"
				if target.SummaryText != nil {
					summary = *target.SummaryText
				}
				textToCopy := fmt.Sprintf("[%s] %s (Status: %s)\nSummary: %s", target.ID, target.Ticker, target.Status, summary)
				if err := CopyToClipboard(textToCopy); err == nil {
					m.StatusNotice = T("session_selector_copied_notice")
				} else {
					m.StatusNotice = fmt.Sprintf("📋 Preview copied (clipboard: %v)", err)
				}
				m.StatusNoticeTime = time.Now()
				return m, nil
			}
			return m, nil
		}

		if msg.Type == tea.KeyUp || k == "up" || (m.FilterQuery == "" && k == "k") {
			if m.Cursor > 0 {
				m.Cursor--
			} else if len(filtered) > 0 {
				m.Cursor = len(filtered) - 1
			}
			return m, nil
		}

		if msg.Type == tea.KeyDown || k == "down" || (m.FilterQuery == "" && k == "j") {
			if m.Cursor < len(filtered)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}
			return m, nil
		}

		if msg.Type == tea.KeyPgUp || k == "pgup" || k == "pageup" {
			if m.Cursor >= 5 {
				m.Cursor -= 5
			} else {
				m.Cursor = 0
			}
			return m, nil
		}

		if msg.Type == tea.KeyPgDown || k == "pgdown" || k == "pagedown" {
			if len(filtered) > 0 {
				if m.Cursor+5 < len(filtered) {
					m.Cursor += 5
				} else {
					m.Cursor = len(filtered) - 1
				}
			}
			return m, nil
		}

		if msg.Type == tea.KeyHome || k == "home" {
			m.Cursor = 0
			return m, nil
		}

		if msg.Type == tea.KeyEnd || k == "end" {
			if len(filtered) > 0 {
				m.Cursor = len(filtered) - 1
			}
			return m, nil
		}

		if msg.Type == tea.KeyBackspace || k == "backspace" {
			if len(m.FilterQuery) > 0 {
				m.FilterQuery = m.FilterQuery[:len(m.FilterQuery)-1]
				m.Cursor = 0
			}
			return m, nil
		}

		if msg.Type == tea.KeyEnter || k == "enter" {
			if len(filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(filtered) {
				selected := filtered[m.Cursor]
				m.SelectedSession = &selected
			}
			return m, tea.Quit
		}

		// Strictly filter text input: Only append printable runes
		if len(msg.Runes) > 0 {
			hasPrintable := false
			for _, r := range msg.Runes {
				if r >= 32 && r != 127 {
					m.FilterQuery += string(r)
					hasPrintable = true
				}
			}
			if hasPrintable {
				m.Cursor = 0
			}
		}
	}
	return m, nil
}

func (m InvestigationSelectorModel) View() string {
	var b strings.Builder

	termW := m.Width
	if termW <= 0 {
		termW = GetTermWidth()
	}
	termH := m.Height
	if termH <= 0 {
		termH = GetTermHeight()
	}
	boxW := termW - 4
	if boxW > termW-2 {
		boxW = termW - 2
	}
	if boxW < 16 {
		boxW = max(10, termW-2)
	}

	sessionBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(0, 1).
		Foreground(ColorFg)

	title := T("investigation_selector_title")
	b.WriteString(sessionTitleStyle.Render(title))
	b.WriteString("\n\n")

	// Render confirmation delete dialog if active (boxless, clean, responsive)
	if m.ConfirmDelete && m.DeleteTarget != nil {
		var delB strings.Builder
		delB.WriteString(sessionTitleStyle.Render(title))
		delB.WriteString("\n\n")

		delHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(T("delete_warning_header"))
		targetID := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(m.DeleteTarget.ID)
		if m.DeleteTarget.Ticker != "" {
			targetID += fmt.Sprintf(" [%s]", m.DeleteTarget.Ticker)
		}

		bodyMsg := lipgloss.NewStyle().Foreground(ColorMuted).Render(T("delete_warning_body"))
		permanentAlert := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(T("delete_warning_permanent"))
		promptMsg := lipgloss.NewStyle().Bold(true).Foreground(ColorFg).Render(TF("delete_warning_prompt", targetID))
		footerMsg := lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render(T("delete_warning_footer"))

		delB.WriteString(delHeader + "\n\n")
		delB.WriteString(bodyMsg + "\n")
		delB.WriteString(permanentAlert + "\n\n")
		delB.WriteString(promptMsg + "\n\n")
		delB.WriteString(footerMsg)

		return "\n" + sessionBoxStyle.Render(delB.String()) + "\n\033[J"
	}

	// Render Export Modal Dialog if active (boxless, clean, responsive)
	if m.ExportModalActive {
		var expB strings.Builder
		expB.WriteString(sessionTitleStyle.Render(title))
		expB.WriteString("\n\n")

		exportTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(T("session_selector_export_title"))
		expB.WriteString(exportTitle + "\n")
		expB.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("export_dialog_prompt")) + "\n")
		expB.WriteString(lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render(T("export_target_location_hint")) + "\n\n")

		formats := []struct {
			name string
			desc string
		}{
			{"Markdown Report (.md)", T("export_fmt_md_desc")},
			{"Raw JSON Audit Trail (.json)", T("export_fmt_json_desc")},
			{"Plain Text Summary (.txt)", T("export_fmt_txt_desc")},
		}

		for i, fmtObj := range formats {
			if i == m.ExportFormatIndex {
				optName := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(fmt.Sprintf("  ▸ [•] %-28s", fmtObj.name))
				optDesc := lipgloss.NewStyle().Foreground(ColorFg).Render(fmt.Sprintf(" - %s", fmtObj.desc))
				expB.WriteString(optName + optDesc + "\n")
			} else {
				optName := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("    [ ] %-28s", fmtObj.name))
				optDesc := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf(" - %s", fmtObj.desc))
				expB.WriteString(optName + optDesc + "\n")
			}
		}

		expB.WriteString("\n")
		expB.WriteString(lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render(T("export_modal_footer_hint")))

		return "\n" + sessionBoxStyle.Render(expB.String()) + "\n\033[J"
	}

	filtered := m.getFilteredInvestigations()
	totalAll := len(m.Investigations)
	totalFiltered := len(filtered)

	// Render temporary status notice message if active
	if m.StatusNotice != "" && time.Since(m.StatusNoticeTime) <= 4*time.Second {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(m.StatusNotice))
		b.WriteString("\n\n")
	}

	// Always render visible search/filter bar box
	searchPlaceholder := T("session_selector_search_placeholder")

	searchVal := m.FilterQuery
	if searchVal == "" {
		searchVal = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(searchPlaceholder)
	} else {
		searchVal = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(searchVal)
	}

	searchBar := fmt.Sprintf("🔍 Filter: [ %s ] (%d/%d)", searchVal, totalFiltered, totalAll)
	b.WriteString(searchBar)
	b.WriteString("\n\n")

	if totalAll == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("sessions_inv_empty")))
		b.WriteString("\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n\033[J"
	}

	if totalFiltered == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("session_selector_no_match")))
		b.WriteString("\n")
		return "\n" + sessionBoxStyle.Render(b.String()) + "\n\033[J"
	}

	// Sliding Viewport Window dynamically bounded by terminal height
	overhead := 10
	if m.StatusNotice != "" {
		overhead += 2
	}
	maxVisible := (termH - overhead) / 2
	if maxVisible < 2 {
		maxVisible = 2
	}
	if maxVisible > 8 {
		maxVisible = 8
	}

	windowStart := 0
	if m.Cursor >= maxVisible {
		windowStart = m.Cursor - maxVisible + 1
	}
	windowEnd := windowStart + maxVisible
	if windowEnd > totalFiltered {
		windowEnd = totalFiltered
		windowStart = max(0, windowEnd-maxVisible)
	}

	if totalFiltered > maxVisible {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(TF("session_selector_showing", windowStart+1, windowEnd, totalFiltered)))
		b.WriteString("\n\n")
	}

	for i := windowStart; i < windowEnd; i++ {
		inv := filtered[i]
		dateStr := inv.StartedAt
		if len(dateStr) > 16 {
			dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
		}

		summary := "-"
		if inv.SummaryText != nil && *inv.SummaryText != "" {
			summary = SanitizePreviewText(*inv.SummaryText)
		}
		summary = Truncate(summary, boxW-10)

		statusStyled := inv.Status
		switch inv.Status {
		case "COMPLETED":
			statusStyled = lipgloss.NewStyle().Foreground(ColorSuccess).Render("COMPLETED")
		case "RUNNING":
			statusStyled = lipgloss.NewStyle().Foreground(ColorAccent).Render("RUNNING")
		case "FAILED":
			statusStyled = lipgloss.NewStyle().Foreground(ColorDanger).Render("FAILED")
		}

		tickerLabel := inv.Ticker
		if tickerLabel == "" {
			tickerLabel = "IDX"
		}
		if inv.IsPinned {
			tickerLabel = "📌 " + tickerLabel
		}

		var lineTitle string
		if boxW < 50 {
			lineTitle = fmt.Sprintf("%s | %s [%s]", inv.ID, tickerLabel, statusStyled)
		} else {
			lineTitle = fmt.Sprintf("%-22s %-8s [%s] (%s)", inv.ID, tickerLabel, statusStyled, dateStr)
		}
		previewLine := fmt.Sprintf("    ↳ %s", summary)

		if i == m.Cursor {
			b.WriteString(sessionCursorStyle.Render("> "))
			b.WriteString(sessionActiveStyle.Render(lineTitle))
		} else {
			b.WriteString(sessionMetaStyle.Render("  "))
			b.WriteString(lineTitle)
		}
		b.WriteString("\n")
		b.WriteString(sessionMetaStyle.Render(previewLine))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render(T("session_selector_hint")))

	return "\n" + sessionBoxStyle.Render(b.String()) + "\n\033[J"
}
