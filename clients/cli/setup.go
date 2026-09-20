package cli

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	wizardTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00E5FF")).
				Padding(0, 1)

	successBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981"))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Bold(true)
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for Niskava Agent environment (.env)",
	Long:  `Launches a step-by-step wizard to configure AI Provider (Gemini/OpenAI/9router), API keys, and local Niskava Agent preferences.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunInteractiveSetup()
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}

// RunInteractiveSetup launches the step-by-step terminal setup wizard.
func RunInteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("=============================================================================")
	fmt.Printf(" %s\n", wizardTitleStyle.Render("NISKAVA AGENT — QUICK SETUP WIZARD (1 MINUTE)"))
	fmt.Println(" Automatic setup for AI Provider connection, Sectors Financial API, & Storage.")
	fmt.Println("=============================================================================")
	fmt.Println()

	// 1. AI Provider Selection
	fmt.Println(accentStyle.Render("Step 1: Select AI Provider for ReAct Research Assistant:"))
	fmt.Println("  [1] Local 9router (http://localhost:20128/v1, Model: hermes) [Recommended]")
	fmt.Println("  [2] Google Gemini Cloud (Google AI Studio, Model: gemini-2.0-flash)")
	fmt.Println("  [3] Custom OpenAI-Compatible (OpenAI, vLLM, Ollama, etc)")
	fmt.Print("\nChoice [1/2/3, default: 1]: ")

	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)
	if providerChoice == "" {
		providerChoice = "1"
	}

	existingKey := os.Getenv("OPENAI_API_KEY")
	if existingKey == "" {
		existingKey = "sk-9router-local-key"
	}

	var (
		aiProvider    = "openai"
		openAIBaseURL = "http://localhost:20128/v1"
		openAIKey     = existingKey
		openAIModel   = "hermes"
		geminiKey     = ""
		geminiModel   = "gemini-2.0-flash"
	)

	switch providerChoice {
	case "2":
		aiProvider = "gemini"
		fmt.Print("\nEnter Google Gemini API Key (https://aistudio.google.com/): ")
		geminiKey, _ = reader.ReadString('\n')
		geminiKey = strings.TrimSpace(geminiKey)

		fmt.Print("Gemini Model Name [default: gemini-2.0-flash]: ")
		m, _ := reader.ReadString('\n')
		m = strings.TrimSpace(m)
		if m != "" {
			geminiModel = m
		}

	case "3":
		aiProvider = "openai"
		fmt.Print("\nEnter Base URL [example: https://api.openai.com/v1]: ")
		openAIBaseURL, _ = reader.ReadString('\n')
		openAIBaseURL = strings.TrimSpace(openAIBaseURL)
		if openAIBaseURL == "" {
			openAIBaseURL = "https://api.openai.com/v1"
		}

		fmt.Print("Enter API Key: ")
		openAIKey, _ = reader.ReadString('\n')
		openAIKey = strings.TrimSpace(openAIKey)

		fmt.Print("Model Name [example: gpt-4o-mini]: ")
		openAIModel, _ = reader.ReadString('\n')
		openAIModel = strings.TrimSpace(openAIModel)
		if openAIModel == "" {
			openAIModel = "gpt-4o-mini"
		}

	default: // 1: 9router
		aiProvider = "openai"
		openAIBaseURL = "http://localhost:20128/v1"
		fmt.Printf("\nBase URL 9router [%s]: ", openAIBaseURL)
		u, _ := reader.ReadString('\n')
		u = strings.TrimSpace(u)
		if u != "" {
			openAIBaseURL = u
		}

		fmt.Printf("9router API Key [%s]: ", openAIKey)
		k, _ := reader.ReadString('\n')
		k = strings.TrimSpace(k)
		if k != "" {
			openAIKey = k
		}

		fmt.Printf("Model Combo [%s]: ", openAIModel)
		m, _ := reader.ReadString('\n')
		m = strings.TrimSpace(m)
		if m != "" {
			openAIModel = m
		}
	}

	// 2. Sectors Financial API Key
	fmt.Println()
	fmt.Println(accentStyle.Render("Step 2: Sectors Financial API v2 (Indonesia Stock Exchange):"))
	fmt.Println("  (Get key at https://sectors.app. Press ENTER for Offline Mock Mode).")
	fmt.Print("Sectors API Key [optional]: ")
	sectorsKey, _ := reader.ReadString('\n')
	sectorsKey = strings.TrimSpace(sectorsKey)

	// 3. Health ping test
	fmt.Println()
	fmt.Println("Testing provider connection...")
	if aiProvider == "openai" && strings.Contains(openAIBaseURL, "localhost:20128") {
		client := http.Client{Timeout: 3 * time.Second}
		req, _ := http.NewRequestWithContext(context.Background(), "GET", openAIBaseURL+"/models", nil)
		req.Header.Set("Authorization", "Bearer "+openAIKey)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			fmt.Println(successBadgeStyle.Render("  [✓] Local 9router connected successfully!"))
		} else {
			fmt.Println("  [!] 9router not active on localhost:20128. Config saved successfully.")
		}
	}

	// 4. Generate .env content
	envContent := fmt.Sprintf(`# =============================================================================
# NISKAVA AGENT — ENVIRONMENT CONFIGURATION (.env)
# Generated automatically via 'niskava setup' on %s
# =============================================================================

# 1. AI MODEL CONFIGURATION (ReAct Agent)
AI_PROVIDER=%s
OPENAI_BASE_URL=%s
OPENAI_API_KEY=%s
OPENAI_MODEL=%s

# Google Gemini (Backup)
GEMINI_API_KEY=%s
GEMINI_MODEL=%s

# 2. SECTORS FINANCIAL API v2 (IDX)
SECTORS_API_KEY=%s
SECTORS_BASE_URL=https://api.sectors.app/v2

# Offline / Mock Mode (Law 5: Credit Conservation)
MOCK_SECTORS=%s
NISKAVA_OFFLINE=%s

# 3. LOCAL STORAGE & ENGINE (Law 4: Local-First SQLite WAL)
NISKAVA_DB_PATH=~/.niskava/niskava.db
NISKAVA_PYTHON_BIN=.venv/bin/python3
NISKAVA_ENGINE_PATH=./backend/engine
NISKAVA_DEFAULT_MARKET=IDX
NISKAVA_PORT=8080

# 4. OSINT HARVESTER
NEWS_HARVEST_MAX_ARTICLES=5
NEWS_TIMEOUT_SECONDS=10
`, time.Now().Format(time.RFC3339),
		aiProvider, openAIBaseURL, openAIKey, openAIModel,
		geminiKey, geminiModel,
		sectorsKey,
		func() string {
			if sectorsKey == "" {
				return "0"
			}
			return "0"
		}(),
		func() string {
			if sectorsKey == "" {
				return "0"
			}
			return "0"
		}(),
	)

	// Write to .env with 0600 permissions
	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	fmt.Println()
	fmt.Println("=============================================================================")
	fmt.Println(successBadgeStyle.Render(" SETUP COMPLETED! Secure .env file generated successfully (0600 permissions)."))
	fmt.Println(" You can launch the interactive interface right away with:")
	fmt.Println("   " + accentStyle.Render("./bin/niskava"))
	fmt.Println("=============================================================================")
	fmt.Println()

	return nil
}
