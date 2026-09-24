// Package config handles dynamic hierarchical configuration resolution for Niskava Agent.
// Resolution order: CLI Flags > Environment Variables > Config File (~/.niskava/config.yaml) > Defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	Memory      MemoryConfig      `yaml:"memory"`
	Telegram    TelegramConfig    `yaml:"telegram"`
}

// TelegramConfig stores parameters for the Telegram Bot integration.
type TelegramConfig struct {
	BotToken     string   `yaml:"bot_token"`
	Enabled      bool     `yaml:"enabled"`
	AllowedUsers []string `yaml:"allowed_users"`
}

// AuthConfig stores API keys and model parameters for external services.
type AuthConfig struct {
	AIProvider      string `yaml:"ai_provider"`
	SectorsAPIKey   string `yaml:"sectors_api_key"`
	SectorsBaseURL  string `yaml:"sectors_base_url"`
	GeminiAPIKey    string `yaml:"gemini_api_key"`
	GeminiModel     string `yaml:"gemini_model"`
	OpenAIAPIKey    string `yaml:"openai_api_key"`
	OpenAIBaseURL   string `yaml:"openai_base_url"`
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
	Language      string `yaml:"language"`
}

// MemoryConfig configures the local conversational graph memory engine.
type MemoryConfig struct {
	Enabled          bool    `yaml:"enabled"`
	DecayLambda      float64 `yaml:"decay_lambda"`
	EgoRadius        int     `yaml:"ego_radius"`
	MaxContextTokens int     `yaml:"max_context_tokens"`
}

// DefaultConfig returns safe baseline configuration values.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDBPath := filepath.Join(homeDir, ".niskava", "niskava.db")

	defaultPythonBin := "python3"
	if runtime.GOOS == "windows" {
		defaultPythonBin = "python"
	}

	return &Config{
		Auth: AuthConfig{
			AIProvider:      "gemini",
			SectorsAPIKey:   "",
			SectorsBaseURL:  "https://api.sectors.app/v2",
			GeminiAPIKey:    "",
			GeminiModel:     "gemini-2.0-flash",
			OpenAIAPIKey:    "",
			OpenAIBaseURL:   "http://localhost:20128/v1",
			OpenAIModel:     "hermes",
			AnthropicAPIKey: "",
			OllamaBaseURL:   "http://localhost:11434",
			OllamaModel:     "deepseek-r1:8b",
		},
		Storage: StorageConfig{
			DBPath: defaultDBPath,
		},
		Engine: EngineConfig{
			PythonBin:  defaultPythonBin,
			EnginePath: "./backend/engine",
		},
		Server: ServerConfig{
			Port: 20128,
		},
		Preferences: PreferencesConfig{
			DefaultMarket: "IDX",
			OfflineMode:   false,
			Language:      "en",
		},
		Memory: MemoryConfig{
			Enabled:          true,
			DecayLambda:      0.05,
			EgoRadius:        2,
			MaxContextTokens: 300,
		},
	}
}

// ExpandHome resolves a leading tilde (~) in file paths across POSIX and Windows.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			sub := strings.ReplaceAll(path[2:], `\`, string(filepath.Separator))
			sub = strings.ReplaceAll(sub, "/", string(filepath.Separator))
			return filepath.Join(home, sub)
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

// LoadFile parses a specific YAML configuration file directly without environment overrides.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(ExpandHome(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config %s: %w", path, err)
	}
	return cfg, nil
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
	if val := os.Getenv("AI_PROVIDER"); val != "" {
		cfg.Auth.AIProvider = val
	}
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
	if val := os.Getenv("OPENAI_BASE_URL"); val != "" {
		cfg.Auth.OpenAIBaseURL = val
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
	if val := os.Getenv("NISKAVA_LANG"); val != "" {
		cfg.Preferences.Language = strings.ToLower(val)
	}
	if val := os.Getenv("NISKAVA_TELEGRAM_TOKEN"); val != "" {
		cfg.Telegram.BotToken = val
	}
	if val := os.Getenv("TELEGRAM_BOT_TOKEN"); val != "" && cfg.Telegram.BotToken == "" {
		cfg.Telegram.BotToken = val
	}
	if val := os.Getenv("NISKAVA_TELEGRAM_ENABLED"); val == "1" || strings.ToLower(val) == "true" {
		cfg.Telegram.Enabled = true
	}
	if val := os.Getenv("NISKAVA_TELEGRAM_ALLOWED_USERS"); val != "" {
		parts := strings.Split(val, ",")
		var cleaned []string
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			cfg.Telegram.AllowedUsers = cleaned
		}
	}

	return cfg, nil
}

// MaskSecret masks sensitive credentials showing only prefix and suffix.
func MaskSecret(secret string) string {
	s := strings.TrimSpace(secret)
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// ConfigView represents the sanitized structure safe for serialization to web clients.
type ConfigView struct {
	Auth        AuthView          `json:"auth"`
	Storage     StorageConfig     `json:"storage"`
	Engine      EngineConfig      `json:"engine"`
	Server      ServerConfig      `json:"server"`
	Preferences PreferencesConfig `json:"preferences"`
	Memory      MemoryConfig      `json:"memory"`
	Telegram    TelegramView      `json:"telegram"`
}

type AuthView struct {
	AIProvider      string `json:"ai_provider"`
	SectorsAPIKey   string `json:"sectors_api_key"`
	HasSectorsKey   bool   `json:"has_sectors_key"`
	SectorsBaseURL  string `json:"sectors_base_url"`
	GeminiAPIKey    string `json:"gemini_api_key"`
	HasGeminiKey    bool   `json:"has_gemini_key"`
	GeminiModel     string `json:"gemini_model"`
	OpenAIAPIKey    string `json:"openai_api_key"`
	HasOpenAIKey    bool   `json:"has_openai_key"`
	OpenAIBaseURL   string `json:"openai_base_url"`
	OpenAIModel     string `json:"openai_model"`
	AnthropicAPIKey string `json:"anthropic_api_key"`
	HasAnthropicKey bool   `json:"has_anthropic_key"`
	OllamaBaseURL   string `json:"ollama_base_url"`
	OllamaModel     string `json:"ollama_model"`
}

type TelegramView struct {
	BotToken     string   `json:"bot_token"`
	HasToken     bool     `json:"has_token"`
	Enabled      bool     `json:"enabled"`
	AllowedUsers []string `json:"allowed_users"`
}

// MaskedView returns a sanitized view of the config without exposing raw secrets.
func (c *Config) MaskedView() ConfigView {
	return ConfigView{
		Auth: AuthView{
			AIProvider:      c.Auth.AIProvider,
			SectorsAPIKey:   MaskSecret(c.Auth.SectorsAPIKey),
			HasSectorsKey:   c.Auth.SectorsAPIKey != "",
			SectorsBaseURL:  c.Auth.SectorsBaseURL,
			GeminiAPIKey:    MaskSecret(c.Auth.GeminiAPIKey),
			HasGeminiKey:    c.Auth.GeminiAPIKey != "",
			GeminiModel:     c.Auth.GeminiModel,
			OpenAIAPIKey:    MaskSecret(c.Auth.OpenAIAPIKey),
			HasOpenAIKey:    c.Auth.OpenAIAPIKey != "",
			OpenAIBaseURL:   c.Auth.OpenAIBaseURL,
			OpenAIModel:     c.Auth.OpenAIModel,
			AnthropicAPIKey: MaskSecret(c.Auth.AnthropicAPIKey),
			HasAnthropicKey: c.Auth.AnthropicAPIKey != "",
			OllamaBaseURL:   c.Auth.OllamaBaseURL,
			OllamaModel:     c.Auth.OllamaModel,
		},
		Storage:     c.Storage,
		Engine:      c.Engine,
		Server:      c.Server,
		Preferences: c.Preferences,
		Memory:      c.Memory,
		Telegram: TelegramView{
			BotToken:     MaskSecret(c.Telegram.BotToken),
			HasToken:     c.Telegram.BotToken != "",
			Enabled:      c.Telegram.Enabled,
			AllowedUsers: c.Telegram.AllowedUsers,
		},
	}
}

// SaveConfig persists the active configuration to the specified YAML file or ~/.niskava/config.yaml with 0600 permissions.
func SaveConfig(cfg *Config, targetPath ...string) error {
	dest := ""
	if len(targetPath) > 0 && targetPath[0] != "" {
		dest = targetPath[0]
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home dir: %w", err)
		}
		dest = filepath.Join(homeDir, ".niskava", "config.yaml")
	}

	dest = ExpandHome(dest)
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	if err := os.WriteFile(dest, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}
