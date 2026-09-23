package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/telegram"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	portFlag     int
	openFlag     bool
	telegramFlag bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start local web workspace server (REST & SSE streaming)",
	Long: `Starts the Niskava background daemon providing REST endpoints and SSE streaming
for the Web Workspace and external clients on http://localhost:20128.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port := portFlag
		if port == 0 && cfg != nil {
			port = cfg.Server.Port
		}
		if port == 0 {
			port = 20128
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		srv, err := server.Start(ctx, port, appDB)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FCD535")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FFFFFF"))

		dbPath := "~/.niskava/niskava.db"
		if cfg != nil && cfg.Storage.DBPath != "" {
			dbPath = cfg.Storage.DBPath
		}

		serverInfo := tui.TF("serve_online_box", srv.URL, dbPath)
		fmt.Println("\n" + box.Render(serverInfo) + "\n")

		// Optional Telegram Bot Daemon in background
		if (telegramFlag || (cfg != nil && cfg.Telegram.Enabled)) && cfg != nil && cfg.Telegram.BotToken != "" {
			sm := srv.SessionManager
			botSvc, botErr := telegram.NewBotService(cfg, appDB, sm)
			if botErr == nil {
				if startErr := botSvc.Start(); startErr == nil {
					defer botSvc.Stop()
					teleBox := lipgloss.NewStyle().
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("#2AABEE")).
						Padding(0, 1).
						Foreground(lipgloss.Color("#FFFFFF"))
					fmt.Println(teleBox.Render("🤖 Telegram Bot Poller Active in background") + "\n")
				}
			}
		}

		if openFlag {
			fmt.Printf(tui.T("serve_opening_browser"), srv.URL)
			_ = server.OpenBrowser(srv.URL)
		}

		<-sigChan
		fmt.Println(tui.T("serve_stopping"))
		cancel()
		fmt.Println(tui.T("serve_stopped_gracefully"))
		return nil
	},
}

func init() {
	serveCmd.Flags().IntVarP(&portFlag, "port", "p", 20128, "server port (default: 20128)")
	serveCmd.Flags().BoolVarP(&openFlag, "open", "o", false, "open web dashboard in browser automatically")
	serveCmd.Flags().BoolVar(&telegramFlag, "telegram", false, "enable Telegram bot long-poller alongside the web server")
	RootCmd.AddCommand(serveCmd)
}
