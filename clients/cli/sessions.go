package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	limitFlag        int
	sessionTypeFlag  string
	exportFormatFlag string
	exportOutFlag    string
)

func printFormattedSessions(database *db.DB, sType string, limit int, w io.Writer) error {
	sType = strings.ToLower(strings.TrimSpace(sType))
	if limit <= 0 {
		limit = 20
	}

	if sType == "all" || sType == "chat" || sType == "chats" {
		chats, _, err := database.ListChatSessions(limit, 0, "")
		if err != nil {
			return fmt.Errorf("failed to retrieve chat sessions: %w", err)
		}

		fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FCD535")).Render(tui.T("sessions_chat_title")))
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "%-22s %-24s %-8s %-16s %s\n", "SESSION ID", "TITLE", "MSGS", "UPDATED AT", "PREVIEW")
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")

		if len(chats) == 0 {
			fmt.Fprintln(w, tui.T("sessions_chat_empty"))
		} else {
			for _, s := range chats {
				preview := s.LastMessagePreview
				if len(preview) > 28 {
					preview = preview[:25] + "..."
				}
				if preview == "" {
					preview = "-"
				}
				dateStr := s.UpdatedAt
				if len(dateStr) > 16 {
					dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
				}
				title := s.Title
				if len(title) > 22 {
					title = title[:19] + "..."
				}
				fmt.Fprintf(w, "%-22s %-24s %-8d %-16s %s\n", s.ID, title, s.MessageCount, dateStr, preview)
			}
		}
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintln(w, lipgloss.NewStyle().Foreground(lipgloss.Color("#848E9C")).Italic(true).Render(tui.T("sessions_chat_tip")))
	}

	if sType == "all" || sType == "investigation" || sType == "investigations" || sType == "inv" {
		investigations, err := database.ListInvestigations(limit)
		if err != nil {
			return fmt.Errorf("failed to retrieve investigations: %w", err)
		}

		fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FCD535")).Render(tui.T("sessions_inv_title")))
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "%-22s %-8s %-12s %-16s %s\n", "SESSION ID", "TICKER", "STATUS", "STARTED AT", "SUMMARY")
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")

		if len(investigations) == 0 {
			fmt.Fprintln(w, tui.T("sessions_inv_empty"))
		} else {
			for _, inv := range investigations {
				summary := "-"
				if inv.SummaryText != nil && *inv.SummaryText != "" {
					summary = *inv.SummaryText
					if len(summary) > 28 {
						summary = summary[:25] + "..."
					}
				}
				statusStyled := inv.Status
				switch inv.Status {
				case "COMPLETED":
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#0ECB81")).Render("COMPLETED")
				case "RUNNING":
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#FCD535")).Render("RUNNING")
				case "FAILED":
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#F6465D")).Render("FAILED")
				}
				dateStr := inv.StartedAt
				if len(dateStr) > 16 {
					dateStr = strings.Replace(dateStr[:16], "T", " ", 1)
				}
				fmt.Fprintf(w, "%-22s %-8s %-12s %-16s %s\n", inv.ID, inv.Ticker, statusStyled, dateStr, summary)
			}
		}
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
	}

	return nil
}

func runSessionsInteractive(cmd *cobra.Command, database *db.DB) string {
	if database == nil {
		return ""
	}

	// Step 1: Open Gateway Sub-menu (Chat Sessions vs Investigation Audit Trails)
	gw := tui.NewSessionGatewayModel()
	pGw := tea.NewProgram(gw, tea.WithAltScreen())
	mGw, errGw := pGw.Run()
	if errGw != nil {
		return ""
	}

	resGw := mGw.(tui.SessionGatewayModel)
	if resGw.Canceled || resGw.Selected == "" {
		return ""
	}

	// Step 2: Route according to user selection
	if resGw.Selected == "investigations" {
		invs, err := database.ListInvestigations(30)
		if err != nil || len(invs) == 0 {
			_ = printFormattedSessions(database, "investigations", 20, os.Stdout)
			tui.PromptPressEscToReturn()
			return ""
		}
		selector := tui.NewInvestigationSelectorModelWithDB(invs, database)
		p := tea.NewProgram(selector, tea.WithAltScreen())
		m, err := p.Run()
		if err == nil {
			res := m.(tui.InvestigationSelectorModel)
			if !res.Canceled && res.SelectedSession != nil {
				return res.SelectedSession.ID
			}
		}
		return ""
	}

	// Default: Chat Sessions
	chats, _, err := database.ListChatSessions(30, 0, "")
	if err != nil || len(chats) == 0 {
		_ = printFormattedSessions(database, "chat", 20, os.Stdout)
		tui.PromptPressEscToReturn()
		return ""
	}

	selector := tui.NewSessionSelectorModelWithDB(chats, database)
	p := tea.NewProgram(selector, tea.WithAltScreen())
	m, err := p.Run()
	if err == nil {
		res := m.(tui.SessionSelectorModel)
		if !res.Canceled && res.SelectedSession != nil {
			return res.SelectedSession.ID
		}
	}
	return ""
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage and inspect chat and investigation sessions from local SQLite storage",
	Long:  `List, search, delete, and export historical chat sessions and investigation audit trails from local SQLite storage.`,
	Example: `  # List 20 recent sessions:
  niskava sessions
  niskava sessions list

  # Delete a session permanently:
  niskava sessions delete CHAT-20260926-0001

  # Search past chat history:
  niskava sessions search ANTM

  # Export session report to Markdown:
  niskava sessions export CHAT-20260926-0001 --format md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		return printFormattedSessions(appDB, sessionTypeFlag, limitFlag, os.Stdout)
	},
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List previous chat and investigation sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		return printFormattedSessions(appDB, sessionTypeFlag, limitFlag, os.Stdout)
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete <SESSION_ID>",
	Short: "Delete a session and its chat history permanently",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		sessionID := strings.TrimSpace(args[0])

		// Try deleting as chat session first
		err := appDB.DeleteChatSession(sessionID)
		if err != nil {
			// If not a chat session, try deleting as an investigation session
			errInv := appDB.DeleteInvestigation(sessionID)
			if errInv != nil {
				return fmt.Errorf("failed to delete session '%s': %w", sessionID, err)
			}
		}

		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#0ECB81")).Bold(true).Render(
			tui.TF("sessions_delete_success", sessionID),
		))
		return nil
	},
}

var sessionsSearchCmd = &cobra.Command{
	Use:   "search <KEYWORD>",
	Short: "Search past chat messages across all sessions by keyword",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		keyword := strings.TrimSpace(args[0])
		results, err := appDB.SearchChatMessages(keyword, 20)
		if err != nil {
			return fmt.Errorf("failed to search messages: %w", err)
		}
		if len(results) == 0 {
			fmt.Println(tui.TF("sessions_search_empty", keyword))
			return nil
		}

		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FCD535")).Render(
			tui.TF("sessions_search_title", keyword),
		))
		fmt.Println("─────────────────────────────────────────────────────────────────────────────")
		for idx, r := range results {
			snip := r.Content
			if len(snip) > 100 {
				snip = snip[:97] + "..."
			}
			fmt.Printf("  %d. [%s] [%s] %s: %s\n", idx+1, r.SessionID, r.CreatedAt, r.Role, snip)
		}
		fmt.Println("─────────────────────────────────────────────────────────────────────────────")
		return nil
	},
}

var sessionsExportCmd = &cobra.Command{
	Use:   "export <SESSION_ID>",
	Short: "Export session transcript to Markdown or JSON report file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		sessionID := strings.TrimSpace(args[0])
		format := strings.ToLower(strings.TrimSpace(exportFormatFlag))
		if format == "" {
			format = "md"
		}

		outPath := exportOutFlag
		if outPath == "" {
			outPath = fmt.Sprintf("niskava_report_%s.%s", sessionID, format)
		}

		history, err := appDB.GetChatHistory(sessionID, 100)
		if (err != nil || len(history) == 0) && strings.HasPrefix(strings.ToUpper(sessionID), "INV-") {
			// Handle investigation session export
			invData, errInv := appDB.GetInvestigation(sessionID)
			if errInv != nil || invData == nil {
				return fmt.Errorf("%s", tui.TF("sessions_export_no_history", sessionID))
			}

			anomalies, _ := appDB.GetAnomaliesByInvestigation(sessionID)
			findings, _ := appDB.ListFindingsByInvestigation(sessionID)

			var content string
			if format == "json" {
				data, _ := json.MarshalIndent(map[string]interface{}{
					"investigation": invData,
					"anomalies":     anomalies,
					"findings":      findings,
				}, "", "  ")
				content = string(data)
			} else {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("# Niskava Agent — Audit & Investigation Report (%s)\n\n", invData.Ticker))
				sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
				sb.WriteString(fmt.Sprintf("- **Ticker:** `%s`\n", invData.Ticker))
				sb.WriteString(fmt.Sprintf("- **Date:** `%s`\n\n---\n\n", time.Now().Format("2006-01-02 15:04:05 MST")))

				if invData.SummaryText != nil && *invData.SummaryText != "" {
					sb.WriteString(fmt.Sprintf("## ⚡ Executive Summary\n%s\n\n---\n\n", *invData.SummaryText))
				}

				if len(anomalies) > 0 {
					sb.WriteString(fmt.Sprintf("## 📊 Quantitative Anomalies (%d Detected)\n\n", len(anomalies)))
					sb.WriteString("| # | Date | Metric | Value | Baseline | Z-Score | Description |\n")
					sb.WriteString("|---|---|---|---|---|---|---|\n")
					for idx, a := range anomalies {
						sb.WriteString(fmt.Sprintf("| %d | %s | %s | %.2f | %.2f | %.2fσ | %s |\n",
							idx+1, a.AnomalyDate, a.MetricType, a.MetricValue, a.BaselineValue, a.ZScore, a.Description))
					}
					sb.WriteString("\n---\n\n")
				}

				if len(findings) > 0 {
					sb.WriteString(fmt.Sprintf("## 🔍 Verified Intelligence Findings (%d Emitted)\n\n", len(findings)))
					for idx, f := range findings {
						sb.WriteString(fmt.Sprintf("### %d. [%s] %s (Confidence: %.0f%%)\n", idx+1, f.VerificationStatus, f.Title, f.ConfidenceScore*100))
						sb.WriteString(fmt.Sprintf("%s\n\n", f.ClaimText))
					}
					sb.WriteString("---\n\n")
				}

				sb.WriteString(tui.T("sessions_export_disclaimer"))
				content = sb.String()
			}

			if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write export file: %w", err)
			}

			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#0ECB81")).Bold(true).Render(
				tui.TF("sessions_export_success", outPath),
			))
			return nil
		}

		if err != nil || len(history) == 0 {
			return fmt.Errorf("%s", tui.TF("sessions_export_no_history", sessionID))
		}

		var content string
		if format == "json" {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"session_id": sessionID,
				"messages":   history,
			}, "", "  ")
			content = string(data)
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# Niskava Agent — Audit & Research Report\n\n"))
			sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
			sb.WriteString(fmt.Sprintf("- **Date:** `%s`\n\n---\n\n", time.Now().Format("2006-01-02 15:04:05 MST")))
			for _, m := range history {
				r := strings.ToLower(strings.TrimSpace(m.Role))
				if r == "user" || r == "human" {
					sb.WriteString(fmt.Sprintf("### 👤 User Prompt\n> %s\n\n", m.Content))
				} else {
					sb.WriteString(fmt.Sprintf("### ⚡ Niskava Agent Findings\n%s\n\n---\n\n", m.Content))
				}
			}
			sb.WriteString(tui.T("sessions_export_disclaimer"))
			content = sb.String()
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write export file: %w", err)
		}

		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#0ECB81")).Bold(true).Render(
			tui.TF("sessions_export_success", outPath),
		))
		return nil
	},
}

func init() {
	sessionsCmd.PersistentFlags().IntVarP(&limitFlag, "limit", "n", 20, "maximum number of sessions to display")
	sessionsCmd.PersistentFlags().StringVarP(&sessionTypeFlag, "type", "t", "all", "session types to display: 'chat', 'investigation', or 'all'")

	sessionsExportCmd.Flags().StringVarP(&exportFormatFlag, "format", "f", "md", "export format: 'md' or 'json'")
	sessionsExportCmd.Flags().StringVarP(&exportOutFlag, "out", "o", "", "output file path")

	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsSearchCmd)
	sessionsCmd.AddCommand(sessionsExportCmd)

	RootCmd.AddCommand(sessionsCmd)
}
