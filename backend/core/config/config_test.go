package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Port != 20128 {
		t.Fatalf("expected default port 20128, got %d", cfg.Server.Port)
	}
	if cfg.Preferences.DefaultMarket != "IDX" {
		t.Fatalf("expected default market IDX, got %s", cfg.Preferences.DefaultMarket)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	os.Setenv("SECTORS_API_KEY", "test_sectors_key_123")
	os.Setenv("NISKAVA_PORT", "9090")
	os.Setenv("NISKAVA_OFFLINE", "1")
	defer func() {
		os.Unsetenv("SECTORS_API_KEY")
		os.Unsetenv("NISKAVA_PORT")
		os.Unsetenv("NISKAVA_OFFLINE")
	}()

	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "config.yaml")

	cfg, err := Load(nonExistentPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Auth.SectorsAPIKey != "test_sectors_key_123" {
		t.Errorf("expected SectorsAPIKey from env, got %s", cfg.Auth.SectorsAPIKey)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if !cfg.Preferences.OfflineMode {
		t.Errorf("expected offline mode true from NISKAVA_OFFLINE=1")
	}
}

func TestPreferencesOfflineMode_DynamicLiveToggle(t *testing.T) {
	// Subtest 1: SECTORS_API_KEY present + stale MOCK_SECTORS=1 -> OfflineMode should auto-toggle to false
	t.Run("auto-toggle live when sectors key present", func(t *testing.T) {
		os.Setenv("SECTORS_API_KEY", "sec_live_key_999")
		os.Setenv("MOCK_SECTORS", "1")
		os.Unsetenv("NISKAVA_OFFLINE")
		defer func() {
			os.Unsetenv("SECTORS_API_KEY")
			os.Unsetenv("MOCK_SECTORS")
		}()

		tempDir := t.TempDir()
		cfg, err := Load(filepath.Join(tempDir, "config.yaml"))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Preferences.OfflineMode {
			t.Errorf("expected OfflineMode=false when valid SECTORS_API_KEY is present even with MOCK_SECTORS=1")
		}
	})

	// Subtest 2: SECTORS_API_KEY present + explicit NISKAVA_OFFLINE=1 -> OfflineMode should be true
	t.Run("explicit offline takes precedence", func(t *testing.T) {
		os.Setenv("SECTORS_API_KEY", "sec_live_key_999")
		os.Setenv("NISKAVA_OFFLINE", "1")
		defer func() {
			os.Unsetenv("SECTORS_API_KEY")
			os.Unsetenv("NISKAVA_OFFLINE")
		}()

		tempDir := t.TempDir()
		cfg, err := Load(filepath.Join(tempDir, "config.yaml"))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if !cfg.Preferences.OfflineMode {
			t.Errorf("expected OfflineMode=true when NISKAVA_OFFLINE=1 is explicitly set")
		}
	})

	// Subtest 3: SECTORS_API_KEY from YAML + stale MOCK_SECTORS=1 -> OfflineMode should auto-toggle to false
	t.Run("yaml sectors key auto-toggles live", func(t *testing.T) {
		os.Unsetenv("SECTORS_API_KEY")
		os.Setenv("MOCK_SECTORS", "1")
		os.Unsetenv("NISKAVA_OFFLINE")
		defer func() {
			os.Unsetenv("MOCK_SECTORS")
		}()

		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, "config.yaml")
		yamlContent := `auth:
  sectors_api_key: "sec_yaml_live_key"
preferences:
  offline_mode: true
`
		if err := os.WriteFile(cfgFile, []byte(yamlContent), 0600); err != nil {
			t.Fatalf("failed to write config.yaml: %v", err)
		}

		cfg, err := Load(cfgFile)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Preferences.OfflineMode {
			t.Errorf("expected OfflineMode=false when Sectors key is in config.yaml")
		}
	})

	// Subtest 4: Empty SECTORS_API_KEY + MOCK_SECTORS=1 -> OfflineMode should be true
	t.Run("empty sectors key with mock sectors", func(t *testing.T) {
		os.Unsetenv("SECTORS_API_KEY")
		os.Setenv("MOCK_SECTORS", "1")
		os.Unsetenv("NISKAVA_OFFLINE")
		defer func() {
			os.Unsetenv("MOCK_SECTORS")
		}()

		tempDir := t.TempDir()
		cfg, err := Load(filepath.Join(tempDir, "config.yaml"))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if !cfg.Preferences.OfflineMode {
			t.Errorf("expected OfflineMode=true when SECTORS_API_KEY is empty and MOCK_SECTORS=1")
		}
	})
}

func TestLoadDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	envContent := `
# Test comment
TEST_DOTENV_KEY=sectors_secret_val
TEST_DOTENV_QUOTED="gemini_secret_val"
`
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("failed to write test env: %v", err)
	}

	loadDotEnv(envPath)
	defer func() {
		os.Unsetenv("TEST_DOTENV_KEY")
		os.Unsetenv("TEST_DOTENV_QUOTED")
	}()

	if os.Getenv("TEST_DOTENV_KEY") != "sectors_secret_val" {
		t.Errorf("expected sectors_secret_val, got %s", os.Getenv("TEST_DOTENV_KEY"))
	}
	if os.Getenv("TEST_DOTENV_QUOTED") != "gemini_secret_val" {
		t.Errorf("expected gemini_secret_val, got %s", os.Getenv("TEST_DOTENV_QUOTED"))
	}
}

func TestTelegramConfigParsing(t *testing.T) {
	os.Setenv("NISKAVA_TELEGRAM_TOKEN", "test-bot-token-xyz")
	os.Setenv("NISKAVA_TELEGRAM_ENABLED", "true")
	os.Setenv("NISKAVA_TELEGRAM_ALLOWED_USERS", "user1,123456")
	defer func() {
		os.Unsetenv("NISKAVA_TELEGRAM_TOKEN")
		os.Unsetenv("NISKAVA_TELEGRAM_ENABLED")
		os.Unsetenv("NISKAVA_TELEGRAM_ALLOWED_USERS")
	}()

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Telegram.BotToken != "test-bot-token-xyz" {
		t.Errorf("expected bot token 'test-bot-token-xyz', got '%s'", cfg.Telegram.BotToken)
	}
	if !cfg.Telegram.Enabled {
		t.Errorf("expected telegram enabled to be true")
	}
	if len(cfg.Telegram.AllowedUsers) != 2 || cfg.Telegram.AllowedUsers[0] != "user1" || cfg.Telegram.AllowedUsers[1] != "123456" {
		t.Errorf("unexpected allowed users: %v", cfg.Telegram.AllowedUsers)
	}
}

func TestTelegramAllowedUsersEmptyReset(t *testing.T) {
	os.Setenv("NISKAVA_TELEGRAM_ALLOWED_USERS", "")
	defer os.Unsetenv("NISKAVA_TELEGRAM_ALLOWED_USERS")
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(cfg.Telegram.AllowedUsers) != 0 {
		t.Errorf("expected empty allowed users, got %v", cfg.Telegram.AllowedUsers)
	}
}

func TestExpandHome_CrossPlatform(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping test: UserHomeDir not available")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~", home},
		{"~/", home},
		{"~/.niskava", filepath.Join(home, ".niskava")},
		{"~/.niskava/config.yaml", filepath.Join(home, ".niskava", "config.yaml")},
		{`~\.niskava\config.yaml`, filepath.Join(home, ".niskava", "config.yaml")},
		{"/var/log/niskava.log", "/var/log/niskava.log"},
		{"relative/path.db", "relative/path.db"},
	}

	for _, tt := range tests {
		got := ExpandHome(tt.input)
		if got != tt.expected {
			t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSaveConfigAndMaskedView(t *testing.T) {
	origKey := os.Getenv("SECTORS_API_KEY")
	_ = os.Unsetenv("SECTORS_API_KEY")
	defer func() {
		if origKey != "" {
			_ = os.Setenv("SECTORS_API_KEY", origKey)
		}
	}()

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Auth.SectorsAPIKey = "sec_test_key_12345"
	cfg.Auth.GeminiAPIKey = "AIzaSyTestGeminiSecret"
	cfg.Auth.AIProvider = "gemini"

	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.Auth.SectorsAPIKey != "sec_test_key_12345" {
		t.Fatalf("expected saved key to match, got %s", loaded.Auth.SectorsAPIKey)
	}

	view := cfg.MaskedView()
	if view.Auth.SectorsAPIKey == "sec_test_key_12345" {
		t.Fatalf("SectorsAPIKey should be masked in view, got: %s", view.Auth.SectorsAPIKey)
	}
	if !strings.Contains(view.Auth.SectorsAPIKey, "****") {
		t.Fatalf("expected mask pattern with ****, got: %s", view.Auth.SectorsAPIKey)
	}
}

func TestSaveDotEnv_SSoT(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	cfg := DefaultConfig()
	cfg.Auth.AIProvider = "openai"
	cfg.Auth.OpenAIBaseURL = "http://localhost:20128/v1"
	cfg.Auth.OpenAIModel = "gpt-4o-mini"
	cfg.Auth.OpenAIAPIKey = "sk-custom-test-123"

	if err := SaveDotEnv(cfg, envPath); err != nil {
		t.Fatalf("SaveDotEnv failed: %v", err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read written .env: %v", err)
	}
	strContent := string(content)

	if !strings.Contains(strContent, "OPENAI_BASE_URL=http://localhost:20128/v1") {
		t.Errorf("expected OPENAI_BASE_URL in .env, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "OPENAI_MODEL=gpt-4o-mini") {
		t.Errorf("expected OPENAI_MODEL in .env, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "OPENAI_API_KEY=sk-custom-test-123") {
		t.Errorf("expected OPENAI_API_KEY in .env, got:\n%s", strContent)
	}
}

func TestLLMTimeoutSecsDefaultAndForwarding(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Preferences.LLMTimeoutSecs != 60.0 {
		t.Errorf("expected default LLMTimeoutSecs 60.0, got %v", cfg.Preferences.LLMTimeoutSecs)
	}
	env := cfg.BuildSubprocessEnv()
	if env["NISKAVA_LLM_TIMEOUT"] != "60.00" {
		t.Errorf("expected NISKAVA_LLM_TIMEOUT=60.00, got %q", env["NISKAVA_LLM_TIMEOUT"])
	}
	cfg.Preferences.LLMTimeoutSecs = 120.0
	env2 := cfg.BuildSubprocessEnv()
	if env2["NISKAVA_LLM_TIMEOUT"] != "120.00" {
		t.Errorf("expected NISKAVA_LLM_TIMEOUT=120.00, got %q", env2["NISKAVA_LLM_TIMEOUT"])
	}
}

func TestLLMTimeoutSecsClamp(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Preferences.LLMTimeoutSecs = 5.0
	env := cfg.BuildSubprocessEnv()
	if env["NISKAVA_LLM_TIMEOUT"] != "10.00" {
		t.Errorf("expected clamped to 10.00, got %q", env["NISKAVA_LLM_TIMEOUT"])
	}
	cfg.Preferences.LLMTimeoutSecs = 999.0
	env2 := cfg.BuildSubprocessEnv()
	if env2["NISKAVA_LLM_TIMEOUT"] != "300.00" {
		t.Errorf("expected clamped to 300.00, got %q", env2["NISKAVA_LLM_TIMEOUT"])
	}
}

func TestSaveDotEnv_DualSync(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(origWd)
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to tempDir: %v", err)
	}

	// Create a local .env in the working directory
	localEnv := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(localEnv, []byte("AI_PROVIDER=openai\nSECTORS_API_KEY=old_sectors_key\n"), 0600); err != nil {
		t.Fatalf("failed to create local .env: %v", err)
	}

	customHomeEnv := filepath.Join(tempDir, "user_home", ".env")
	cfg := DefaultConfig()
	cfg.Auth.AIProvider = "gemini"
	cfg.Auth.SectorsAPIKey = "new_sectors_key_999"
	cfg.Telegram.Enabled = true

	// Call SaveDotEnv targeting customHomeEnv
	err := SaveDotEnv(cfg, customHomeEnv)
	if err != nil {
		t.Fatalf("SaveDotEnv failed: %v", err)
	}

	// Verify customHomeEnv has AI_PROVIDER=gemini and SECTORS_API_KEY=new_sectors_key_999
	homeData, err := os.ReadFile(customHomeEnv)
	if err != nil {
		t.Fatalf("failed to read home .env: %v", err)
	}
	if !strings.Contains(string(homeData), "AI_PROVIDER=gemini") {
		t.Errorf("expected home .env to contain AI_PROVIDER=gemini, got: %s", string(homeData))
	}
	if !strings.Contains(string(homeData), "SECTORS_API_KEY=new_sectors_key_999") {
		t.Errorf("expected home .env to contain new_sectors_key_999, got: %s", string(homeData))
	}

	// Verify local .env in cwd was ALSO updated to AI_PROVIDER=gemini and new_sectors_key_999
	localData, err := os.ReadFile(localEnv)
	if err != nil {
		t.Fatalf("failed to read local .env: %v", err)
	}
	if !strings.Contains(string(localData), "AI_PROVIDER=gemini") {
		t.Errorf("expected local .env to be dual-synced to AI_PROVIDER=gemini, got: %s", string(localData))
	}
	if !strings.Contains(string(localData), "SECTORS_API_KEY=new_sectors_key_999") {
		t.Errorf("expected local .env to be dual-synced to new_sectors_key_999, got: %s", string(localData))
	}
}

func TestBuildSubprocessEnv_OfflineAndTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Preferences.OfflineMode = false
	cfg.Preferences.LLMTimeoutSecs = 120.0

	env := cfg.BuildSubprocessEnv()
	if env["NISKAVA_OFFLINE"] != "0" {
		t.Errorf("expected NISKAVA_OFFLINE=0 when OfflineMode=false, got %q", env["NISKAVA_OFFLINE"])
	}
	if env["MOCK_SECTORS"] != "0" {
		t.Errorf("expected MOCK_SECTORS=0 when OfflineMode=false, got %q", env["MOCK_SECTORS"])
	}
	if env["NISKAVA_LLM_TIMEOUT"] != "120.00" {
		t.Errorf("expected NISKAVA_LLM_TIMEOUT=120.00, got %q", env["NISKAVA_LLM_TIMEOUT"])
	}

	cfg.Preferences.OfflineMode = true
	envOffline := cfg.BuildSubprocessEnv()
	if envOffline["NISKAVA_OFFLINE"] != "1" {
		t.Errorf("expected NISKAVA_OFFLINE=1 when OfflineMode=true, got %q", envOffline["NISKAVA_OFFLINE"])
	}
	if envOffline["MOCK_SECTORS"] != "1" {
		t.Errorf("expected MOCK_SECTORS=1 when OfflineMode=true, got %q", envOffline["MOCK_SECTORS"])
	}
}

func TestDynamicProviderAndModelResolution(t *testing.T) {
	// Case 1: Gemini Provider with custom model
	cfg := DefaultConfig()
	cfg.Auth.AIProvider = "gemini"
	cfg.Auth.GeminiAPIKey = "AIzaSyTestKey123"
	cfg.Auth.GeminiModel = "gemini-2.5-flash"
	cfg.Auth.OpenAIModel = "gpt-4o-mini"

	if cfg.GetActiveProvider() != "gemini" {
		t.Errorf("expected provider gemini, got %s", cfg.GetActiveProvider())
	}
	if cfg.GetActiveModel() != "gemini-2.5-flash" {
		t.Errorf("expected model gemini-2.5-flash, got %s", cfg.GetActiveModel())
	}

	env := cfg.BuildSubprocessEnv()
	if env["AI_PROVIDER"] != "gemini" {
		t.Errorf("expected env AI_PROVIDER=gemini, got %s", env["AI_PROVIDER"])
	}
	if env["NISKAVA_MODEL"] != "gemini-2.5-flash" {
		t.Errorf("expected env NISKAVA_MODEL=gemini-2.5-flash, got %s", env["NISKAVA_MODEL"])
	}
	if env["OPENAI_MODEL"] != "gemini-2.5-flash" {
		t.Errorf("expected normalized OPENAI_MODEL=gemini-2.5-flash for Gemini adapter, got %s", env["OPENAI_MODEL"])
	}
	if env["OPENAI_BASE_URL"] != "https://generativelanguage.googleapis.com/v1beta/openai" {
		t.Errorf("expected Google AI Studio base url, got %s", env["OPENAI_BASE_URL"])
	}

	// Case 2: OpenAI Provider with custom model
	cfg2 := DefaultConfig()
	cfg2.Auth.AIProvider = "openai"
	cfg2.Auth.OpenAIModel = "deepseek-chat"
	cfg2.Auth.OpenAIBaseURL = "https://api.deepseek.com/v1"

	if cfg2.GetActiveProvider() != "openai" {
		t.Errorf("expected provider openai, got %s", cfg2.GetActiveProvider())
	}
	if cfg2.GetActiveModel() != "deepseek-chat" {
		t.Errorf("expected model deepseek-chat, got %s", cfg2.GetActiveModel())
	}
	env2 := cfg2.BuildSubprocessEnv()
	if env2["NISKAVA_MODEL"] != "deepseek-chat" {
		t.Errorf("expected NISKAVA_MODEL=deepseek-chat, got %s", env2["NISKAVA_MODEL"])
	}

	// Case 3: Ollama Provider
	cfg3 := DefaultConfig()
	cfg3.Auth.AIProvider = "ollama"
	cfg3.Auth.OllamaModel = "qwen2.5:7b"
	cfg3.Auth.OllamaBaseURL = "http://localhost:11434"

	if cfg3.GetActiveProvider() != "ollama" {
		t.Errorf("expected provider ollama, got %s", cfg3.GetActiveProvider())
	}
	if cfg3.GetActiveModel() != "qwen2.5:7b" {
		t.Errorf("expected model qwen2.5:7b, got %s", cfg3.GetActiveModel())
	}
	env3 := cfg3.BuildSubprocessEnv()
	if env3["NISKAVA_MODEL"] != "qwen2.5:7b" {
		t.Errorf("expected NISKAVA_MODEL=qwen2.5:7b, got %s", env3["NISKAVA_MODEL"])
	}
	if env3["OPENAI_BASE_URL"] != "http://localhost:11434/v1" {
		t.Errorf("expected http://localhost:11434/v1, got %s", env3["OPENAI_BASE_URL"])
	}
}
