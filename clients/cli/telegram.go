package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
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

var (
	telegramTestChatID  string
	telegramTestMessage string
)

var telegramStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Telegram bot connection status, identity, and whitelist",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			var err error
			cfg, err = config.Load("")
			if err != nil || cfg == nil {
				cfg = config.DefaultConfig()
			}
		}

		token := strings.TrimSpace(cfg.Telegram.BotToken)
		if telegramTokenFlag != "" {
			token = telegramTokenFlag
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2AABEE")).
			Padding(0, 1).
			Width(74)

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2AABEE"))
		okBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorSuccess).Render("[ONLINE / VERIFIED]")
		warnBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorWarning).Render("[NOT CONFIGURED]")
		errBadge := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorDanger).Render("[OFFLINE / ERROR]")

		var sb strings.Builder
		sb.WriteString(titleStyle.Render("🤖 NISKAVA TELEGRAM BOT DIAGNOSTICS") + "\n\n")

		if token == "" {
			sb.WriteString(fmt.Sprintf("Status:       %s\n", warnBadge))
			sb.WriteString("Bot Token:    (empty)\n")
			sb.WriteString("Autostart:    Disabled\n\n")
			sb.WriteString("Run 'niskava setup' to configure your Telegram Bot Token,\n")
			sb.WriteString("or set NISKAVA_TELEGRAM_TOKEN in your environment / .env file.\n")
			fmt.Println("\n" + box.Render(sb.String()) + "\n")
			return nil
		}

		sb.WriteString(fmt.Sprintf("Token:        %s\n", tui.MaskAPIKey(token)))
		autostartStr := "Disabled"
		if cfg.Telegram.Enabled {
			autostartStr = "Enabled (starts with 'niskava serve')"
		}
		sb.WriteString(fmt.Sprintf("Autostart:    %s\n", autostartStr))

		// Live API verification ping
		info, err := FetchTelegramBotInfo(token, nil)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Connection:   %s (%v)\n", errBadge, err))
		} else {
			sb.WriteString(fmt.Sprintf("Connection:   %s\n", okBadge))
			sb.WriteString(fmt.Sprintf("Bot Name:     %s\n", info.FirstName))
			sb.WriteString(fmt.Sprintf("Username:     @%s\n", info.Username))
			sb.WriteString(fmt.Sprintf("Bot ID:       %d\n", info.ID))
		}

		sb.WriteString("\nAllowed Users (Whitelist):\n")
		if len(cfg.Telegram.AllowedUsers) == 0 {
			sb.WriteString("  ⚠️  " + lipgloss.NewStyle().Foreground(tui.ColorWarning).Render("Public Access (No whitelist configured — all users can message bot!)") + "\n")
		} else {
			for i, u := range cfg.Telegram.AllowedUsers {
				sb.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, u))
			}
		}

		sb.WriteString("\n💡 Tip: Check your Telegram User ID using @userinfobot or @RawDataBot in Telegram.\n")
		sb.WriteString("   Manage whitelist: 'niskava telegram user add <id_or_username>'")

		fmt.Println("\n" + box.Render(sb.String()) + "\n")
		return nil
	},
}

var telegramUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users", "whitelist"},
	Short:   "Manage authorized Telegram users whitelist",
}

var telegramUserListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all whitelisted Telegram user IDs and usernames",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			var err error
			cfg, err = config.Load("")
			if err != nil || cfg == nil {
				cfg = config.DefaultConfig()
			}
		}

		fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2AABEE")).Render("📋 Niskava Telegram Whitelist:"))
		if len(cfg.Telegram.AllowedUsers) == 0 {
			fmt.Println("  ⚠️  Whitelist is currently empty. Bot is in OPEN/PUBLIC mode.")
			fmt.Println("     To restrict access, run: niskava telegram user add <user_id_or_username>")
		} else {
			for i, u := range cfg.Telegram.AllowedUsers {
				fmt.Printf("  [%d] %s\n", i+1, u)
			}
			fmt.Printf("\nTotal: %d user(s) authorized.\n", len(cfg.Telegram.AllowedUsers))
		}
		fmt.Println("💡 Tip: To find your numeric Telegram ID, send /start to @userinfobot on Telegram.")
		return nil
	},
}

var telegramUserAddCmd = &cobra.Command{
	Use:   "add <user_id_or_username> [more...]",
	Short: "Add user(s) to the Telegram whitelist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			var err error
			cfg, err = config.Load("")
			if err != nil || cfg == nil {
				cfg = config.DefaultConfig()
			}
		}

		addedCount := 0
		for _, arg := range args {
			added, err := AddTelegramAllowedUser(cfg, arg)
			if err != nil {
				return err
			}
			if added {
				fmt.Printf("  %s Added %q to Telegram whitelist.\n", lipgloss.NewStyle().Foreground(tui.ColorSuccess).Render("[✓]"), CleanTelegramUsernameOrID(arg))
				addedCount++
			} else {
				fmt.Printf("  %s %q is already in the whitelist.\n", lipgloss.NewStyle().Foreground(tui.ColorWarning).Render("[!]"), CleanTelegramUsernameOrID(arg))
			}
		}

		if addedCount > 0 {
			if err := config.SaveConfig(cfg, ""); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("\nConfiguration successfully updated in ~/.niskava/config.yaml and .env\n")
		}
		return nil
	},
}

var telegramUserRemoveCmd = &cobra.Command{
	Use:     "remove <user_id_or_username> [more...]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove user(s) from the Telegram whitelist",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			var err error
			cfg, err = config.Load("")
			if err != nil || cfg == nil {
				cfg = config.DefaultConfig()
			}
		}

		removedCount := 0
		for _, arg := range args {
			removed, err := RemoveTelegramAllowedUser(cfg, arg)
			if err != nil {
				return err
			}
			if removed {
				fmt.Printf("  %s Removed %q from Telegram whitelist.\n", lipgloss.NewStyle().Foreground(tui.ColorSuccess).Render("[✓]"), CleanTelegramUsernameOrID(arg))
				removedCount++
			} else {
				fmt.Printf("  %s %q was not found in the whitelist.\n", lipgloss.NewStyle().Foreground(tui.ColorWarning).Render("[!]"), CleanTelegramUsernameOrID(arg))
			}
		}

		if removedCount > 0 {
			if err := config.SaveConfig(cfg, ""); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("\nConfiguration successfully updated in ~/.niskava/config.yaml and .env\n")
		}
		return nil
	},
}

var telegramTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test message to verify Telegram bot delivery",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			var err error
			cfg, err = config.Load("")
			if err != nil || cfg == nil {
				cfg = config.DefaultConfig()
			}
		}

		token := strings.TrimSpace(cfg.Telegram.BotToken)
		if telegramTokenFlag != "" {
			token = telegramTokenFlag
		}
		if token == "" {
			return fmt.Errorf("telegram bot token is not configured. Provide via --token flag, NISKAVA_TELEGRAM_TOKEN env, or 'niskava setup'")
		}

		if strings.TrimSpace(telegramTestChatID) == "" {
			return fmt.Errorf("--chat-id is required. Run: niskava telegram test --chat-id <YOUR_CHAT_ID>")
		}

		fmt.Printf("Sending test message to chat ID %s...\n", telegramTestChatID)
		if err := SendTelegramTestMessage(token, telegramTestChatID, telegramTestMessage, nil); err != nil {
			return fmt.Errorf("failed to send test message: %w", err)
		}

		fmt.Printf("%s Test message delivered successfully!\n", lipgloss.NewStyle().Foreground(tui.ColorSuccess).Render("[✓]"))
		return nil
	},
}

// SendTelegramTestMessage sends an explicit test message to the specified Telegram chat ID.
func SendTelegramTestMessage(token string, chatID string, message string, client *http.Client) error {
	cleanToken := strings.TrimSpace(token)
	if cleanToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	cleanChatID := strings.TrimSpace(chatID)
	if cleanChatID == "" {
		return fmt.Errorf("chat ID is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	payload := map[string]interface{}{
		"chat_id": cleanChatID,
		"text":    message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cleanToken)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("network error sending test message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("invalid response from Telegram API: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}
	return nil
}

func init() {
	telegramCmd.Flags().StringVar(&telegramTokenFlag, "token", "", "Telegram Bot Token from @BotFather (overrides config)")

	telegramUserCmd.AddCommand(telegramUserListCmd)
	telegramUserCmd.AddCommand(telegramUserAddCmd)
	telegramUserCmd.AddCommand(telegramUserRemoveCmd)

	telegramTestCmd.Flags().StringVar(&telegramTestChatID, "chat-id", "", "Target Telegram Chat ID to send test message")
	telegramTestCmd.Flags().StringVar(&telegramTestMessage, "message", "🔔 Niskava Agent: Koneksi Telegram Bot berhasil diverifikasi.", "Custom test message text")

	telegramCmd.AddCommand(telegramStatusCmd)
	telegramCmd.AddCommand(telegramUserCmd)
	telegramCmd.AddCommand(telegramTestCmd)

	RootCmd.AddCommand(telegramCmd)
}

// TelegramBotInfo represents information about the verified bot from Telegram API getMe.
type TelegramBotInfo struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type telegramGetMeResponse struct {
	OK          bool            `json:"ok"`
	Result      TelegramBotInfo `json:"result"`
	Description string          `json:"description"`
}

// FetchTelegramBotInfo connects to Telegram API to fetch bot info for verification.
func FetchTelegramBotInfo(token string, client *http.Client) (*TelegramBotInfo, error) {
	return fetchTelegramBotInfoWithBaseURL(token, client, "https://api.telegram.org")
}

func fetchTelegramBotInfoWithBaseURL(token string, client *http.Client, baseURL string) (*TelegramBotInfo, error) {
	cleanToken := strings.TrimSpace(token)
	if cleanToken == "" {
		return nil, fmt.Errorf("telegram token is empty")
	}
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	url := fmt.Sprintf("%s/bot%s/getMe", strings.TrimRight(baseURL, "/"), cleanToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error connecting to Telegram API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp telegramGetMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("invalid response from Telegram API: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API error: %s", apiResp.Description)
	}
	return &apiResp.Result, nil
}

// CleanTelegramUsernameOrID normalizes telegram identifier by trimming spaces and stripping leading @.
func CleanTelegramUsernameOrID(val string) string {
	cleaned := strings.TrimSpace(val)
	cleaned = strings.TrimPrefix(cleaned, "@")
	return cleaned
}

// AddTelegramAllowedUser appends a new user to the allowed users whitelist if not already present.
func AddTelegramAllowedUser(c *config.Config, user string) (bool, error) {
	if c == nil {
		return false, fmt.Errorf("configuration is nil")
	}
	cleaned := CleanTelegramUsernameOrID(user)
	if cleaned == "" {
		return false, fmt.Errorf("user identifier cannot be empty")
	}
	for _, u := range c.Telegram.AllowedUsers {
		if strings.EqualFold(strings.TrimSpace(u), cleaned) {
			return false, nil // Already exists
		}
	}
	c.Telegram.AllowedUsers = append(c.Telegram.AllowedUsers, cleaned)
	return true, nil
}

// RemoveTelegramAllowedUser deletes a user from the allowed users whitelist.
func RemoveTelegramAllowedUser(c *config.Config, user string) (bool, error) {
	if c == nil {
		return false, fmt.Errorf("configuration is nil")
	}
	cleaned := CleanTelegramUsernameOrID(user)
	if cleaned == "" {
		return false, fmt.Errorf("user identifier cannot be empty")
	}
	idx := -1
	for i, u := range c.Telegram.AllowedUsers {
		if strings.EqualFold(strings.TrimSpace(u), cleaned) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false, nil // Not found
	}
	c.Telegram.AllowedUsers = append(c.Telegram.AllowedUsers[:idx], c.Telegram.AllowedUsers[idx+1:]...)
	return true, nil
}
