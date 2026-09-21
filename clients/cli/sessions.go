package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	limitFlag       int
	sessionTypeFlag string
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

		fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FCD535")).Render("\n💬 RIWAYAT SESI CHAT AI (CONVERSATIONAL SESSIONS)"))
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "%-22s %-24s %-8s %-16s %s\n", "SESSION ID", "TITLE", "MSGS", "UPDATED AT", "PREVIEW")
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")

		if len(chats) == 0 {
			fmt.Fprintln(w, "  (Belum ada riwayat sesi chat tersimpan)")
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
		fmt.Fprintln(w, lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true).Render("Tip: Gunakan 'niskava -s <SESSION_ID>' atau '/resume <ID>' di REPL untuk melanjutkan sesi."))
	}

	if sType == "all" || sType == "investigation" || sType == "investigations" || sType == "inv" {
		investigations, err := database.ListInvestigations(limit)
		if err != nil {
			return fmt.Errorf("failed to retrieve investigations: %w", err)
		}

		fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5FF")).Render("\n📊 RIWAYAT INVESTIGASI AUDIT PASAR (PIPELINE SESSIONS)"))
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "%-22s %-8s %-12s %-16s %s\n", "SESSION ID", "TICKER", "STATUS", "STARTED AT", "SUMMARY")
		fmt.Fprintln(w, "─────────────────────────────────────────────────────────────────────────────")

		if len(investigations) == 0 {
			fmt.Fprintln(w, "  (Belum ada riwayat investigasi)")
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
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676")).Render("COMPLETED")
				case "RUNNING":
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E5FF")).Render("RUNNING")
				case "FAILED":
					statusStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1744")).Render("FAILED")
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
		fmt.Println("Tekan Enter untuk kembali ke Menu...")
		_, _ = fmt.Scanln()
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
	Short: "List previous chat and investigation sessions from local SQLite storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appDB == nil {
			return fmt.Errorf("database not initialized")
		}
		return printFormattedSessions(appDB, sessionTypeFlag, limitFlag, os.Stdout)
	},
}

func init() {
	sessionsCmd.Flags().IntVarP(&limitFlag, "limit", "n", 20, "maximum number of sessions to display")
	sessionsCmd.Flags().StringVarP(&sessionTypeFlag, "type", "t", "all", "session types to display: 'chat', 'investigation', or 'all'")
	RootCmd.AddCommand(sessionsCmd)
}
