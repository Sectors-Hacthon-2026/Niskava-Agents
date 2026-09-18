package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var limitFlag int

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List previous investigation sessions from local SQLite storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		investigations, err := appDB.ListInvestigations(limitFlag)
		if err != nil {
			return fmt.Errorf("failed to retrieve sessions: %w", err)
		}

		if len(investigations) == 0 {
			fmt.Println("Belum ada sesi investigasi yang tersimpan di ~/.niskava/niskava.db.")
			fmt.Println("Jalankan 'niskava investigate <TICKER>' untuk memulai investigasi baru.")
			return nil
		}

		fmt.Println("\nRIWAYAT SESI INVESTIGASI (AUDIT TRAIL):")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("%-20s %-8s %-12s %-20s %s\n", "SESSION ID", "TICKER", "STATUS", "STARTED AT", "SUMMARY")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────")

		for _, inv := range investigations {
			summary := "-"
			if inv.SummaryText != nil && *inv.SummaryText != "" {
				summary = *inv.SummaryText
				if len(summary) > 40 {
					summary = summary[:37] + "..."
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
			if len(dateStr) > 19 {
				dateStr = strings.Replace(dateStr[:19], "T", " ", 1)
			}

			fmt.Printf("%-20s %-8s %-12s %-20s %s\n", inv.ID, inv.Ticker, statusStyled, dateStr, summary)
		}
		fmt.Println("─────────────────────────────────────────────────────────────────────────────")
		return nil
	},
}

func init() {
	sessionsCmd.Flags().IntVarP(&limitFlag, "limit", "l", 20, "maximum number of sessions to display")
	RootCmd.AddCommand(sessionsCmd)
}
