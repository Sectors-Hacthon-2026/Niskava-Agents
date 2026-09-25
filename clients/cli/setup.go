package cli

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	wizardTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(tui.ColorBg).
				Background(tui.ColorAccent).
				Padding(0, 1)

	wizardStepStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.ColorAccent)

	wizardItemBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(tui.ColorAccent)

	wizardMutedStyle = lipgloss.NewStyle().
				Foreground(tui.ColorMuted)

	wizardSuccessBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(tui.ColorSuccess)

	wizardWarnBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(tui.ColorWarning)

	wizardErrBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(tui.ColorDanger)

	setupCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tui.ColorAccent).
			Width(78).
			Padding(0, 1).
			Foreground(tui.ColorFg)
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for Niskava Agent environment (.env & config.yaml)",
	Long:  `Launches a dynamic step-by-step wizard to configure AI Provider (OpenRouter/Gemini/OpenAI/DeepSeek/Groq/Ollama/9router), API keys, and local Niskava Agent preferences.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setup wizard initializes configuration from scratch; does not require initial DB
		if cfg == nil {
			cfg, _ = config.Load(cfgFile)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunInteractiveSetup()
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}

// SetupParams encapsulates all configuration attributes gathered during the wizard.
type SetupParams struct {
	AIProvider     string
	OpenAIBaseURL  string
	OpenAIKey      string
	OpenAIModel    string
	GeminiKey      string
	GeminiModel    string
	SectorsKey     string
	PythonBin      string
	LLMTimeoutSecs float64 // Inference timeout in seconds (10–300). 0 → defaults to 60.
}

// ProviderPreset holds standard endpoints and defaults for popular AI providers.
type ProviderPreset struct {
	ID             string
	Name           string
	BaseURL        string
	DefaultModel   string
	KeyPlaceholder string
	Description    string
}

// GetProviderPresets returns supported pre-configured AI providers and gateways.
func GetProviderPresets() map[string]ProviderPreset {
	return map[string]ProviderPreset{
		"openrouter": {
			ID:             "openrouter",
			Name:           "OpenRouter AI Gateway",
			BaseURL:        "https://openrouter.ai/api/v1",
			DefaultModel:   "deepseek/deepseek-chat",
			KeyPlaceholder: "sk-or-v1-...",
			Description:    "Universal router for 200+ models (Claude, GPT, DeepSeek, Gemini, etc.)",
		},
		"gemini": {
			ID:             "gemini",
			Name:           "Google Gemini Cloud (AI Studio)",
			BaseURL:        "https://generativelanguage.googleapis.com/v1beta/openai",
			DefaultModel:   "gemini-2.0-flash",
			KeyPlaceholder: "AIzaSy...",
			Description:    "Direct Google AI Studio API via standard OpenAI-compatible protocol",
		},
		"openai": {
			ID:             "openai",
			Name:           "OpenAI Official API",
			BaseURL:        "https://api.openai.com/v1",
			DefaultModel:   "gpt-4o-mini",
			KeyPlaceholder: "sk-proj-...",
			Description:    "Direct OpenAI Cloud Platform",
		},
		"deepseek": {
			ID:             "deepseek",
			Name:           "DeepSeek Cloud API",
			BaseURL:        "https://api.deepseek.com/v1",
			DefaultModel:   "deepseek-chat",
			KeyPlaceholder: "sk-...",
			Description:    "Direct DeepSeek official API (DeepSeek-V3 / R1)",
		},
		"groq": {
			ID:             "groq",
			Name:           "Groq High-Speed Cloud",
			BaseURL:        "https://api.groq.com/openai/v1",
			DefaultModel:   "llama-3.3-70b-versatile",
			KeyPlaceholder: "gsk_...",
			Description:    "Ultra-low latency Llama 3 inference on LPU",
		},
		"ollama": {
			ID:             "ollama",
			Name:           "Ollama Local Model",
			BaseURL:        "http://localhost:11434/v1",
			DefaultModel:   "llama3.2",
			KeyPlaceholder: "none (local)",
			Description:    "Privacy-first offline local models",
		},
		"router_local": {
			ID:             "router_local",
			Name:           "Local Router (9router / LiteLLM)",
			BaseURL:        "http://localhost:20128/v1",
			DefaultModel:   "hermes",
			KeyPlaceholder: "sk-...",
			Description:    "Local proxy running on port 20128 or custom",
		},
	}
}

// BuildEnvContent formats complete .env file content honoring Law 5 (offline mock mode).
func BuildEnvContent(p SetupParams) string {
	mockVal := "0"
	if strings.TrimSpace(p.SectorsKey) == "" {
		mockVal = "1"
	}

	pyBin := p.PythonBin
	if pyBin == "" {
		pyBin = "python3"
		if runtime.GOOS == "windows" {
			pyBin = "python"
		}
	}

	timeoutSecs := p.LLMTimeoutSecs
	if timeoutSecs <= 0 {
		timeoutSecs = 60.0
	}
	if timeoutSecs < 10.0 {
		timeoutSecs = 10.0
	}
	if timeoutSecs > 300.0 {
		timeoutSecs = 300.0
	}

	return fmt.Sprintf(`# =============================================================================
# NISKAVA AGENT — ENVIRONMENT CONFIGURATION (.env)
# Generated automatically via 'niskava setup' on %s
# =============================================================================

# 1. AI MODEL CONFIGURATION (ReAct Agent)
AI_PROVIDER=%s
OPENAI_BASE_URL=%s
OPENAI_API_KEY=%s
OPENAI_MODEL=%s

# Google Gemini (Direct / Backup)
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
NISKAVA_PYTHON_BIN=%s
NISKAVA_ENGINE_PATH=./backend/engine
NISKAVA_DEFAULT_MARKET=IDX
NISKAVA_PORT=8080

# 4. SECTORS NEWS ENGINE
NEWS_HARVEST_MAX_ARTICLES=5
NEWS_TIMEOUT_SECONDS=10

# 5. INFERENCE TIMEOUT — how long to wait for LLM response (seconds, range 10–300)
# Profiles: fast=25 | balanced=60 | deep=120 | local_llm=180
NISKAVA_LLM_TIMEOUT=%.2f
`, time.Now().Format(time.RFC3339),
		p.AIProvider, p.OpenAIBaseURL, p.OpenAIKey, p.OpenAIModel,
		p.GeminiKey, p.GeminiModel,
		p.SectorsKey,
		mockVal, mockVal,
		pyBin,
		timeoutSecs,
	)
}

// SaveSetupConfiguration writes both local .env and global ~/.niskava/config.yaml with 0600 permissions.
func SaveSetupConfiguration(p SetupParams) error {
	envContent := BuildEnvContent(p)
	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	// Persist to ~/.niskava/config.yaml for CLI-Web synchronization
	cfg, err := config.Load("")
	if err != nil || cfg == nil {
		cfg = config.DefaultConfig()
	}
	cfg.Auth.AIProvider = p.AIProvider
	cfg.Auth.OpenAIBaseURL = p.OpenAIBaseURL
	cfg.Auth.OpenAIAPIKey = p.OpenAIKey
	cfg.Auth.OpenAIModel = p.OpenAIModel
	cfg.Auth.GeminiAPIKey = p.GeminiKey
	cfg.Auth.GeminiModel = p.GeminiModel
	cfg.Auth.SectorsAPIKey = p.SectorsKey
	cfg.Preferences.OfflineMode = (strings.TrimSpace(p.SectorsKey) == "")
	if p.LLMTimeoutSecs > 0 {
		cfg.Preferences.LLMTimeoutSecs = p.LLMTimeoutSecs
	} else {
		cfg.Preferences.LLMTimeoutSecs = 60.0
	}
	if p.PythonBin != "" {
		cfg.Engine.PythonBin = p.PythonBin
	}

	_ = config.SaveConfig(cfg)
	return nil
}

// TestLiveConnection performs a lightweight HTTP ping against target provider or sectors endpoint.
func TestLiveConnection(ctx context.Context, target, baseURL, apiKey string) (bool, string, time.Duration) {
	start := time.Now()
	client := &http.Client{Timeout: 4 * time.Second}

	target = strings.ToLower(strings.TrimSpace(target))
	if target == "sectors" {
		if apiKey == "" {
			return true, "Offline mock mode active (no network ping needed)", 0
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.sectors.app/v2/daily/BBCA/?format=json", nil)
		if err != nil {
			return false, fmt.Sprintf("Build request error: %v", err), time.Since(start)
		}
		req.Header.Set("Authorization", apiKey)
		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Sprintf("Connection failed: %v", err), time.Since(start)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true, "Connected to Sectors v2 API successfully", time.Since(start)
		}
		return false, fmt.Sprintf("Sectors API returned HTTP %d", resp.StatusCode), time.Since(start)
	}

	// AI Provider / Router ping: standard /models endpoint
	cleanBase := strings.TrimRight(baseURL, "/")
	if cleanBase == "" {
		cleanBase = "https://api.openai.com/v1"
	}
	testURL := cleanBase + "/models"
	if target == "gemini" && apiKey != "" && !strings.Contains(cleanBase, "googleapis.com") {
		testURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return false, fmt.Sprintf("Build request error: %v", err), time.Since(start)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Endpoint unreachable (%v)", err), time.Since(start)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	if resp.StatusCode == http.StatusOK {
		return true, fmt.Sprintf("HTTP 200 OK (Latency: %dms)", elapsed.Milliseconds()), elapsed
	} else if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, "Authentication failed (Check API Key)", elapsed
	}
	return false, fmt.Sprintf("Endpoint returned HTTP %d", resp.StatusCode), elapsed
}

// DetectPythonEnvironment checks for .venv or system python3 and verifies runtime compatibility.
func DetectPythonEnvironment() (string, string, bool) {
	venvCandidates := []string{
		filepath.Join(".venv", "bin", "python3"),
		filepath.Join(".venv", "bin", "python"),
		filepath.Join(".venv", "Scripts", "python.exe"),
		filepath.Join("venv", "bin", "python3"),
		filepath.Join("venv", "bin", "python"),
		filepath.Join("venv", "Scripts", "python.exe"),
		filepath.Join("backend", "engine", ".venv", "bin", "python3"),
		filepath.Join("backend", "engine", ".venv", "bin", "python"),
		filepath.Join("backend", "engine", ".venv", "Scripts", "python.exe"),
		filepath.Join("backend", "engine", "venv", "bin", "python3"),
		filepath.Join("backend", "engine", "venv", "bin", "python"),
		filepath.Join("backend", "engine", "venv", "Scripts", "python.exe"),
	}

	for _, cand := range venvCandidates {
		if _, err := os.Stat(cand); err == nil {
			// Test if numpy is available in venv
			cmd := exec.Command(cand, "-c", "import numpy; print('ok')")
			if out, err := cmd.Output(); err == nil && strings.TrimSpace(string(out)) == "ok" {
				return cand, "Virtual environment (.venv) active with quantitative dependencies", true
			}
			return cand, "Virtual environment (.venv) found, but dependencies may need: pip install -r backend/engine/requirements.txt", false
		}
	}

	// Fallback to system python3 / python
	sysLookups := []string{"python3", "python"}
	for _, name := range sysLookups {
		if path, err := exec.LookPath(name); err == nil {
			cmd := exec.Command(path, "--version")
			if out, err := cmd.CombinedOutput(); err == nil {
				verStr := strings.TrimSpace(string(out))
				return path, fmt.Sprintf("System %s detected (%s). Recommendation: run 'make venv' for isolated runtime", name, verStr), true
			}
			return path, fmt.Sprintf("System %s detected", name), true
		}
	}

	return "python3", "Python interpreter not found on PATH. Please install Python 3.11+", false
}

// BootstrapPythonEnvironment creates a virtual environment in rootDir/.venv and installs requirements.txt.
func BootstrapPythonEnvironment(rootDir, sysPython string) (string, error) {
	if sysPython == "" || sysPython == "python3" || sysPython == "python" {
		if path, err := exec.LookPath("python3"); err == nil {
			sysPython = path
		} else if path, err := exec.LookPath("python"); err == nil {
			sysPython = path
		} else {
			return "", fmt.Errorf("system Python interpreter not found on PATH (requires Python 3.11+)")
		}
	} else {
		if _, err := os.Stat(sysPython); err != nil {
			if _, err := exec.LookPath(sysPython); err != nil {
				return "", fmt.Errorf("python interpreter %q not found: %w", sysPython, err)
			}
		}
	}

	venvDir := filepath.Join(rootDir, ".venv")
	cmdVenv := exec.Command(sysPython, "-m", "venv", venvDir)
	if out, err := cmdVenv.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create virtual environment: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	// Locate pip inside the new venv
	pipCandidates := []string{
		filepath.Join(venvDir, "Scripts", "pip.exe"),
		filepath.Join(venvDir, "Scripts", "pip"),
		filepath.Join(venvDir, "bin", "pip3"),
		filepath.Join(venvDir, "bin", "pip"),
	}
	var pipBin string
	for _, cand := range pipCandidates {
		if _, err := os.Stat(cand); err == nil {
			pipBin = cand
			break
		}
	}
	if pipBin == "" {
		return "", fmt.Errorf("pip not found in newly created virtualenv at %s", venvDir)
	}

	// Upgrade pip (best effort)
	_ = exec.Command(pipBin, "install", "--upgrade", "pip").Run()

	// Locate requirements.txt
	reqCandidates := []string{
		filepath.Join(rootDir, "backend", "engine", "requirements.txt"),
		filepath.Join(rootDir, "requirements.txt"),
	}
	var reqPath string
	for _, cand := range reqCandidates {
		if _, err := os.Stat(cand); err == nil {
			reqPath = cand
			break
		}
	}
	if reqPath == "" {
		return "", fmt.Errorf("backend/engine/requirements.txt not found in %s", rootDir)
	}

	cmdPip := exec.Command(pipBin, "install", "-r", reqPath)
	if out, err := cmdPip.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to install requirements via pip: %s (%w)", strings.TrimSpace(string(out)), err)
	}

	// Locate python binary in new venv
	pyCandidates := []string{
		filepath.Join(venvDir, "Scripts", "python.exe"),
		filepath.Join(venvDir, "bin", "python3"),
		filepath.Join(venvDir, "bin", "python"),
	}
	for _, cand := range pyCandidates {
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	return "", fmt.Errorf("python binary not found in virtualenv at %s", venvDir)
}

// RunInteractiveSetup launches the step-by-step terminal setup wizard.
func RunInteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)
	presets := GetProviderPresets()

	fmt.Println()
	var headerBox strings.Builder
	headerBox.WriteString(wizardTitleStyle.Render("NISKAVA AGENT — DYNAMIC SETUP WIZARD (MULTI-PROVIDER)") + "\n")
	headerBox.WriteString(wizardMutedStyle.Render("Universal setup for External Routers (OpenRouter/9router), Cloud APIs, & Sectors v2."))
	fmt.Println(setupCardStyle.Render(headerBox.String()))
	fmt.Println()

	// Step 1: Select AI Inference Provider Category
	fmt.Println(wizardStepStyle.Render("Step 1: Select AI Model Inference Category:"))
	fmt.Printf("  %s %s %s\n",
		wizardItemBadgeStyle.Render("[1]"),
		"External AI Router / Gateway (OpenRouter, LiteLLM, 9router, Custom Gateway)",
		wizardMutedStyle.Render("[Recommended for Flexibility]"))
	fmt.Printf("  %s %s\n",
		wizardItemBadgeStyle.Render("[2]"),
		"Direct Cloud AI API (Google Gemini, OpenAI, DeepSeek, Groq)")
	fmt.Printf("  %s %s\n",
		wizardItemBadgeStyle.Render("[3]"),
		"Local / Private Offline Model (Ollama, LM Studio, vLLM)")
	fmt.Printf("  %s %s\n",
		wizardItemBadgeStyle.Render("[4]"),
		"Custom Any OpenAI-Compatible (Fully Manual Endpoint)")
	fmt.Printf("\n%s Choice [1/2/3/4, default: 1]: ", wizardStepStyle.Render("►"))

	catChoice, _ := reader.ReadString('\n')
	catChoice = strings.TrimSpace(catChoice)
	if catChoice == "" {
		catChoice = "1"
	}

	var (
		aiProvider    = "openai"
		openAIBaseURL = "https://openrouter.ai/api/v1"
		openAIKey     = ""
		openAIModel   = "deepseek/deepseek-chat"
		geminiKey     = ""
		geminiModel   = "gemini-2.0-flash"
		testTarget    = "openai"
	)

	switch catChoice {
	case "2": // Direct Cloud AI
		fmt.Println()
		fmt.Println(wizardStepStyle.Render("  Choose Cloud AI Provider:"))
		fmt.Printf("    %s Google Gemini Cloud (Google AI Studio) %s\n", wizardItemBadgeStyle.Render("[1]"), wizardMutedStyle.Render("[Fast & High Token Limit]"))
		fmt.Printf("    %s DeepSeek Official Cloud (DeepSeek-V3 / R1)\n", wizardItemBadgeStyle.Render("[2]"))
		fmt.Printf("    %s Groq High-Speed Cloud (Llama 3.3 70B Versatile)\n", wizardItemBadgeStyle.Render("[3]"))
		fmt.Printf("    %s OpenAI Official Platform (GPT-4o mini)\n", wizardItemBadgeStyle.Render("[4]"))
		fmt.Printf("  %s Sub-choice [1/2/3/4, default: 1]: ", wizardStepStyle.Render("►"))

		cloudSub, _ := reader.ReadString('\n')
		cloudSub = strings.TrimSpace(cloudSub)
		if cloudSub == "" || cloudSub == "1" {
			// Gemini
			preset := presets["gemini"]
			aiProvider = "gemini"
			testTarget = "gemini"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel
			geminiModel = preset.DefaultModel

			fmt.Printf("\n  %s Enter Google Gemini API Key (https://aistudio.google.com/): ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			geminiKey = strings.TrimSpace(k)
			openAIKey = geminiKey

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				geminiModel = m
				openAIModel = m
			}
		} else if cloudSub == "2" {
			// DeepSeek
			preset := presets["deepseek"]
			aiProvider = "openai"
			testTarget = "deepseek"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel

			fmt.Printf("\n  %s Enter DeepSeek API Key (https://platform.deepseek.com/): ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			openAIKey = strings.TrimSpace(k)

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		} else if cloudSub == "3" {
			// Groq
			preset := presets["groq"]
			aiProvider = "openai"
			testTarget = "groq"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel

			fmt.Printf("\n  %s Enter Groq API Key (https://console.groq.com/): ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			openAIKey = strings.TrimSpace(k)

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		} else {
			// OpenAI
			preset := presets["openai"]
			aiProvider = "openai"
			testTarget = "openai"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel

			fmt.Printf("\n  %s Enter OpenAI API Key: ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			openAIKey = strings.TrimSpace(k)

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		}

	case "3": // Local / Private Model
		fmt.Println()
		fmt.Println(wizardStepStyle.Render("  Choose Local Inference Engine:"))
		fmt.Printf("    %s Ollama Local (http://localhost:11434/v1) %s\n", wizardItemBadgeStyle.Render("[1]"), wizardMutedStyle.Render("[Recommended for Local]"))
		fmt.Printf("    %s LM Studio / vLLM / LocalAI (http://localhost:1234/v1)\n", wizardItemBadgeStyle.Render("[2]"))
		fmt.Printf("  %s Sub-choice [1/2, default: 1]: ", wizardStepStyle.Render("►"))

		localSub, _ := reader.ReadString('\n')
		localSub = strings.TrimSpace(localSub)
		if localSub == "" || localSub == "1" {
			preset := presets["ollama"]
			aiProvider = "openai"
			testTarget = "ollama"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel
			openAIKey = "ollama-local"

			fmt.Printf("  %s Ollama Base URL [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.BaseURL))
			u, _ := reader.ReadString('\n')
			u = strings.TrimSpace(u)
			if u != "" {
				openAIBaseURL = u
			}

			fmt.Printf("  %s Model Tag [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		} else {
			aiProvider = "openai"
			testTarget = "openai"
			openAIBaseURL = "http://localhost:1234/v1"
			openAIModel = "local-model"
			openAIKey = "not-needed"

			fmt.Printf("  %s Base URL [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(openAIBaseURL))
			u, _ := reader.ReadString('\n')
			u = strings.TrimSpace(u)
			if u != "" {
				openAIBaseURL = u
			}

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(openAIModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		}

	case "4": // Custom Manual Endpoint
		aiProvider = "openai"
		testTarget = "openai"
		fmt.Printf("\n  %s Enter Base URL [example: https://my-custom-proxy.com/v1]: ", wizardStepStyle.Render("►"))
		u, _ := reader.ReadString('\n')
		u = strings.TrimSpace(u)
		if u != "" {
			openAIBaseURL = u
		}

		fmt.Printf("  %s Enter API Key: ", wizardStepStyle.Render("►"))
		k, _ := reader.ReadString('\n')
		openAIKey = strings.TrimSpace(k)

		fmt.Printf("  %s Enter Model Name: ", wizardStepStyle.Render("►"))
		m, _ := reader.ReadString('\n')
		m = strings.TrimSpace(m)
		if m != "" {
			openAIModel = m
		} else {
			openAIModel = "default-model"
		}

	default: // 1: External AI Router
		fmt.Println()
		fmt.Println(wizardStepStyle.Render("  Choose External Router / Proxy:"))
		fmt.Printf("    %s OpenRouter AI Gateway (200+ models: Claude, DeepSeek, Gemini, etc.) %s\n",
			wizardItemBadgeStyle.Render("[1]"), wizardMutedStyle.Render("[Recommended]"))
		fmt.Printf("    %s Local Gateway (9router / LiteLLM on http://localhost:20128/v1)\n",
			wizardItemBadgeStyle.Render("[2]"))
		fmt.Printf("    %s Custom AI Gateway / Proxy\n",
			wizardItemBadgeStyle.Render("[3]"))
		fmt.Printf("  %s Sub-choice [1/2/3, default: 1]: ", wizardStepStyle.Render("►"))

		routerSub, _ := reader.ReadString('\n')
		routerSub = strings.TrimSpace(routerSub)

		if routerSub == "" || routerSub == "1" {
			// OpenRouter
			preset := presets["openrouter"]
			aiProvider = "openai"
			testTarget = "openrouter"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel

			fmt.Printf("\n  %s Enter OpenRouter API Key (https://openrouter.ai/keys): ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			openAIKey = strings.TrimSpace(k)

			fmt.Printf("  %s Model String [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		} else if routerSub == "2" {
			// 9router / local proxy
			preset := presets["router_local"]
			aiProvider = "openai"
			testTarget = "router_local"
			openAIBaseURL = preset.BaseURL
			openAIModel = preset.DefaultModel
			openAIKey = "sk-9router-local-key"

			fmt.Printf("\n  %s Local Router Base URL [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.BaseURL))
			u, _ := reader.ReadString('\n')
			u = strings.TrimSpace(u)
			if u != "" {
				openAIBaseURL = u
			}

			fmt.Printf("  %s Router API Key [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(openAIKey))
			k, _ := reader.ReadString('\n')
			k = strings.TrimSpace(k)
			if k != "" {
				openAIKey = k
			}

			fmt.Printf("  %s Model Name [default: %s]: ", wizardStepStyle.Render("►"), wizardMutedStyle.Render(preset.DefaultModel))
			m, _ := reader.ReadString('\n')
			m = strings.TrimSpace(m)
			if m != "" {
				openAIModel = m
			}
		} else {
			aiProvider = "openai"
			testTarget = "openai"
			fmt.Printf("\n  %s Enter Router Base URL: ", wizardStepStyle.Render("►"))
			u, _ := reader.ReadString('\n')
			openAIBaseURL = strings.TrimSpace(u)

			fmt.Printf("  %s Enter Router API Key: ", wizardStepStyle.Render("►"))
			k, _ := reader.ReadString('\n')
			openAIKey = strings.TrimSpace(k)

			fmt.Printf("  %s Enter Model Name: ", wizardStepStyle.Render("►"))
			m, _ := reader.ReadString('\n')
			openAIModel = strings.TrimSpace(m)
		}
	}

	// Step 2: Sectors Financial API Key (IDX)
	fmt.Println()
	fmt.Println(wizardStepStyle.Render("Step 2: Sectors Financial API v2 (Indonesia Stock Exchange):"))
	fmt.Println(wizardMutedStyle.Render("  (Get free key at https://sectors.app. Press ENTER for 100% Offline Mock Mode)."))
	fmt.Printf("%s Sectors API Key [optional]: ", wizardStepStyle.Render("►"))
	sectorsKey, _ := reader.ReadString('\n')
	sectorsKey = strings.TrimSpace(sectorsKey)

	if sectorsKey == "" {
		fmt.Printf("  %s %s\n", wizardSuccessBadgeStyle.Render("[✓]"), wizardMutedStyle.Render("Offline Mock Mode enabled (Law 5: Credit Conservation). All financial data fixtures active."))
	}

	// Step 3: Live Verification Ping
	fmt.Println()
	fmt.Println(wizardStepStyle.Render("Step 3: Live Connection Diagnostics:"))
	fmt.Print(wizardMutedStyle.Render("  Testing AI Provider endpoint... "))
	activeKey := openAIKey
	if aiProvider == "gemini" && geminiKey != "" {
		activeKey = geminiKey
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ok, msg, _ := TestLiveConnection(ctx, testTarget, openAIBaseURL, activeKey)
	if ok {
		fmt.Printf("%s %s\n", wizardSuccessBadgeStyle.Render("[✓ CONNECTED]"), wizardMutedStyle.Render(msg))
	} else {
		fmt.Printf("%s %s\n", wizardWarnBadgeStyle.Render("[! NOTICE]"), wizardMutedStyle.Render(msg))
		fmt.Println(wizardMutedStyle.Render("    (Configuration will still be saved. You can verify network or update keys anytime)."))
	}

	if sectorsKey != "" {
		fmt.Print(wizardMutedStyle.Render("  Testing Sectors Financial API... "))
		secOK, secMsg, _ := TestLiveConnection(ctx, "sectors", "", sectorsKey)
		if secOK {
			fmt.Printf("%s %s\n", wizardSuccessBadgeStyle.Render("[✓ CONNECTED]"), wizardMutedStyle.Render(secMsg))
		} else {
			fmt.Printf("%s %s\n", wizardWarnBadgeStyle.Render("[! NOTICE]"), wizardMutedStyle.Render(secMsg))
		}
	}

	// Step 4: Python Runtime Check
	fmt.Println()
	fmt.Println(wizardStepStyle.Render("Step 4: Python Engine Runtime Check:"))
	pyBin, pyDesc, pyReady := DetectPythonEnvironment()
	if pyReady {
		fmt.Printf("  %s %s (%s)\n", wizardSuccessBadgeStyle.Render("[✓ READY]"), pyBin, wizardMutedStyle.Render(pyDesc))
	} else {
		fmt.Printf("  %s %s (%s)\n", wizardWarnBadgeStyle.Render("[! ACTION REQUIRED]"), pyBin, wizardMutedStyle.Render(pyDesc))
		fmt.Printf("\n  %s Would you like Niskava to set up .venv and install requirements automatically? [Y/n, default: Y]: ", wizardStepStyle.Render("►"))
		autoChoice, _ := reader.ReadString('\n')
		autoChoice = strings.ToLower(strings.TrimSpace(autoChoice))
		if autoChoice == "" || autoChoice == "y" || autoChoice == "yes" {
			fmt.Printf("  %s Bootstrapping Python virtual environment (.venv) and installing requirements... ", wizardMutedStyle.Render("►"))
			wd, _ := os.Getwd()
			root := ipc.ResolveRepoRoot(wd)
			newPy, err := BootstrapPythonEnvironment(root, pyBin)
			if err != nil {
				fmt.Printf("%s\n    Error: %v\n", wizardErrBadgeStyle.Render("[FAILED]"), err)
			} else {
				fmt.Printf("%s\n", wizardSuccessBadgeStyle.Render("[SUCCESS]"))
				pyBin = newPy
				pyDesc = "Virtual environment (.venv) successfully created with dependencies"
				pyReady = true
				fmt.Printf("  %s %s (%s)\n", wizardSuccessBadgeStyle.Render("[✓ READY]"), pyBin, wizardMutedStyle.Render(pyDesc))
			}
		}
	}

	// Step 5: Inference Timeout Profile
	fmt.Println()
	fmt.Println(wizardStepStyle.Render("Step 5: AI Inference Timeout Profile:"))
	fmt.Println(wizardMutedStyle.Render("  Sets how long Niskava waits for the LLM to respond per reasoning step."))
	fmt.Printf("  %s %s %s\n",
		wizardItemBadgeStyle.Render("[1]"),
		"Fast / Cloud API          — 25 seconds",
		wizardMutedStyle.Render("Best for: OpenAI, Claude, Gemini API"))
	fmt.Printf("  %s %s %s\n",
		wizardItemBadgeStyle.Render("[2]"),
		"Balanced / Adaptive       — 60 seconds",
		wizardMutedStyle.Render("[Recommended] Handles multi-tool IDX analysis"))
	fmt.Printf("  %s %s %s\n",
		wizardItemBadgeStyle.Render("[3]"),
		"Deep Reasoning / Cloud    — 120 seconds",
		wizardMutedStyle.Render("Best for: o3, Gemini 2.5 Pro, DeepSeek R2"))
	fmt.Printf("  %s %s %s\n",
		wizardItemBadgeStyle.Render("[4]"),
		"Local LLM / Slow Hardware — 180 seconds",
		wizardMutedStyle.Render("Best for: Ollama on CPU, vLLM on low-VRAM GPU"))
	fmt.Printf("  %s Custom (enter seconds)\n",
		wizardItemBadgeStyle.Render("[5]"))
	fmt.Printf("\n%s Choice [1/2/3/4/5, default: 2]: ", wizardStepStyle.Render("►"))

	timeoutChoice, _ := reader.ReadString('\n')
	timeoutChoice = strings.TrimSpace(timeoutChoice)

	var chosenTimeoutSecs float64
	switch timeoutChoice {
	case "1":
		chosenTimeoutSecs = 25.0
	case "3":
		chosenTimeoutSecs = 120.0
	case "4":
		chosenTimeoutSecs = 180.0
	case "5":
		fmt.Printf("  %s Enter timeout in seconds (10–300): ", wizardStepStyle.Render("►"))
		rawSecs, _ := reader.ReadString('\n')
		rawSecs = strings.TrimSpace(rawSecs)
		var parsed float64
		if _, err := fmt.Sscanf(rawSecs, "%f", &parsed); err == nil && parsed >= 10 && parsed <= 300 {
			chosenTimeoutSecs = parsed
		} else {
			fmt.Printf("  %s Invalid value. Using Balanced (60s).\n", wizardWarnBadgeStyle.Render("[!]"))
			chosenTimeoutSecs = 60.0
		}
	default:
		chosenTimeoutSecs = 60.0
	}
	fmt.Printf("  %s Inference timeout set to %.0fs per LLM call.\n",
		wizardSuccessBadgeStyle.Render("[✓]"),
		chosenTimeoutSecs)

	// Step 6: Save Configuration
	params := SetupParams{
		AIProvider:     aiProvider,
		OpenAIBaseURL:  openAIBaseURL,
		OpenAIKey:      openAIKey,
		OpenAIModel:    openAIModel,
		GeminiKey:      geminiKey,
		GeminiModel:    geminiModel,
		SectorsKey:     sectorsKey,
		PythonBin:      pyBin,
		LLMTimeoutSecs: chosenTimeoutSecs,
	}

	if err := SaveSetupConfiguration(params); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	var completeBox strings.Builder
	completeBox.WriteString(wizardSuccessBadgeStyle.Render("SETUP COMPLETED! Configuration saved synchronously to:") + "\n")
	completeBox.WriteString(wizardMutedStyle.Render("  • Local Project : .env (0600 permissions)\n"))
	completeBox.WriteString(wizardMutedStyle.Render("  • Global Profile: ~/.niskava/config.yaml (Synchronized for CLI & Web Workspace)\n\n"))
	completeBox.WriteString(wizardMutedStyle.Render("Active Configuration Summary:\n"))
	completeBox.WriteString(fmt.Sprintf("  • Provider : %s (%s)\n", aiProvider, openAIModel))
	completeBox.WriteString(fmt.Sprintf("  • Endpoint : %s\n", openAIBaseURL))
	completeBox.WriteString(fmt.Sprintf("  • Timeout  : %.0fs per LLM step (Adaptive scaling)\n", chosenTimeoutSecs))
	if sectorsKey == "" {
		completeBox.WriteString("  • Sectors  : Offline Mock Mode (100% Free / Cached)\n")
	} else {
		completeBox.WriteString(fmt.Sprintf("  • Sectors  : Live Key Configured (%s)\n", config.MaskSecret(sectorsKey)))
	}
	completeBox.WriteString("\n" + wizardMutedStyle.Render("Launch Niskava Terminal & Web Workspace with:\n"))
	completeBox.WriteString("  " + wizardStepStyle.Render("./bin/niskava") + wizardMutedStyle.Render(" (or 'go run ./cmd/niskava')"))
	fmt.Println(setupCardStyle.Render(completeBox.String()))
	fmt.Println()

	return nil
}
