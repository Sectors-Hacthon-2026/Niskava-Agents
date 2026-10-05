package cli

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
)

func TestBuildEnvContentNeverWritesMockSectors1(t *testing.T) {
	params := SetupParams{
		SectorsKey:  "", // empty key
		AIProvider:  "gemini",
		GeminiKey:   "gk-test",
		GeminiModel: "gemini-2.0-flash",
	}
	env := BuildEnvContent(params)
	if strings.Contains(env, "MOCK_SECTORS=1") {
		t.Errorf("BuildEnvContent must never write MOCK_SECTORS=1, got:\n%s", env)
	}
	if strings.Contains(env, "NISKAVA_OFFLINE=1") {
		t.Errorf("BuildEnvContent must never write NISKAVA_OFFLINE=1, got:\n%s", env)
	}
	if !strings.Contains(env, "MOCK_SECTORS=0") {
		t.Errorf("BuildEnvContent must always write MOCK_SECTORS=0, got:\n%s", env)
	}
	if !strings.Contains(env, "NISKAVA_OFFLINE=0") {
		t.Errorf("BuildEnvContent must always write NISKAVA_OFFLINE=0, got:\n%s", env)
	}
}

func TestBuildEnvContentWithRealKey(t *testing.T) {
	params := SetupParams{
		SectorsKey:  "sectors-real-key-xyz",
		AIProvider:  "gemini",
		GeminiKey:   "gk-test",
		GeminiModel: "gemini-2.0-flash",
	}
	env := BuildEnvContent(params)
	if !strings.Contains(env, "SECTORS_API_KEY=sectors-real-key-xyz") {
		t.Errorf("BuildEnvContent must write the provided SECTORS_API_KEY, got:\n%s", env)
	}
	if !strings.Contains(env, "MOCK_SECTORS=0") {
		t.Errorf("BuildEnvContent must write MOCK_SECTORS=0 when key is provided, got:\n%s", env)
	}
	if !strings.Contains(env, "NISKAVA_OFFLINE=0") {
		t.Errorf("BuildEnvContent must write NISKAVA_OFFLINE=0 when key is provided, got:\n%s", env)
	}
}

func TestBuildEnvContent_MockMode(t *testing.T) {
	// Case 1: Empty Sectors Key with LLM configured -> MOCK_SECTORS=0 and NISKAVA_OFFLINE=0
	envOffline := BuildEnvContent(SetupParams{
		AIProvider:    "openai",
		OpenAIBaseURL: "https://openrouter.ai/api/v1",
		OpenAIKey:     "test-key",
		OpenAIModel:   "deepseek/deepseek-chat",
		SectorsKey:    "",
		PythonBin:     "python3",
	})
	if strings.Contains(envOffline, "MOCK_SECTORS=1") {
		t.Errorf("expected no MOCK_SECTORS=1 when SectorsKey is empty, got:\n%s", envOffline)
	}
	if !strings.Contains(envOffline, "MOCK_SECTORS=0") {
		t.Errorf("expected MOCK_SECTORS=0 when SectorsKey is empty, got:\n%s", envOffline)
	}
	if !strings.Contains(envOffline, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0 when SectorsKey is empty and AIProvider is openai, got:\n%s", envOffline)
	}

	// Case 2: Provided Sectors Key -> Must disable Mock Mode (MOCK_SECTORS=0)
	envLive := BuildEnvContent(SetupParams{
		AIProvider:  "gemini",
		GeminiKey:   "dummy-gemini-key",
		GeminiModel: "gemini-2.0-flash",
		SectorsKey:  "sec-live-token-123",
		PythonBin:   "python3",
	})
	if !strings.Contains(envLive, "MOCK_SECTORS=0") {
		t.Errorf("expected MOCK_SECTORS=0 when SectorsKey is provided, got:\n%s", envLive)
	}
	if !strings.Contains(envLive, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0 when SectorsKey is provided, got:\n%s", envLive)
	}

	// Case 3: Explicit offline provider -> Never write MOCK_SECTORS=1 or NISKAVA_OFFLINE=1
	envExplicitOffline := BuildEnvContent(SetupParams{
		AIProvider: "offline",
		SectorsKey: "",
		PythonBin:  "python3",
	})
	if strings.Contains(envExplicitOffline, "MOCK_SECTORS=1") {
		t.Errorf("expected no MOCK_SECTORS=1, got:\n%s", envExplicitOffline)
	}
	if strings.Contains(envExplicitOffline, "NISKAVA_OFFLINE=1") {
		t.Errorf("expected no NISKAVA_OFFLINE=1 when AIProvider is offline, got:\n%s", envExplicitOffline)
	}
	if !strings.Contains(envExplicitOffline, "MOCK_SECTORS=0") {
		t.Errorf("expected MOCK_SECTORS=0, got:\n%s", envExplicitOffline)
	}
	if !strings.Contains(envExplicitOffline, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0, got:\n%s", envExplicitOffline)
	}
}

func TestBuildEnvContent_DecouplesMockFromLLM(t *testing.T) {
	// Missing SectorsKey with valid Gemini key produces MOCK_SECTORS=0 and NISKAVA_OFFLINE=0
	envGeminiMock := BuildEnvContent(SetupParams{
		AIProvider:  "gemini",
		GeminiKey:   "AIzaSyValidGeminiKey",
		GeminiModel: "gemini-2.0-flash",
		SectorsKey:  "",
		PythonBin:   "python3",
	})
	if strings.Contains(envGeminiMock, "MOCK_SECTORS=1") {
		t.Errorf("expected no MOCK_SECTORS=1 when SectorsKey is missing, got:\n%s", envGeminiMock)
	}
	if !strings.Contains(envGeminiMock, "MOCK_SECTORS=0") {
		t.Errorf("expected MOCK_SECTORS=0 when SectorsKey is missing, got:\n%s", envGeminiMock)
	}
	if !strings.Contains(envGeminiMock, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0 when Gemini key is valid, got:\n%s", envGeminiMock)
	}

	// Providing SectorsKey produces MOCK_SECTORS=0 and NISKAVA_OFFLINE=0
	envLive := BuildEnvContent(SetupParams{
		AIProvider:  "gemini",
		GeminiKey:   "AIzaSyValidGeminiKey",
		GeminiModel: "gemini-2.0-flash",
		SectorsKey:  "sec-live-token-456",
		PythonBin:   "python3",
	})
	if !strings.Contains(envLive, "MOCK_SECTORS=0") {
		t.Errorf("expected MOCK_SECTORS=0 when SectorsKey is provided, got:\n%s", envLive)
	}
	if !strings.Contains(envLive, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0 when SectorsKey is provided, got:\n%s", envLive)
	}
}

func TestGetProviderPresets(t *testing.T) {
	presets := GetProviderPresets()
	if len(presets) == 0 {
		t.Fatal("expected non-empty provider presets")
	}

	// Verify OpenRouter preset
	openRouter, exists := presets["openrouter"]
	if !exists {
		t.Fatal("missing openrouter preset")
	}
	if openRouter.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("unexpected openrouter base url: %s", openRouter.BaseURL)
	}

	// Verify DeepSeek preset
	deepseek, exists := presets["deepseek"]
	if !exists {
		t.Fatal("missing deepseek preset")
	}
	if deepseek.BaseURL != "https://api.deepseek.com/v1" {
		t.Errorf("unexpected deepseek base url: %s", deepseek.BaseURL)
	}

	// Verify Ollama preset
	ollama, exists := presets["ollama"]
	if !exists {
		t.Fatal("missing ollama preset")
	}
	if ollama.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("unexpected ollama base url: %s", ollama.BaseURL)
	}
}

func TestDetectPythonEnvironment(t *testing.T) {
	bin, desc, ready := DetectPythonEnvironment()
	if bin == "" {
		t.Error("expected non-empty python binary")
	}
	if desc == "" {
		t.Error("expected non-empty python description")
	}
	t.Logf("Detected Python: bin=%s ready=%v desc=%s", bin, ready, desc)
}

func TestDetectPythonEnvironmentChecksRequiredPackages(t *testing.T) {
	bin, desc, ready := DetectPythonEnvironment()
	if bin == "" {
		t.Fatal("expected non-empty python binary")
	}
	// If ready is true, verify that numpy, pydantic, etc. actually import without error
	if ready {
		cmd := exec.Command(bin, "-c", "import requests, numpy, pydantic, networkx")
		if err := cmd.Run(); err != nil {
			t.Fatalf("DetectPythonEnvironment reported ready=true, but imports failed: %v (desc: %s)", err, desc)
		}
	}
}

func TestTestLiveConnection_MockServer(t *testing.T) {
	// Create mock OpenAI-compatible server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data": [{"id": "test-model"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok, msg, _ := TestLiveConnection(ctx, "custom", ts.URL, "test-key")
	if !ok {
		t.Errorf("expected TestLiveConnection to succeed, got msg: %s", msg)
	}
}

func TestBootstrapPythonEnvironment_Validation(t *testing.T) {
	_, err := BootstrapPythonEnvironment(t.TempDir(), "nonexistent-python-bin-xyz")
	if err == nil {
		t.Error("expected error for nonexistent python binary, got nil")
	}
}

func TestBuildEnvContentIncludesLLMTimeout(t *testing.T) {
	p := SetupParams{
		AIProvider:     "openai",
		OpenAIBaseURL:  "http://localhost:20128/v1",
		OpenAIModel:    "gpt-4o-mini",
		LLMTimeoutSecs: 90.0,
	}
	content := BuildEnvContent(p)
	if !strings.Contains(content, "NISKAVA_LLM_TIMEOUT=90.00") {
		t.Errorf("expected NISKAVA_LLM_TIMEOUT=90.00 in .env content, got:\n%s", content)
	}
}

func TestBuildEnvContentTimeoutDefaultsTo60(t *testing.T) {
	p := SetupParams{
		AIProvider:    "openai",
		OpenAIBaseURL: "https://api.openai.com/v1",
		OpenAIModel:   "gpt-4o-mini",
	}
	content := BuildEnvContent(p)
	if !strings.Contains(content, "NISKAVA_LLM_TIMEOUT=60.00") {
		t.Errorf("expected NISKAVA_LLM_TIMEOUT=60.00 in .env content, got:\n%s", content)
	}
}

func TestGetLaunchCommandHint(t *testing.T) {
	// If running outside repo root with go.mod, hint should be 'niskava'
	hintStandalone := getLaunchCommandHint("/tmp/random-folder-xyz")
	if hintStandalone != "niskava" {
		t.Errorf("expected 'niskava', got '%s'", hintStandalone)
	}

	wd, _ := os.Getwd()
	hintInRepo := getLaunchCommandHint(wd)
	if !strings.Contains(hintInRepo, "cmd/niskava") && !strings.Contains(hintInRepo, "bin/niskava") {
		t.Errorf("expected repo-specific command in repo, got '%s'", hintInRepo)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(not configured)"},
		{"short", "••••"},
		{"sk-1234567890abcdef", "sk-12••••cdef"},
		{"AIzaSyBx1234567890987654321", "AIzaSy••••4321"},
	}

	for _, tt := range tests {
		got := MaskAPIKey(tt.input)
		if got != tt.expected {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestPromptWithDefaultPreservesExistingOnEnter(t *testing.T) {
	input := "\n" // User simply pressed ENTER
	reader := bufio.NewReader(strings.NewReader(input))
	result := PromptWithDefault(reader, "Enter Key", "existing-secret-key", true)
	if result != "existing-secret-key" {
		t.Errorf("expected existing key preserved, got %q", result)
	}
}

func TestRenderConfigurationDashboard(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Auth.AIProvider = "gemini"
	cfg.Auth.GeminiModel = "gemini-2.5-flash"
	cfg.Auth.GeminiAPIKey = "AIzaSyTest1234567890"
	cfg.Auth.SectorsAPIKey = "sec_test_1234567890"

	dashboard := RenderConfigurationDashboard(cfg)
	if !strings.Contains(dashboard, "gemini") {
		t.Errorf("expected dashboard to contain provider 'gemini', got:\n%s", dashboard)
	}
	if !strings.Contains(dashboard, "gemini-2.5-flash") {
		t.Errorf("expected dashboard to contain model 'gemini-2.5-flash', got:\n%s", dashboard)
	}
	if !strings.Contains(dashboard, "AIzaSy••••7890") {
		t.Errorf("expected dashboard to contain masked Gemini key, got:\n%s", dashboard)
	}
}

func TestRenderConfigurationDashboard_Telegram(t *testing.T) {
	testCfg := config.DefaultConfig()
	testCfg.Telegram.BotToken = "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	testCfg.Telegram.Enabled = true
	testCfg.Telegram.AllowedUsers = []string{"11223344", "analyst_idx"}

	dashboard := RenderConfigurationDashboard(testCfg)
	if !strings.Contains(dashboard, "Telegram Bot") {
		t.Errorf("expected dashboard to mention Telegram Bot, got:\n%s", dashboard)
	}
	if !strings.Contains(dashboard, "2 users") {
		t.Errorf("expected dashboard to report 2 users whitelisted, got:\n%s", dashboard)
	}

	// Test unconfigured
	emptyCfg := config.DefaultConfig()
	emptyDashboard := RenderConfigurationDashboard(emptyCfg)
	if !strings.Contains(emptyDashboard, "Telegram Bot") {
		t.Errorf("expected dashboard to mention Telegram Bot for unconfigured, got:\n%s", emptyDashboard)
	}
	if !strings.Contains(emptyDashboard, "Not configured (Optional)") {
		t.Errorf("expected dashboard to show 'Not configured (Optional)', got:\n%s", emptyDashboard)
	}
}

func TestConfigureTelegramWizard_Disable(t *testing.T) {
	testCfg := config.DefaultConfig()
	testCfg.Telegram.Enabled = true
	testCfg.Telegram.BotToken = "old-token"

	reader := bufio.NewReader(strings.NewReader("n\n"))
	err := configureTelegramWizard(reader, testCfg)
	if err != nil {
		t.Fatalf("configureTelegramWizard returned error: %v", err)
	}
	if testCfg.Telegram.Enabled {
		t.Errorf("expected Telegram to be disabled, got enabled=true")
	}
}

func TestConfigureTelegramWizard_EnableAndConfigure(t *testing.T) {
	testCfg := config.DefaultConfig()
	testCfg.Telegram.Enabled = false

	input := "y\n123456789:TestTokenMock\nidx_trader, @market_watcher\n"
	reader := bufio.NewReader(strings.NewReader(input))
	err := configureTelegramWizard(reader, testCfg)
	if err != nil {
		t.Fatalf("configureTelegramWizard returned error: %v", err)
	}
	if !testCfg.Telegram.Enabled {
		t.Errorf("expected Telegram.Enabled = true")
	}
	if testCfg.Telegram.BotToken != "123456789:TestTokenMock" {
		t.Errorf("expected BotToken '123456789:TestTokenMock', got %q", testCfg.Telegram.BotToken)
	}
	if len(testCfg.Telegram.AllowedUsers) != 2 {
		t.Fatalf("expected 2 allowed users, got %d", len(testCfg.Telegram.AllowedUsers))
	}
	if testCfg.Telegram.AllowedUsers[0] != "idx_trader" || testCfg.Telegram.AllowedUsers[1] != "market_watcher" {
		t.Errorf("unexpected allowed users: %v", testCfg.Telegram.AllowedUsers)
	}
}

func TestIsCancelInput(t *testing.T) {
	cancels := []string{
		"q\n", "Q\n", "exit\n", "EXIT\n", "quit\n", "cancel\n", "esc\n", ":q\n",
		"\x1b", "\x1b\n", "\x03", "\x03\n",
	}
	for _, c := range cancels {
		if !IsCancelInput(c) {
			t.Errorf("expected IsCancelInput(%q) to be true", c)
		}
	}

	nonCancels := []string{
		"1\n", "y\n", "n\n", "sk-ant-api-key\n", "gemini-2.5-flash\n", "\n", "   \n",
	}
	for _, nc := range nonCancels {
		if IsCancelInput(nc) {
			t.Errorf("expected IsCancelInput(%q) to be false", nc)
		}
	}
}

func TestIsBackInput(t *testing.T) {
	backs := []string{"0", "0\n", "b", "B\n", "back", "BACK\n"}
	for _, b := range backs {
		if !IsBackInput(b) {
			t.Errorf("expected IsBackInput(%q) to be true", b)
		}
	}

	nonBacks := []string{"1", "2", "3", "4", "5", "gemini"}
	for _, nb := range nonBacks {
		if IsBackInput(nb) {
			t.Errorf("expected IsBackInput(%q) to be false", nb)
		}
	}
}

func TestPromptWithDefault_Cancel(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("q\n"))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected PromptWithDefault to panic with setupCancelSignal on 'q'")
		}
		if _, ok := r.(setupCancelSignal); !ok {
			t.Fatalf("expected setupCancelSignal, got %T: %v", r, r)
		}
	}()
	_ = PromptWithDefault(reader, "API Key", "current-key", true)
}
