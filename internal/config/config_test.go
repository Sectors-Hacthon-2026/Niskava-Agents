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
