package cli

import (
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
	telegramTokenFlag string
)

var telegramCmd = &cobra.Command{
	Use:     "telegram",
	Aliases: []string{"bot"},
	Short:   "Start Telegram conversational bot runner",
	Long: `Starts the Niskava Agent Telegram Bot using long-polling.
Incoming messages are routed to the autonomous ReAct investigation pipeline
and synchronized to the local SQLite database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if telegramTokenFlag != "" && cfg != nil {
			cfg.Telegram.BotToken = telegramTokenFlag
		}

		if cfg == nil || cfg.Telegram.BotToken == "" {
			return fmt.Errorf("telegram bot token is required. Provide via --token flag, NISKAVA_TELEGRAM_TOKEN env, or ~/.niskava/config.yaml")
		}

		sm := server.NewSessionManager()
		botSvc, err := telegram.NewBotService(cfg, appDB, sm)
		if err != nil {
			return fmt.Errorf("failed to initialize telegram bot: %w", err)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		if err := botSvc.Start(); err != nil {
			return fmt.Errorf("failed to start telegram bot service: %w", err)
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2AABEE")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FFFFFF"))

		botUsername := "niskava_agent_bot"
		if botSvc.Bot() != nil && botSvc.Bot().Me != nil && botSvc.Bot().Me.Username != "" {
			botUsername = botSvc.Bot().Me.Username
		}

		dbPath := "~/.niskava/niskava.db"
		if cfg != nil && cfg.Storage.DBPath != "" {
			dbPath = cfg.Storage.DBPath
		}

		infoText := fmt.Sprintf("🤖 Niskava Telegram Bot Online\nBot: @%s\nDB:  %s\nMode: Conversational ReAct (Typing Status Active)\nPress Ctrl+C to stop.", botUsername, dbPath)
		fmt.Println("\n" + box.Render(infoText) + "\n")

		<-sigChan
		fmt.Println("\n" + tui.T("serve_stopping"))
		botSvc.Stop()
		fmt.Println(tui.T("serve_stopped_gracefully"))
		return nil
	},
}

func init() {
	telegramCmd.Flags().StringVar(&telegramTokenFlag, "token", "", "Telegram Bot Token from @BotFather (overrides config)")
	RootCmd.AddCommand(telegramCmd)
}
