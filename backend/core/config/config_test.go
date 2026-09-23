package config

import (
	"os"
	"path/filepath"
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
	os.Setenv("MOCK_SECTORS", "1")
	defer func() {
		os.Unsetenv("SECTORS_API_KEY")
		os.Unsetenv("NISKAVA_PORT")
		os.Unsetenv("MOCK_SECTORS")
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
		t.Errorf("expected offline mode true from MOCK_SECTORS=1")
	}
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
