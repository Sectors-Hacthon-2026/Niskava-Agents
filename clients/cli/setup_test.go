package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildEnvContent_MockMode(t *testing.T) {
	// Case 1: Empty Sectors Key -> Must activate Mock Mode
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
	if !strings.Contains(envOffline, "NISKAVA_OFFLINE=1") {
		t.Errorf("expected NISKAVA_OFFLINE=1 when SectorsKey is empty, got:\n%s", envOffline)
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
