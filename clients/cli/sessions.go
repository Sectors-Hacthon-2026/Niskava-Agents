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
	chats, _, err := database.ListChatSessions(30, 0, "")
	if err != nil || len(chats) == 0 {
		_ = printFormattedSessions(database, "all", 20, os.Stdout)
		tui.PromptPressEscToReturn()
		return ""
	}

	selector := tui.NewSessionSelectorModel(chats)
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
	Long: `List, search, delete, and export historical chat sessions and investigation audit trails from local SQLite storage.`,
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
		if err := appDB.DeleteChatSession(sessionID); err != nil {
			return fmt.Errorf("failed to delete session '%s': %w", sessionID, err)
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
				if m.Role == "user" {
					sb.WriteString(fmt.Sprintf("### 👤 User Prompt\n> %s\n\n", m.Content))
				} else if m.Role == "assistant" {
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
	sessionsCmd.Flags().IntVarP(&limitFlag, "limit", "n", 20, "maximum number of sessions to display")
	sessionsCmd.Flags().StringVarP(&sessionTypeFlag, "type", "t", "all", "session types to display: 'chat', 'investigation', or 'all'")

	sessionsExportCmd.Flags().StringVarP(&exportFormatFlag, "format", "f", "md", "export format: 'md' or 'json'")
	sessionsExportCmd.Flags().StringVarP(&exportOutFlag, "out", "o", "", "output file path")

	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsSearchCmd)
	sessionsCmd.AddCommand(sessionsExportCmd)

	RootCmd.AddCommand(sessionsCmd)
}
