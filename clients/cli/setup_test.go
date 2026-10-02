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

func TestBuildEnvContent_MockMode(t *testing.T) {
	// Case 1: Empty Sectors Key with LLM configured -> MOCK_SECTORS=1 but NISKAVA_OFFLINE=0 so LLM can converse
	envOffline := BuildEnvContent(SetupParams{
		AIProvider:    "openai",
		OpenAIBaseURL: "https://openrouter.ai/api/v1",
		OpenAIKey:     "test-key",
		OpenAIModel:   "deepseek/deepseek-chat",
		SectorsKey:    "",
		PythonBin:     "python3",
	})
	if !strings.Contains(envOffline, "MOCK_SECTORS=1") {
		t.Errorf("expected MOCK_SECTORS=1 when SectorsKey is empty, got:\n%s", envOffline)
	}
	if !strings.Contains(envOffline, "NISKAVA_OFFLINE=0") {
		t.Errorf("expected NISKAVA_OFFLINE=0 when SectorsKey is empty and AIProvider is openai, got:\n%s", envOffline)
	}

	// Case 2: Provided Sectors Key -> Must disable Mock Mode
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

	// Case 3: Explicit offline provider -> NISKAVA_OFFLINE=1
	envExplicitOffline := BuildEnvContent(SetupParams{
		AIProvider: "offline",
		SectorsKey: "",
		PythonBin:  "python3",
	})
	if !strings.Contains(envExplicitOffline, "MOCK_SECTORS=1") {
		t.Errorf("expected MOCK_SECTORS=1, got:\n%s", envExplicitOffline)
	}
	if !strings.Contains(envExplicitOffline, "NISKAVA_OFFLINE=1") {
		t.Errorf("expected NISKAVA_OFFLINE=1 when AIProvider is offline, got:\n%s", envExplicitOffline)
	}
}

func TestBuildEnvContent_DecouplesMockFromLLM(t *testing.T) {
	// Missing SectorsKey with valid Gemini key produces MOCK_SECTORS=1 but NISKAVA_OFFLINE=0
	envGeminiMock := BuildEnvContent(SetupParams{
		AIProvider:  "gemini",
		GeminiKey:   "AIzaSyValidGeminiKey",
		GeminiModel: "gemini-2.0-flash",
		SectorsKey:  "",
		PythonBin:   "python3",
	})
	if !strings.Contains(envGeminiMock, "MOCK_SECTORS=1") {
		t.Errorf("expected MOCK_SECTORS=1 when SectorsKey is missing, got:\n%s", envGeminiMock)
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
		OpenAIModel:    "hermes",
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
