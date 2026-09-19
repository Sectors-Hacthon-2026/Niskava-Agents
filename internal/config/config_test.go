package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Server.Port)
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

// TestLoadDotEnvToConfig verifies a key written into a .env-style file is
// picked up by Load() via loadDotEnv (the real production path).
func TestLoadDotEnvToConfig(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	envContent := "SECTORS_API_KEY=sectors_key_from_file_123\nNISKAVA_PORT=9191\n"
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("failed to write test env: %v", err)
	}

	// loadDotEnv sets env vars only when absent; ensure clean slate.
	os.Unsetenv("SECTORS_API_KEY")
	defer os.Unsetenv("SECTORS_API_KEY")

	loadDotEnv(envPath)
	cfg, err := Load(filepath.Join(tempDir, "config.yaml"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Auth.SectorsAPIKey != "sectors_key_from_file_123" {
		t.Errorf("expected SectorsAPIKey from .env file, got %q", cfg.Auth.SectorsAPIKey)
	}
	if cfg.Server.Port != 9191 {
		t.Errorf("expected port 9191 from .env file, got %d", cfg.Server.Port)
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
