// Package config handles dynamic hierarchical configuration resolution for Niskava Agent.
// Resolution order: CLI Flags > Environment Variables > Config File (~/.niskava/config.yaml) > Defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the complete runtime configuration.
type Config struct {
	Auth        AuthConfig        `yaml:"auth"`
	Storage     StorageConfig     `yaml:"storage"`
	Engine      EngineConfig      `yaml:"engine"`
	Server      ServerConfig      `yaml:"server"`
	Preferences PreferencesConfig `yaml:"preferences"`
}

// AuthConfig stores API keys and model parameters for external services.
type AuthConfig struct {
	SectorsAPIKey   string `yaml:"sectors_api_key"`
	SectorsBaseURL  string `yaml:"sectors_base_url"`
	GeminiAPIKey    string `yaml:"gemini_api_key"`
	GeminiModel     string `yaml:"gemini_model"`
	OpenAIAPIKey    string `yaml:"openai_api_key"`
	OpenAIModel     string `yaml:"openai_model"`
	AnthropicAPIKey string `yaml:"anthropic_api_key"`
	OllamaBaseURL   string `yaml:"ollama_base_url"`
	OllamaModel     string `yaml:"ollama_model"`
}

// StorageConfig stores persistence parameters.
type StorageConfig struct {
	DBPath string `yaml:"db_path"`
}

// EngineConfig holds parameters for invoking the Python Agent Engine.
type EngineConfig struct {
	PythonBin  string `yaml:"python_bin"`
	EnginePath string `yaml:"entrypoint"`
}

// ServerConfig configures the local REST/SSE server.
type ServerConfig struct {
	Port int `yaml:"port"`
}

// PreferencesConfig configures agent behavior preferences.
type PreferencesConfig struct {
	DefaultMarket string `yaml:"default_market"`
	OfflineMode   bool   `yaml:"offline_mode"`
}

// DefaultConfig returns safe baseline configuration values.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(homeDir, ".niskava", "niskava.db")

	return &Config{
		Auth: AuthConfig{
			SectorsAPIKey:   "",
			SectorsBaseURL:  "https://api.sectors.app/v2",
			GeminiAPIKey:    "",
			GeminiModel:     "gemini-2.0-flash",
			OpenAIAPIKey:    "",
			OpenAIModel:     "gpt-4o-mini",
			AnthropicAPIKey: "",
			OllamaBaseURL:   "http://localhost:11434",
			OllamaModel:     "deepseek-r1:8b",
		},
		Storage: StorageConfig{
			DBPath: defaultDBPath,
		},
		Engine: EngineConfig{
			PythonBin:  "python3",
			EnginePath: "./engine",
		},
		Server: ServerConfig{
			Port: 8080,
		},
		Preferences: PreferencesConfig{
			DefaultMarket: "IDX",
			OfflineMode:   false,
		},
	}
}

// ExpandHome resolves a leading tilde (~) in file paths.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// loadDotEnv reads key-value pairs from .env files and sets them if not already set.
func loadDotEnv(paths ...string) {
	for _, path := range paths {
		expanded := ExpandHome(path)
		data, err := os.ReadFile(expanded)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"'`)
				if os.Getenv(key) == "" {
					_ = os.Setenv(key, val)
				}
			}
		}
	}
}

// Load reads and merges configuration from defaults, ~/.niskava/config.yaml, and environment variables.
func Load(customConfigPath string) (*Config, error) {
	// 0. Auto-load .env files (project root and ~/.niskava/.env)
	loadDotEnv(".env", "~/.niskava/.env")

	cfg := DefaultConfig()

	// 1. Resolve configuration file path
	configPath := customConfigPath
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(homeDir, ".niskava", "config.yaml")
		}
	} else {
		configPath = ExpandHome(configPath)
	}

	// 2. Read config file if exists
	if _, err := os.Stat(configPath); err == nil {
		data, readErr := os.ReadFile(configPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, readErr)
		}
		if unmarshalErr := yaml.Unmarshal(data, cfg); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse yaml config %s: %w", configPath, unmarshalErr)
		}
	}

	// 3. Override from Environment Variables (higher priority than file)
	if val := os.Getenv("SECTORS_API_KEY"); val != "" {
		cfg.Auth.SectorsAPIKey = val
	}
	if val := os.Getenv("SECTORS_BASE_URL"); val != "" {
		cfg.Auth.SectorsBaseURL = val
	}
	if val := os.Getenv("GEMINI_API_KEY"); val != "" {
		cfg.Auth.GeminiAPIKey = val
	}
	if val := os.Getenv("GEMINI_MODEL"); val != "" {
		cfg.Auth.GeminiModel = val
	}
	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		cfg.Auth.OpenAIAPIKey = val
	}
	if val := os.Getenv("OPENAI_MODEL"); val != "" {
		cfg.Auth.OpenAIModel = val
	}
	if val := os.Getenv("ANTHROPIC_API_KEY"); val != "" {
		cfg.Auth.AnthropicAPIKey = val
	}
	if val := os.Getenv("OLLAMA_BASE_URL"); val != "" {
		cfg.Auth.OllamaBaseURL = val
	}
	if val := os.Getenv("OLLAMA_MODEL"); val != "" {
		cfg.Auth.OllamaModel = val
	}
	if val := os.Getenv("NISKAVA_DB_PATH"); val != "" {
		cfg.Storage.DBPath = ExpandHome(val)
	} else {
		cfg.Storage.DBPath = ExpandHome(cfg.Storage.DBPath)
	}
	if val := os.Getenv("NISKAVA_PYTHON_BIN"); val != "" {
		cfg.Engine.PythonBin = val
	}
	if val := os.Getenv("NISKAVA_ENGINE_PATH"); val != "" {
		cfg.Engine.EnginePath = val
	}
	if val := os.Getenv("NISKAVA_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = p
		}
	}
	if val := os.Getenv("NISKAVA_DEFAULT_MARKET"); val != "" {
		cfg.Preferences.DefaultMarket = strings.ToUpper(val)
	}
	if val := os.Getenv("MOCK_SECTORS"); val == "1" || strings.ToLower(val) == "true" {
		cfg.Preferences.OfflineMode = true
	}
	if val := os.Getenv("NISKAVA_OFFLINE"); val == "1" || strings.ToLower(val) == "true" {
		cfg.Preferences.OfflineMode = true
	}

	return cfg, nil
}
