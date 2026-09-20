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
	Short: "Wizard interaktif konfigurasi environment Niskava Agent (.env)",
	Long:  `Menjalankan panduan langkah demi langkah untuk mengatur AI Provider (9router/Gemini/OpenAI), API key, dan preferensi lokal Niskava Agent.`,
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
	fmt.Printf(" %s\n", wizardTitleStyle.Render("NISKAVA AGENT — QUICK SETUP WIZARD (1 MENIT)"))
	fmt.Println(" Konfigurasi otomatis koneksi AI Provider, Sectors Financial API, dan Storage.")
	fmt.Println("=============================================================================")
	fmt.Println()

	// 1. AI Provider Selection
	fmt.Println(accentStyle.Render("Langkah 1: Pilih AI Provider untuk ReAct Research Assistant:"))
	fmt.Println("  [1] 9router Lokal (http://localhost:20128/v1, Model: hermes) [Rekomendasi]")
	fmt.Println("  [2] Google Gemini Cloud (Google AI Studio, Model: gemini-3.6-flash)")
	fmt.Println("  [3] Custom OpenAI-Compatible (OpenAI, vLLM, Ollama, dll)")
	fmt.Print("\nPilihan [1/2/3, default: 1]: ")

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
		geminiModel   = "gemini-3.6-flash"
	)

	switch providerChoice {
	case "2":
		aiProvider = "gemini"
		fmt.Print("\nMasukkan Google Gemini API Key (https://aistudio.google.com/): ")
		geminiKey, _ = reader.ReadString('\n')
		geminiKey = strings.TrimSpace(geminiKey)

		fmt.Print("Model Gemini [default: gemini-3.6-flash]: ")
		m, _ := reader.ReadString('\n')
		m = strings.TrimSpace(m)
		if m != "" {
			geminiModel = m
		}

	case "3":
		aiProvider = "openai"
		fmt.Print("\nMasukkan Base URL [contoh: https://api.openai.com/v1]: ")
		openAIBaseURL, _ = reader.ReadString('\n')
		openAIBaseURL = strings.TrimSpace(openAIBaseURL)
		if openAIBaseURL == "" {
			openAIBaseURL = "https://api.openai.com/v1"
		}

		fmt.Print("Masukkan API Key: ")
		openAIKey, _ = reader.ReadString('\n')
		openAIKey = strings.TrimSpace(openAIKey)

		fmt.Print("Model Name [contoh: gpt-4o-mini]: ")
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
			if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
				openAIBaseURL = "http://localhost:20128/v1"
			} else {
				openAIBaseURL = u
			}
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
	fmt.Println(accentStyle.Render("Langkah 2: Sectors Financial API v2 (Bursa Efek Indonesia):"))
	fmt.Println("  (Dapatkan key di https://sectors.app. Tekan ENTER untuk mode Offline Mock).")
	fmt.Print("Sectors API Key [opsional]: ")
	sectorsKey, _ := reader.ReadString('\n')
	sectorsKey = strings.TrimSpace(sectorsKey)

	// 3. Health ping test
	fmt.Println()
	fmt.Println("Menguji koneksi provider...")
	if aiProvider == "openai" && strings.Contains(openAIBaseURL, "localhost:20128") {
		client := http.Client{Timeout: 3 * time.Second}
		req, _ := http.NewRequestWithContext(context.Background(), "GET", openAIBaseURL+"/models", nil)
		req.Header.Set("Authorization", "Bearer "+openAIKey)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			fmt.Println(successBadgeStyle.Render("  [✓] 9router lokal terhubung dengan baik!"))
		} else {
			fmt.Println("  [!] 9router belum aktif di localhost:20128. Konfigurasi tetap disimpan.")
		}
	}

	// 4. Generate .env content
	envContent := fmt.Sprintf(`# =============================================================================
# NISKAVA AGENT — ENVIRONMENT CONFIGURATION (.env)
# Disusun otomatis via 'niskava setup' pada %s
# =============================================================================

# 1. AI MODEL CONFIGURATION (ReAct Agent)
AI_PROVIDER=%s
OPENAI_BASE_URL=%s
OPENAI_API_KEY=%s
OPENAI_MODEL=%s

# Google Gemini (Cadangan)
GEMINI_API_KEY=%s
GEMINI_MODEL=%s

# 2. SECTORS FINANCIAL API v2 (IDX)
SECTORS_API_KEY=%s
SECTORS_BASE_URL=https://api.sectors.app/v2

# Mode Offline / Mock (Hukum 5: Hemat Kredit)
MOCK_SECTORS=%s
NISKAVA_OFFLINE=%s

# 3. LOCAL STORAGE & ENGINE (Hukum 4: Local-First SQLite WAL)
NISKAVA_DB_PATH=~/.niskava/niskava.db
NISKAVA_PYTHON_BIN=.venv/bin/python3
NISKAVA_ENGINE_PATH=./engine
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
		return fmt.Errorf("gagal menyimpan file .env: %w", err)
	}

	fmt.Println()
	fmt.Println("=============================================================================")
	fmt.Println(successBadgeStyle.Render(" SETUP SELESAI! File .env berhasil dibuat dengan aman (izin 0600)."))
	fmt.Println(" Anda dapat langsung memulai antarmuka interaktif dengan:")
	fmt.Println("   " + accentStyle.Render("./bin/niskava"))
	fmt.Println("=============================================================================")
	fmt.Println()

	return nil
}
