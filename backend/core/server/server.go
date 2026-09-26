// Package server provides the background REST/SSE daemon server for Niskava Agent.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/telegram"
)

// SessionManager manages active running session streams and allows cancellation (OpenCode pattern).
type SessionManager struct {
	mu     sync.Mutex
	active map[string]context.CancelFunc
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		active: make(map[string]context.CancelFunc),
	}
}

func (sm *SessionManager) Register(sessionID string, cancel context.CancelFunc) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.active[sessionID]; exists {
		return false
	}
	sm.active[sessionID] = cancel
	return true
}

func (sm *SessionManager) Unregister(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.active, sessionID)
}

func (sm *SessionManager) Abort(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if cancel, exists := sm.active[sessionID]; exists {
		cancel()
		delete(sm.active, sessionID)
		return true
	}
	return false
}

func (sm *SessionManager) IsBusy(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, exists := sm.active[sessionID]
	return exists
}

func (sm *SessionManager) AbortAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id, cancel := range sm.active {
		cancel()
		delete(sm.active, id)
	}
}

// Server encapsulates the background HTTP server instance.
type Server struct {
	httpServer     *http.Server
	Port           int
	DB             *db.DB
	URL            string
	SessionManager *SessionManager
	Config         *config.Config
	ConfigPath     string
	cfgMu          sync.RWMutex
	BotService     *telegram.BotService
	botMu          sync.Mutex
}

// buildSubprocessEnv extracts active authentication and preferences from s.Config into dynamic environment variables.
func (s *Server) buildSubprocessEnv() map[string]string {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.Config.BuildSubprocessEnv()
}

// syncTelegramBotState synchronizes running Telegram bot service with the latest s.Config settings.
func (s *Server) syncTelegramBotState() {
	s.cfgMu.RLock()
	cfg := s.Config
	s.cfgMu.RUnlock()

	s.botMu.Lock()
	defer s.botMu.Unlock()

	if cfg == nil {
		return
	}

	trimmedToken := strings.TrimSpace(cfg.Telegram.BotToken)
	botRunning := s.BotService != nil && s.BotService.IsStarted()

	// If disabled or empty token, stop bot if running
	if !cfg.Telegram.Enabled || trimmedToken == "" {
		if botRunning {
			s.BotService.Stop()
			s.BotService = nil
		}
		return
	}

	// If already running with the exact same token, no need to recreate
	if botRunning && s.BotService.Token() == trimmedToken {
		return
	}

	// Token changed or bot not running: stop existing instance if any
	if botRunning {
		s.BotService.Stop()
		s.BotService = nil
	}

	if botSvc, err := telegram.NewBotService(cfg, s.DB, s.SessionManager); err == nil {
		s.BotService = botSvc
		_ = s.BotService.Start()
	}
}

// ChatRequest represents the JSON payload for /api/chat.
type ChatRequest struct {
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id,omitempty"`
}

// CreateSessionRequest represents the JSON payload to create a new session.
type CreateSessionRequest struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Model string `json:"model,omitempty"`
}

// UpdateSessionRequest represents mutable attributes of a session.
type UpdateSessionRequest struct {
	Title    *string `json:"title,omitempty"`
	IsPinned *bool   `json:"is_pinned,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// ForkSessionRequest represents the payload to fork a conversation.
type ForkSessionRequest struct {
	NewID         string `json:"new_id,omitempty"`
	Title         string `json:"title,omitempty"`
	UpToMessageID string `json:"up_to_message_id,omitempty"`
}

// UpdateSettingsRequest defines the payload structure for patching system configuration.
type UpdateSettingsRequest struct {
	Auth *struct {
		AIProvider      *string `json:"ai_provider"`
		SectorsAPIKey   *string `json:"sectors_api_key"`
		SectorsBaseURL  *string `json:"sectors_base_url"`
		GeminiAPIKey    *string `json:"gemini_api_key"`
		GeminiModel     *string `json:"gemini_model"`
		OpenAIAPIKey    *string `json:"openai_api_key"`
		OpenAIBaseURL   *string `json:"openai_base_url"`
		OpenAIModel     *string `json:"openai_model"`
		AnthropicAPIKey *string `json:"anthropic_api_key"`
		OllamaBaseURL   *string `json:"ollama_base_url"`
		OllamaModel     *string `json:"ollama_model"`
	} `json:"auth"`
	Preferences *struct {
		DefaultMarket  *string  `json:"default_market"`
		OfflineMode    *bool    `json:"offline_mode"`
		Language       *string  `json:"language"`
		LLMTimeoutSecs *float64 `json:"llm_timeout_secs"`
	} `json:"preferences"`
	Storage *struct {
		DBPath *string `json:"db_path"`
	} `json:"storage"`
	Engine *struct {
		PythonBin  *string `json:"python_bin"`
		EnginePath *string `json:"entrypoint"`
	} `json:"engine"`
	Memory *struct {
		Enabled          *bool    `json:"enabled"`
		DecayLambda      *float64 `json:"decay_lambda"`
		EgoRadius        *int     `json:"ego_radius"`
		MaxContextTokens *int     `json:"max_context_tokens"`
	} `json:"memory"`
	Telegram *struct {
		BotToken     *string   `json:"bot_token"`
		Enabled      *bool     `json:"enabled"`
		AllowedUsers *[]string `json:"allowed_users"`
	} `json:"telegram"`
}

// Start launches the background HTTP server on the specified port (or auto-finds free port).
func Start(ctx context.Context, requestedPort int, database *db.DB, cfg *config.Config) (*Server, error) {
	mux := http.NewServeMux()

	activeCfg := cfg
	if activeCfg == nil {
		activeCfg = config.DefaultConfig()
	}

	s := &Server{
		Port:           requestedPort,
		DB:             database,
		SessionManager: NewSessionManager(),
		Config:         activeCfg,
	}

	if activeCfg.Telegram.Enabled && strings.TrimSpace(activeCfg.Telegram.BotToken) != "" {
		if botSvc, err := telegram.NewBotService(s.Config, s.DB, s.SessionManager); err == nil {
			s.BotService = botSvc
			_ = s.BotService.Start()
		}
	}

	// Helper for CORS preflight and headers
	enableCORS := func(w http.ResponseWriter, r *http.Request) bool {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return true
		}
		return false
	}

	// Helper to send JSON responses
	sendJSON := func(w http.ResponseWriter, status int, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(data)
	}

	// 1. Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"app":       "Niskava Agent",
			"version":   "1.0.0",
			"market":    "IDX",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Settings API endpoint (GET masked view, PATCH/PUT update & hot-reload)
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}

		s.cfgMu.RLock()
		activeCfg := s.Config
		s.cfgMu.RUnlock()

		if activeCfg == nil {
			activeCfg = config.DefaultConfig()
		}

		if r.Method == http.MethodGet {
			sendJSON(w, http.StatusOK, activeCfg.MaskedView())
			return
		}

		if r.Method == http.MethodPatch || r.Method == http.MethodPut {
			var req UpdateSettingsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "invalid json: %v"}`, err), http.StatusBadRequest)
				return
			}

			s.cfgMu.Lock()
			if s.Config == nil {
				s.Config = config.DefaultConfig()
			}

			if req.Auth != nil {
				if req.Auth.AIProvider != nil && *req.Auth.AIProvider != "" {
					s.Config.Auth.AIProvider = *req.Auth.AIProvider
				}
				if req.Auth.SectorsAPIKey != nil {
					if *req.Auth.SectorsAPIKey == "" {
						s.Config.Auth.SectorsAPIKey = ""
					} else if !strings.Contains(*req.Auth.SectorsAPIKey, "****") {
						s.Config.Auth.SectorsAPIKey = *req.Auth.SectorsAPIKey
					}
				}
				if req.Auth.SectorsBaseURL != nil && *req.Auth.SectorsBaseURL != "" {
					s.Config.Auth.SectorsBaseURL = *req.Auth.SectorsBaseURL
				}
				if req.Auth.GeminiAPIKey != nil {
					if *req.Auth.GeminiAPIKey == "" {
						s.Config.Auth.GeminiAPIKey = ""
					} else if !strings.Contains(*req.Auth.GeminiAPIKey, "****") {
						s.Config.Auth.GeminiAPIKey = *req.Auth.GeminiAPIKey
					}
				}
				if req.Auth.GeminiModel != nil && *req.Auth.GeminiModel != "" {
					s.Config.Auth.GeminiModel = *req.Auth.GeminiModel
				}
				if req.Auth.OpenAIAPIKey != nil {
					if *req.Auth.OpenAIAPIKey == "" {
						s.Config.Auth.OpenAIAPIKey = ""
					} else if !strings.Contains(*req.Auth.OpenAIAPIKey, "****") {
						s.Config.Auth.OpenAIAPIKey = *req.Auth.OpenAIAPIKey
					}
				}
				if req.Auth.OpenAIBaseURL != nil && *req.Auth.OpenAIBaseURL != "" {
					s.Config.Auth.OpenAIBaseURL = *req.Auth.OpenAIBaseURL
				}
				if req.Auth.OpenAIModel != nil && *req.Auth.OpenAIModel != "" {
					s.Config.Auth.OpenAIModel = *req.Auth.OpenAIModel
				}
				if req.Auth.AnthropicAPIKey != nil {
					if *req.Auth.AnthropicAPIKey == "" {
						s.Config.Auth.AnthropicAPIKey = ""
					} else if !strings.Contains(*req.Auth.AnthropicAPIKey, "****") {
						s.Config.Auth.AnthropicAPIKey = *req.Auth.AnthropicAPIKey
					}
				}
				if req.Auth.OllamaBaseURL != nil && *req.Auth.OllamaBaseURL != "" {
					s.Config.Auth.OllamaBaseURL = *req.Auth.OllamaBaseURL
				}
				if req.Auth.OllamaModel != nil && *req.Auth.OllamaModel != "" {
					s.Config.Auth.OllamaModel = *req.Auth.OllamaModel
				}
			}

			if req.Preferences != nil {
				if req.Preferences.DefaultMarket != nil && *req.Preferences.DefaultMarket != "" {
					s.Config.Preferences.DefaultMarket = strings.ToUpper(*req.Preferences.DefaultMarket)
				}
				if req.Preferences.OfflineMode != nil {
					s.Config.Preferences.OfflineMode = *req.Preferences.OfflineMode
				}
				if req.Preferences.Language != nil && *req.Preferences.Language != "" {
					s.Config.Preferences.Language = strings.ToLower(*req.Preferences.Language)
				}
				if req.Preferences.LLMTimeoutSecs != nil {
					v := *req.Preferences.LLMTimeoutSecs
					if v < 10.0 {
						v = 10.0
					}
					if v > 300.0 {
						v = 300.0
					}
					s.Config.Preferences.LLMTimeoutSecs = v
				}
			}

			if req.Storage != nil && req.Storage.DBPath != nil && *req.Storage.DBPath != "" {
				s.Config.Storage.DBPath = *req.Storage.DBPath
			}
			if req.Engine != nil {
				if req.Engine.PythonBin != nil && *req.Engine.PythonBin != "" {
					s.Config.Engine.PythonBin = *req.Engine.PythonBin
				}
				if req.Engine.EnginePath != nil && *req.Engine.EnginePath != "" {
					s.Config.Engine.EnginePath = *req.Engine.EnginePath
				}
			}
			if req.Memory != nil {
				if req.Memory.Enabled != nil {
					s.Config.Memory.Enabled = *req.Memory.Enabled
				}
				if req.Memory.DecayLambda != nil {
					s.Config.Memory.DecayLambda = *req.Memory.DecayLambda
				}
				if req.Memory.EgoRadius != nil {
					s.Config.Memory.EgoRadius = *req.Memory.EgoRadius
				}
				if req.Memory.MaxContextTokens != nil {
					s.Config.Memory.MaxContextTokens = *req.Memory.MaxContextTokens
				}
			}
			telegramUpdated := false
			if req.Telegram != nil {
				telegramUpdated = true
				if req.Telegram.BotToken != nil {
					if *req.Telegram.BotToken == "" {
						s.Config.Telegram.BotToken = ""
					} else if !strings.Contains(*req.Telegram.BotToken, "****") {
						s.Config.Telegram.BotToken = *req.Telegram.BotToken
					}
				}
				if req.Telegram.Enabled != nil {
					s.Config.Telegram.Enabled = *req.Telegram.Enabled
				}
				if req.Telegram.AllowedUsers != nil {
					s.Config.Telegram.AllowedUsers = *req.Telegram.AllowedUsers
				}
			}

			_ = config.SaveConfig(s.Config, s.ConfigPath)
			view := s.Config.MaskedView()
			s.cfgMu.Unlock()

			if telegramUpdated {
				s.syncTelegramBotState()
			}

			sendJSON(w, http.StatusOK, view)
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	// Test Connection endpoint
	mux.HandleFunc("/api/settings/test-connection", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Target  string `json:"target"`
			APIKey  string `json:"api_key,omitempty"`
			BaseURL string `json:"base_url,omitempty"`
			Model   string `json:"model,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "invalid json: %v"}`, err), http.StatusBadRequest)
			return
		}

		target := strings.ToLower(strings.TrimSpace(req.Target))
		validTargets := map[string]bool{"sectors": true, "gemini": true, "openai": true, "ollama": true, "anthropic": true}
		if !validTargets[target] {
			http.Error(w, `{"error": "invalid target, must be one of: sectors, gemini, openai, ollama, anthropic"}`, http.StatusBadRequest)
			return
		}

		start := time.Now()
		type responseType struct {
			Target    string `json:"target"`
			Success   bool   `json:"success"`
			Message   string `json:"message"`
			LatencyMs int64  `json:"latency_ms"`
		}
		resp := responseType{
			Target: target,
		}

		s.cfgMu.RLock()
		cfg := s.Config
		s.cfgMu.RUnlock()
		if cfg == nil {
			cfg = config.DefaultConfig()
		}

		switch target {
		case "sectors":
			key := req.APIKey
			if key == "" || strings.Contains(key, "****") {
				key = cfg.Auth.SectorsAPIKey
			}
			if key == "" && !cfg.Preferences.OfflineMode && os.Getenv("MOCK_SECTORS") != "1" {
				resp.Success = false
				resp.Message = "Sectors API key is not configured"
				break
			}
			if cfg.Preferences.OfflineMode || os.Getenv("MOCK_SECTORS") == "1" {
				resp.Success = true
				resp.Message = "Sectors mock mode active (offline testing)"
				break
			}

			client := &http.Client{Timeout: 5 * time.Second}
			httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://api.sectors.app/v2/daily/BBCA/?format=json", nil)
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Failed to build request: %v", err)
				break
			}
			httpReq.Header.Set("Authorization", key)
			httpResp, err := client.Do(httpReq)
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Connection failed: %v", err)
				break
			}
			defer httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusOK {
				resp.Success = true
				resp.Message = "Connected to Sectors v2 API successfully"
			} else if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden {
				resp.Success = false
				resp.Message = "Invalid Sectors API key (Unauthorized)"
			} else {
				resp.Success = false
				resp.Message = fmt.Sprintf("Sectors API returned HTTP %d", httpResp.StatusCode)
			}

		case "ollama":
			baseURL := req.BaseURL
			if baseURL == "" {
				baseURL = cfg.Auth.OllamaBaseURL
			}
			if baseURL == "" {
				baseURL = "http://localhost:11434"
			}
			baseURL = strings.TrimRight(baseURL, "/")

			client := &http.Client{Timeout: 3 * time.Second}
			httpResp, err := client.Get(baseURL + "/api/tags")
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Failed to reach Ollama endpoint at %s: %v", baseURL, err)
				break
			}
			defer httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusOK {
				resp.Success = true
				resp.Message = fmt.Sprintf("Ollama instance reached at %s", baseURL)
			} else {
				resp.Success = false
				resp.Message = fmt.Sprintf("Ollama returned HTTP %d", httpResp.StatusCode)
			}

		case "gemini":
			key := req.APIKey
			if key == "" || strings.Contains(key, "****") {
				key = cfg.Auth.GeminiAPIKey
			}
			if key == "" {
				resp.Success = false
				resp.Message = "Gemini API key is not configured"
				break
			}
			if cfg.Preferences.OfflineMode {
				resp.Success = true
				resp.Message = "Offline mode active (mock verification)"
				break
			}
			client := &http.Client{Timeout: 5 * time.Second}
			testURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", key)
			httpResp, err := client.Get(testURL)
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Connection to Gemini failed: %v", err)
				break
			}
			defer httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusOK {
				resp.Success = true
				resp.Message = "Gemini API key verified successfully"
			} else {
				resp.Success = false
				resp.Message = fmt.Sprintf("Gemini returned HTTP %d", httpResp.StatusCode)
			}

		case "openai":
			key := req.APIKey
			if key == "" || strings.Contains(key, "****") {
				key = cfg.Auth.OpenAIAPIKey
			}
			baseURL := req.BaseURL
			if baseURL == "" {
				baseURL = cfg.Auth.OpenAIBaseURL
			}
			if baseURL == "" {
				baseURL = "https://api.openai.com/v1"
			}
			baseURL = strings.TrimRight(baseURL, "/")
			if cfg.Preferences.OfflineMode {
				resp.Success = true
				resp.Message = "Offline mode active (mock verification)"
				break
			}
			client := &http.Client{Timeout: 5 * time.Second}
			httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, baseURL+"/models", nil)
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Failed to build request: %v", err)
				break
			}
			if key != "" {
				httpReq.Header.Set("Authorization", "Bearer "+key)
			}
			httpResp, err := client.Do(httpReq)
			if err != nil {
				resp.Success = false
				resp.Message = fmt.Sprintf("Connection failed: %v", err)
				break
			}
			defer httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusOK {
				resp.Success = true
				resp.Message = "OpenAI compatible endpoint reached successfully"
			} else {
				resp.Success = false
				resp.Message = fmt.Sprintf("Endpoint returned HTTP %d", httpResp.StatusCode)
			}

		case "anthropic":
			key := req.APIKey
			if key == "" || strings.Contains(key, "****") {
				key = cfg.Auth.AnthropicAPIKey
			}
			if key == "" {
				resp.Success = false
				resp.Message = "Anthropic API key is not configured"
				break
			}
			resp.Success = true
			resp.Message = "Anthropic key format verified"
		}

		resp.LatencyMs = time.Since(start).Milliseconds()
		sendJSON(w, http.StatusOK, resp)
	})

	// Telegram Settings endpoint (GET view, PATCH/PUT update)
	mux.HandleFunc("/api/settings/telegram", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}

		if r.Method == http.MethodGet {
			s.cfgMu.RLock()
			cfg := s.Config
			var enabled bool
			var hasToken bool
			var allowedUsers []string
			if cfg != nil {
				enabled = cfg.Telegram.Enabled
				hasToken = strings.TrimSpace(cfg.Telegram.BotToken) != ""
				allowedUsers = cfg.Telegram.AllowedUsers
			}
			s.cfgMu.RUnlock()

			if allowedUsers == nil {
				allowedUsers = []string{}
			}

			s.botMu.Lock()
			botRunning := s.BotService != nil && s.BotService.IsStarted()
			var botUsername string
			if s.BotService != nil {
				botUsername = s.BotService.BotUsername()
			}
			s.botMu.Unlock()

			status := "NOT_CONFIGURED"
			if botRunning {
				status = "RUNNING"
			} else if hasToken {
				status = "STOPPED"
			}

			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":        status,
				"bot_username":  botUsername,
				"enabled":       enabled,
				"allowed_users": allowedUsers,
				"has_token":     hasToken,
			})
			return
		}

		if r.Method == http.MethodPatch || r.Method == http.MethodPut {
			var req struct {
				BotToken     *string       `json:"bot_token"`
				Enabled      *bool         `json:"enabled"`
				AllowedUsers []interface{} `json:"allowed_users"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "invalid json: %v"}`, err), http.StatusBadRequest)
				return
			}

			s.cfgMu.Lock()
			if s.Config == nil {
				s.Config = config.DefaultConfig()
			}
			if req.BotToken != nil {
				if *req.BotToken == "" {
					s.Config.Telegram.BotToken = ""
				} else if !strings.Contains(*req.BotToken, "****") {
					s.Config.Telegram.BotToken = *req.BotToken
				}
			}
			if req.Enabled != nil {
				s.Config.Telegram.Enabled = *req.Enabled
			}
			if req.AllowedUsers != nil {
				var users []string
				for _, u := range req.AllowedUsers {
					switch val := u.(type) {
					case string:
						trimmed := strings.TrimSpace(val)
						if trimmed != "" {
							users = append(users, trimmed)
						}
					case float64:
						users = append(users, fmt.Sprintf("%.0f", val))
					}
				}
				s.Config.Telegram.AllowedUsers = users
			}
			_ = config.SaveConfig(s.Config, s.ConfigPath)

			enabled := s.Config.Telegram.Enabled
			hasToken := strings.TrimSpace(s.Config.Telegram.BotToken) != ""
			allowedUsers := s.Config.Telegram.AllowedUsers
			s.cfgMu.Unlock()

			s.syncTelegramBotState()

			if allowedUsers == nil {
				allowedUsers = []string{}
			}

			s.botMu.Lock()
			botRunning := s.BotService != nil && s.BotService.IsStarted()
			var botUsername string
			if s.BotService != nil {
				botUsername = s.BotService.BotUsername()
			}
			s.botMu.Unlock()

			status := "NOT_CONFIGURED"
			if botRunning {
				status = "RUNNING"
			} else if hasToken {
				status = "STOPPED"
			}

			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":        status,
				"bot_username":  botUsername,
				"enabled":       enabled,
				"allowed_users": allowedUsers,
				"has_token":     hasToken,
			})
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	// Telegram Daemon Lifecycle Start endpoint
	mux.HandleFunc("/api/telegram/start", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		s.cfgMu.RLock()
		token := ""
		if s.Config != nil {
			token = strings.TrimSpace(s.Config.Telegram.BotToken)
		}
		s.cfgMu.RUnlock()

		if token == "" {
			sendJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":  "telegram bot token is not configured",
				"status": "NOT_CONFIGURED",
			})
			return
		}

		s.botMu.Lock()
		defer s.botMu.Unlock()

		if s.BotService == nil || !s.BotService.IsStarted() || s.BotService.Token() != token {
			if s.BotService != nil && s.BotService.IsStarted() {
				s.BotService.Stop()
				s.BotService = nil
			}
			svc, err := telegram.NewBotService(s.Config, s.DB, s.SessionManager)
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"error":  fmt.Sprintf("failed to initialize telegram bot: %v", err),
					"status": "STOPPED",
				})
				return
			}
			s.BotService = svc
			if err := s.BotService.Start(); err != nil {
				sendJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"error":  fmt.Sprintf("failed to start telegram bot: %v", err),
					"status": "STOPPED",
				})
				return
			}
		}

		s.cfgMu.Lock()
		if s.Config != nil {
			s.Config.Telegram.Enabled = true
			_ = config.SaveConfig(s.Config, s.ConfigPath)
		}
		s.cfgMu.Unlock()

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "RUNNING",
			"bot_username": s.BotService.BotUsername(),
			"message":      "telegram bot started successfully",
		})
	})

	// Telegram Daemon Lifecycle Stop endpoint
	mux.HandleFunc("/api/telegram/stop", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		s.botMu.Lock()
		if s.BotService != nil && s.BotService.IsStarted() {
			s.BotService.Stop()
		}
		s.botMu.Unlock()

		s.cfgMu.Lock()
		if s.Config != nil {
			s.Config.Telegram.Enabled = false
			_ = config.SaveConfig(s.Config, s.ConfigPath)
		}
		s.cfgMu.Unlock()

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "STOPPED",
			"message": "telegram bot stopped successfully",
		})
	})

	// Telegram Test Message endpoint
	mux.HandleFunc("/api/telegram/test", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ChatID interface{} `json:"chat_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "invalid json: %v"}`, err), http.StatusBadRequest)
			return
		}

		var chatID int64
		switch v := req.ChatID.(type) {
		case float64:
			chatID = int64(v)
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err == nil {
				chatID = parsed
			}
		}

		if chatID == 0 {
			http.Error(w, `{"error": "chat_id is required"}`, http.StatusBadRequest)
			return
		}

		s.cfgMu.RLock()
		isOffline := false
		if s.Config != nil {
			isOffline = s.Config.Preferences.OfflineMode
		}
		s.cfgMu.RUnlock()

		if isOffline || os.Getenv("TELEGRAM_OFFLINE") == "1" {
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "ok",
				"mock":    true,
				"chat_id": chatID,
				"message": "test message delivered (mock offline mode)",
			})
			return
		}

		s.botMu.Lock()
		active := s.BotService != nil && s.BotService.IsStarted()
		svc := s.BotService
		s.botMu.Unlock()

		if !active || svc == nil {
			http.Error(w, `{"error": "telegram bot is not running", "status": "STOPPED"}`, http.StatusBadRequest)
			return
		}

		testMsg := "🔔 Niskava Agent: Koneksi Telegram Bot berhasil diverifikasi."
		if err := svc.SendTestMessage(chatID, testMsg); err != nil {
			sendJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"error":   fmt.Sprintf("failed to send telegram test message: %v", err),
				"chat_id": chatID,
			})
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"chat_id": chatID,
			"message": "test message sent successfully",
		})
	})

	// System: Sectors API Credit & Cache Usage endpoint (Law 5 compliance)
	mux.HandleFunc("/api/system/sectors-usage", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		stats, err := database.GetSectorsCacheStats()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"total_entries":          stats.TotalEntries,
			"expired_entries":        stats.ExpiredEntries,
			"permanent_entries":      stats.PermanentEntries,
			"estimated_credit_saved": stats.TotalEntries,
			"status":                 "HEALTHY",
		})
	})

	// System: Diagnostics and runtime health metrics
	mux.HandleFunc("/api/system/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}

		dbPath := ""
		var dbSizeBytes int64
		if database != nil && database.Path != "" {
			dbPath = database.Path
			if fi, err := os.Stat(dbPath); err == nil {
				dbSizeBytes = fi.Size()
			}
		}

		totalSessions := 0
		if database != nil {
			_, total, err := database.ListChatSessions(1, 0, "")
			if err == nil {
				totalSessions = total
			}
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":              "OK",
			"app":                 "Niskava Agent",
			"go_version":          runtime.Version(),
			"os":                  runtime.GOOS,
			"arch":                runtime.GOARCH,
			"num_cpu":             runtime.NumCPU(),
			"database_path":       dbPath,
			"database_size_bytes": dbSizeBytes,
			"total_sessions":      totalSessions,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
		})
	})

	// System: Clean expired cache entries
	mux.HandleFunc("/api/system/cache/clean", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		cleaned, err := database.CleanExpiredCache()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "failed to clean expired cache: %v"}`, err), http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":          "ok",
			"cleaned_entries": cleaned,
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 2. Chat Sessions Collection API (GET list, POST create)
	mux.HandleFunc("/api/chat/sessions", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			limit := 50
			offset := 0
			if lStr := r.URL.Query().Get("limit"); lStr != "" {
				if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			if oStr := r.URL.Query().Get("offset"); oStr != "" {
				if parsed, err := strconv.Atoi(oStr); err == nil && parsed >= 0 {
					offset = parsed
				}
			}
			search := r.URL.Query().Get("q")
			sessions, total, err := database.ListChatSessions(limit, offset, search)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			if sessions == nil {
				sessions = []db.ChatSession{}
			}
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"total":    total,
				"limit":    limit,
				"offset":   offset,
				"sessions": sessions,
			})

		case http.MethodPost:
			var req CreateSessionRequest
			if r.Body != nil && r.ContentLength > 0 {
				_ = json.NewDecoder(r.Body).Decode(&req)
			}
			if req.ID == "" {
				req.ID = fmt.Sprintf("CHAT-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
			}
			if req.Title == "" {
				req.Title = "Sesi Riset Pasar"
			}
			if req.Model == "" {
				req.Model = "hermes"
			}
			sess := &db.ChatSession{
				ID:     req.ID,
				Title:  req.Title,
				Model:  req.Model,
				Status: "IDLE",
			}
			if err := database.CreateChatSession(sess); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			sendJSON(w, http.StatusCreated, map[string]interface{}{
				"status":  "created",
				"session": sess,
			})

		default:
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// 2b. Global Chat Message Search API
	mux.HandleFunc("/api/chat/search", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "GET required"}`, http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query().Get("q")
		limit := 50
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		results, err := database.SearchChatMessages(query, limit)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		if results == nil {
			results = []db.ChatSearchResult{}
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"query":   query,
			"total":   len(results),
			"results": results,
		})
	})

	// 3. Chat Session Item & Actions API (/api/chat/sessions/{id} and subpaths)
	mux.HandleFunc("/api/chat/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		subPath := strings.TrimPrefix(r.URL.Path, "/api/chat/sessions/")
		subPath = strings.Trim(subPath, "/")
		parts := strings.Split(subPath, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, `{"error": "session ID required"}`, http.StatusBadRequest)
			return
		}

		sessionID := parts[0]

		// Singular session item operations: /api/chat/sessions/{id}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				sess, err := database.GetChatSession(sessionID)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
					return
				}
				if sess == nil {
					http.Error(w, `{"error": "session not found"}`, http.StatusNotFound)
					return
				}
				sendJSON(w, http.StatusOK, map[string]interface{}{
					"session": sess,
				})

			case http.MethodPatch:
				var req UpdateSessionRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
					return
				}
				if err := database.UpdateChatSession(sessionID, req.Title, req.IsPinned, req.Status); err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
					return
				}
				updated, _ := database.GetChatSession(sessionID)
				sendJSON(w, http.StatusOK, map[string]interface{}{
					"status":  "updated",
					"session": updated,
				})

			case http.MethodDelete:
				// If currently streaming, abort first
				_ = s.SessionManager.Abort(sessionID)
				if err := database.DeleteChatSession(sessionID); err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
					return
				}
				sendJSON(w, http.StatusOK, map[string]interface{}{
					"status":     "deleted",
					"session_id": sessionID,
				})

			default:
				http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
			}
			return
		}

		// Action subpaths: /api/chat/sessions/{id}/{action}
		action := parts[1]
		switch action {
		case "fork":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error": "POST required"}`, http.StatusMethodNotAllowed)
				return
			}
			var req ForkSessionRequest
			if r.Body != nil && r.ContentLength > 0 {
				_ = json.NewDecoder(r.Body).Decode(&req)
			}
			if req.NewID == "" {
				req.NewID = fmt.Sprintf("CHAT-%s-FORK-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
			}
			if err := database.ForkChatSession(sessionID, req.NewID, req.Title, req.UpToMessageID); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			forkedSess, _ := database.GetChatSession(req.NewID)
			sendJSON(w, http.StatusCreated, map[string]interface{}{
				"status":  "forked",
				"session": forkedSess,
			})

		case "abort":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error": "POST required"}`, http.StatusMethodNotAllowed)
				return
			}
			aborted := s.SessionManager.Abort(sessionID)
			idle := "IDLE"
			_ = database.UpdateChatSession(sessionID, nil, nil, &idle)
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":      "aborted",
				"session_id":  sessionID,
				"was_running": aborted,
			})

		case "reset":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error": "POST required"}`, http.StatusMethodNotAllowed)
				return
			}
			_ = s.SessionManager.Abort(sessionID)
			if err := database.ClearSessionHistory(sessionID); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":     "reset",
				"session_id": sessionID,
			})

		case "messages":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error": "GET required"}`, http.StatusMethodNotAllowed)
				return
			}
			limit := 100
			if lStr := r.URL.Query().Get("limit"); lStr != "" {
				if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			messages, err := database.GetChatHistory(sessionID, limit)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			if messages == nil {
				messages = []db.ChatMessage{}
			}
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"session_id": sessionID,
				"total":      len(messages),
				"messages":   messages,
			})

		case "export":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error": "GET required"}`, http.StatusMethodNotAllowed)
				return
			}
			sess, err := database.GetChatSession(sessionID)
			if err != nil || sess == nil {
				http.Error(w, `{"error": "session not found"}`, http.StatusNotFound)
				return
			}
			messages, err := database.GetChatHistory(sessionID, 500)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}

			format := strings.ToLower(r.URL.Query().Get("format"))
			if format == "json" {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="niskava-session-%s.json"`, sessionID))
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"session":  sess,
					"messages": messages,
				})
				return
			}

			// Default to structured Markdown report (Law 2 compliant)
			var md strings.Builder
			md.WriteString(fmt.Sprintf("# Laporan Riset Pasar: %s\n\n", sess.Title))
			md.WriteString("| Parameter | Nilai |\n")
			md.WriteString("|---|---|\n")
			md.WriteString(fmt.Sprintf("| **Session ID** | `%s` |\n", sess.ID))
			md.WriteString(fmt.Sprintf("| **Model AI** | `%s` |\n", sess.Model))
			md.WriteString(fmt.Sprintf("| **Total Pesan** | %d |\n", len(messages)))
			md.WriteString(fmt.Sprintf("| **Waktu Ekspor** | %s |\n\n", time.Now().UTC().Format(time.RFC3339)))

			// Law 2: Strict Financial Non-Advisory Boundary
			md.WriteString("> [!IMPORTANT]\n")
			md.WriteString("> **DISCLAIMER (Non-Advisory Market Intelligence):**\n")
			md.WriteString("> Niskava Agent adalah platform intelijen dan riset pasar modal otonom untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi atau broker berizin. Seluruh temuan, skor anomali, dan korelasi bukti disajikan secara deskriptif untuk tujuan riset dan verifikasi fakta, serta BUKAN merupakan rekomendasi beli/jual atau target harga investasi.\n\n")

			md.WriteString("## Transkrip Percakapan & Temuan Riset\n\n")
			for idx, msg := range messages {
				timeStr := msg.CreatedAt
				if len(timeStr) > 19 {
					timeStr = strings.Replace(timeStr[:19], "T", " ", 1)
				}
				if msg.Role == "user" {
					md.WriteString(fmt.Sprintf("### 👤 Pengguna (Turn %d) — *%s*\n\n", (idx/2)+1, timeStr))
					md.WriteString(fmt.Sprintf("%s\n\n", msg.Content))
				} else {
					md.WriteString(fmt.Sprintf("### 🤖 Niskava Agent (%s) — *%s*\n\n", sess.Model, timeStr))
					if msg.Thought != nil && strings.TrimSpace(*msg.Thought) != "" {
						md.WriteString("<details>\n<summary>🔍 Proses Berpikir Analitis (Chain-of-Thought)</summary>\n\n")
						md.WriteString(fmt.Sprintf("%s\n\n", *msg.Thought))
						md.WriteString("</details>\n\n")
					}
					md.WriteString(fmt.Sprintf("%s\n\n", msg.Content))
				}
				md.WriteString("---\n\n")
			}

			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="niskava-session-%s.md"`, sessionID))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(md.String()))
			return

		default:
			http.Error(w, `{"error": "unknown session action"}`, http.StatusNotFound)
		}
	})

	// 4. Investigations sessions list endpoint (pipeline audit sessions)
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessions, err := database.ListInvestigations(50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(sessions),
			"data":  sessions,
		})
	})

	// 4b. Investigation Details, Anomalies, and Findings API (/api/investigations/{id} and subpaths)
	mux.HandleFunc("/api/investigations/", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		subPath := strings.TrimPrefix(r.URL.Path, "/api/investigations/")
		subPath = strings.Trim(subPath, "/")
		parts := strings.Split(subPath, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, `{"error": "investigation ID required"}`, http.StatusBadRequest)
			return
		}

		sessionID := parts[0]

		if len(parts) == 1 {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			inv, err := database.GetInvestigation(sessionID)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
			if inv == nil {
				http.Error(w, `{"error": "investigation not found"}`, http.StatusNotFound)
				return
			}
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"investigation": inv,
			})
			return
		}

		if len(parts) == 2 {
			action := parts[1]
			switch action {
			case "anomalies":
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				anomalies, err := database.GetAnomaliesByInvestigation(sessionID)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
					return
				}
				sendJSON(w, http.StatusOK, map[string]interface{}{
					"investigation_id": sessionID,
					"total":            len(anomalies),
					"anomalies":        anomalies,
				})
				return
			case "findings":
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				findings, err := database.ListFindingsByInvestigation(sessionID)
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
					return
				}
				sendJSON(w, http.StatusOK, map[string]interface{}{
					"investigation_id": sessionID,
					"total":            len(findings),
					"findings":         findings,
				})
				return
			default:
				http.Error(w, `{"error": "unknown investigation sub-resource"}`, http.StatusNotFound)
				return
			}
		}

		http.Error(w, `{"error": "unknown investigation sub-resource"}`, http.StatusNotFound)
	})

	// Also allow /api/investigations to list investigations (alias to /api/sessions)
	mux.HandleFunc("/api/investigations", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessions, err := database.ListInvestigations(50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(sessions),
			"data":  sessions,
		})
	})

	// 5. Chat History endpoint (backward compatible)
	mux.HandleFunc("/api/chat/history", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, `{"error": "session_id parameter required"}`, http.StatusBadRequest)
			return
		}

		history, err := database.GetChatHistory(sessionID, 50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"session_id": sessionID,
			"total":      len(history),
			"messages":   history,
		})
	})

	// 5b. Chat Reset endpoint (supports single session reset or clearing all sessions)
	mux.HandleFunc("/api/chat/reset", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "POST required"}`, http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		if sessionID != "" {
			_ = s.SessionManager.Abort(sessionID)
			if err := database.DeleteChatSession(sessionID); err != nil {
				_ = database.ClearSessionHistory(sessionID)
			}
			sendJSON(w, http.StatusOK, map[string]interface{}{
				"status":     "reset",
				"session_id": sessionID,
			})
			return
		}

		// Reset all sessions and messages
		s.SessionManager.AbortAll()
		if err := database.ClearAllChatSessions(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		sendJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "all_reset",
			"message": "all chat history and sessions cleared",
		})
	})

	// 6. Conversational Chat SSE Streaming endpoint
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.Prompt) == "" {
			http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
			return
		}

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = fmt.Sprintf("WEB-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
		}

		// Concurrency protection: reject if session is already running
		if s.SessionManager.IsBusy(sessionID) {
			http.Error(w, `{"error": "session is currently processing another turn"}`, http.StatusConflict)
			return
		}

		// Ensure chat session exists in database
		if database != nil {
			sess, _ := database.GetChatSession(sessionID)
			if sess == nil {
				sessionTitle := strings.TrimSpace(req.Prompt)
				if len(sessionTitle) > 42 {
					sessionTitle = sessionTitle[:39] + "..."
				}
				if sessionTitle == "" {
					sessionTitle = "Sesi Riset Pasar"
				}
				_ = database.CreateChatSession(&db.ChatSession{
					ID:     sessionID,
					Title:  sessionTitle,
					Model:  "hermes",
					Status: "BUSY",
				})
			} else {
				busyStatus := "BUSY"
				_ = database.UpdateChatSession(sessionID, nil, nil, &busyStatus)
			}
		}

		// Save User Message
		if database != nil {
			_ = database.SaveChatMessage(&db.ChatMessage{
				ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
				SessionID: sessionID,
				Role:      "user",
				Content:   req.Prompt,
				Status:    "COMPLETED",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			})
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Disable read and write deadlines for long-running SSE streaming sessions
		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})
		_ = rc.SetReadDeadline(time.Time{})

		// Create cancellable context for this chat execution
		chatCtx, cancelChat := context.WithCancel(r.Context())
		defer cancelChat()

		if !s.SessionManager.Register(sessionID, cancelChat) {
			http.Error(w, `{"error": "session is currently busy"}`, http.StatusConflict)
			return
		}
		defer s.SessionManager.Unregister(sessionID)

		defer func() {
			if database != nil {
				idleStatus := "IDLE"
				_ = database.UpdateChatSession(sessionID, nil, nil, &idleStatus)
			}
		}()

		s.cfgMu.RLock()
		activeCfg := s.Config
		s.cfgMu.RUnlock()
		if activeCfg == nil {
			activeCfg = config.DefaultConfig()
		}

		configuredBin := activeCfg.Engine.PythonBin
		if envBin := os.Getenv("NISKAVA_PYTHON_BIN"); envBin != "" {
			configuredBin = envBin
		}
		pythonBin := ipc.ResolvePythonBin(configuredBin)

		dbPath := ""
		if s.DB != nil && s.DB.Path != "" {
			dbPath = s.DB.Path
		} else if activeCfg.Storage.DBPath != "" {
			dbPath = config.ExpandHome(activeCfg.Storage.DBPath)
		} else {
			homeDir, _ := os.UserHomeDir()
			dbPath = filepath.Join(homeDir, ".niskava", "niskava.db")
			if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
				dbPath = customDB
			}
		}

		wd, _ := os.Getwd()

		chatLang := activeCfg.Preferences.Language
		if envLang := os.Getenv("NISKAVA_LANG"); envLang != "" {
			chatLang = envLang
		}
		if chatLang == "" {
			chatLang = "id"
		}

		isOffline := activeCfg.Preferences.OfflineMode || os.Getenv("NISKAVA_OFFLINE") == "1" || os.Getenv("MOCK_SECTORS") == "1"

		runnerParams := ipc.RunnerParams{
			PythonBin:    pythonBin,
			WorkDir:      wd,
			DBPath:       dbPath,
			Prompt:       req.Prompt,
			SessionID:    sessionID,
			Offline:      isOffline,
			Language:     chatLang,
			EnvOverrides: s.buildSubprocessEnv(),
		}

		eventsChan, errChan := ipc.RunSubprocess(chatCtx, runnerParams)

		// Periodic keep-alive comment to prevent proxy/socket timeouts during long multi-tool LLM turns
		keepAliveTicker := time.NewTicker(15 * time.Second)
		defer keepAliveTicker.Stop()

		var assistantResponse strings.Builder
		wasAborted := false

		for {
			select {
			case <-keepAliveTicker.C:
				// SSE comment line: keep-alive (ignored by event parsers, prevents socket idle death)
				fmt.Fprintf(w, ": keep-alive\n\n")
				flusher.Flush()

			case <-chatCtx.Done():
				wasAborted = true
				fmt.Fprintf(w, "event: session_error\ndata: {\"error\": \"execution aborted by user or context cancelled\"}\n\n")
				flusher.Flush()
				if database != nil && assistantResponse.Len() > 0 {
					_ = database.SaveChatMessage(&db.ChatMessage{
						ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
						SessionID: sessionID,
						Role:      "assistant",
						Content:   assistantResponse.String() + " [Aborted]",
						Status:    "ABORTED",
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
					})
				}
				return

			case err, ok := <-errChan:
				if !ok {
					errChan = nil
					continue
				}
				if err != nil && !wasAborted {
					errPayload, _ := json.Marshal(map[string]interface{}{
						"event":      "session_error",
						"session_id": sessionID,
						"error":      err.Error(),
					})
					fmt.Fprintf(w, "event: session_error\ndata: %s\n\n", errPayload)
					flusher.Flush()

					if database != nil && assistantResponse.Len() > 0 {
						_ = database.SaveChatMessage(&db.ChatMessage{
							ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
							SessionID: sessionID,
							Role:      "assistant",
							Content:   assistantResponse.String() + "\n\n[Analysis interrupted due to upstream network issue]",
							Status:    "FAILED",
							CreatedAt: time.Now().UTC().Format(time.RFC3339),
						})
					}
					return
				}

			case ev, ok := <-eventsChan:
				if !ok {
					// Complete
					fmt.Fprintf(w, "event: done\ndata: {\"session_id\": \"%s\"}\n\n", sessionID)
					flusher.Flush()

					// Save assistant response
					if database != nil && assistantResponse.Len() > 0 {
						_ = database.SaveChatMessage(&db.ChatMessage{
							ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
							SessionID: sessionID,
							Role:      "assistant",
							Content:   assistantResponse.String(),
							Status:    "COMPLETED",
							CreatedAt: time.Now().UTC().Format(time.RFC3339),
						})
						_ = database.TouchChatSession(sessionID, assistantResponse.String())
					}
					return
				}

				if ev.Event == ipc.EventAgentMessageChunk {
					assistantResponse.WriteString(ev.Chunk)
				}

				dataBytes, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, string(dataBytes))
				flusher.Flush()
			}
		}
	})

	// 5. Memory Graph JSON endpoint (registered on both /api/graph and /api/graph/data for web workspace compatibility)
	graphHandler := func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		ticker := r.URL.Query().Get("ticker")
		depthStr := r.URL.Query().Get("depth")
		minWeightStr := r.URL.Query().Get("min_weight")
		nodeTypesStr := r.URL.Query().Get("node_types")

		filter := db.MemoryGraphFilter{
			SessionID: sessionID,
			Ticker:    ticker,
		}
		if depthStr != "" {
			if d, err := strconv.Atoi(depthStr); err == nil {
				filter.Depth = d
			}
		}
		if minWeightStr != "" {
			if mw, err := strconv.ParseFloat(minWeightStr, 64); err == nil {
				filter.MinWeight = mw
			}
		}
		if nodeTypesStr != "" {
			parts := strings.Split(nodeTypesStr, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					filter.NodeTypes = append(filter.NodeTypes, trimmed)
				}
			}
		}

		nodes, edges, err := database.GetFilteredMemoryGraph(filter)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id":  sessionID,
			"total_nodes": len(nodes),
			"total_edges": len(edges),
			"nodes":       nodes,
			"edges":       edges,
		})
	}
	mux.HandleFunc("/api/graph", graphHandler)
	mux.HandleFunc("/api/graph/data", graphHandler)

	// Memory Graph Stats endpoint
	mux.HandleFunc("/api/graph/stats", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		ticker := r.URL.Query().Get("ticker")
		depthStr := r.URL.Query().Get("depth")
		nodeTypesStr := r.URL.Query().Get("node_types")

		filter := db.MemoryGraphFilter{
			SessionID: sessionID,
			Ticker:    ticker,
		}
		if depthStr != "" {
			if d, err := strconv.Atoi(depthStr); err == nil {
				filter.Depth = d
			}
		}
		if nodeTypesStr != "" {
			parts := strings.Split(nodeTypesStr, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					filter.NodeTypes = append(filter.NodeTypes, trimmed)
				}
			}
		}

		stats, err := database.GetMemoryGraphStats(filter)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(stats)
	})

	// 6. Interactive Memory Graph View endpoint (serves full Market Intelligence visualizer)
	mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		s.cfgMu.RLock()
		activeCfg := s.Config
		s.cfgMu.RUnlock()
		if activeCfg == nil {
			activeCfg = config.DefaultConfig()
		}

		configuredBin := activeCfg.Engine.PythonBin
		if envBin := os.Getenv("NISKAVA_PYTHON_BIN"); envBin != "" {
			configuredBin = envBin
		}
		pythonBin := ipc.ResolvePythonBin(configuredBin)

		dbPath := ""
		if s.DB != nil && s.DB.Path != "" {
			dbPath = s.DB.Path
		} else if activeCfg.Storage.DBPath != "" {
			dbPath = config.ExpandHome(activeCfg.Storage.DBPath)
		} else {
			homeDir, _ := os.UserHomeDir()
			dbPath = filepath.Join(homeDir, ".niskava", "niskava.db")
			if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
				dbPath = customDB
			}
		}

		sessionID := r.URL.Query().Get("session_id")
		ticker := r.URL.Query().Get("ticker")
		depth := r.URL.Query().Get("depth")
		nodeTypes := r.URL.Query().Get("node_types")
		isEmbed := r.URL.Query().Get("embed") == "true"

		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("niskava_graph_%d.html", time.Now().UnixNano()))

		args := []string{"-m", "engine.runner", "--db-path", dbPath, "--export-graph-html", tmpFile}
		if sessionID != "" {
			args = append(args, "--session", sessionID)
		}
		if ticker != "" {
			args = append(args, "--ticker", ticker)
		}
		if depth != "" {
			args = append(args, "--depth", depth)
		}
		if nodeTypes != "" {
			args = append(args, "--node-types", nodeTypes)
		}
		if isEmbed {
			args = append(args, "--embed")
		}

		wd, _ := os.Getwd()
		cmd := exec.CommandContext(r.Context(), pythonBin, args...)
		cmd.Dir = wd
		pythonPath := filepath.Join(wd, "backend") + string(filepath.ListSeparator) + wd
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		baseEnv := os.Environ()
		subEnv := s.buildSubprocessEnv()
		overrideKeys := make(map[string]bool)
		for k := range subEnv {
			overrideKeys[k] = true
		}
		overrideKeys["PYTHONPATH"] = true
		overrideKeys["PYTHONIOENCODING"] = true
		overrideKeys["PYTHONUTF8"] = true

		cleanEnv := make([]string, 0, len(baseEnv)+len(overrideKeys))
		for _, envVar := range baseEnv {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 && overrideKeys[parts[0]] {
				continue
			}
			cleanEnv = append(cleanEnv, envVar)
		}

		cleanEnv = append(cleanEnv,
			"PYTHONPATH="+pythonPath,
			"PYTHONIOENCODING=utf-8",
			"PYTHONUTF8=1",
		)
		for k, v := range subEnv {
			if strings.TrimSpace(k) != "" && v != "" {
				cleanEnv = append(cleanEnv, fmt.Sprintf("%s=%s", k, v))
			}
		}
		cmd.Env = cleanEnv
		if out, err := cmd.CombinedOutput(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate graph visualization: %v\nOutput: %s", err, string(out)), http.StatusInternalServerError)
			return
		}

		defer os.Remove(tmpFile)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, tmpFile)
	})

	// 7. Interactive AI Assistant Web Workspace
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, RenderWorkspaceHTML(s.Port))
	})

	// Find free port if requested port is taken
	addr := fmt.Sprintf("127.0.0.1:%d", requestedPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, fmt.Errorf("failed to bind listener: %w", err)
		}
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	s.Port = actualPort
	s.URL = fmt.Sprintf("http://localhost:%d", actualPort)

	s.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second, // Protect against Slowloris attacks
		ReadTimeout:       0,                // 0 disables connection-wide read deadline for SSE streaming
		WriteTimeout:      0,                // 0 disables global write deadline, essential for long-running SSE streaming sessions
	}

	go func() {
		_ = s.httpServer.Serve(listener)
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)

		s.botMu.Lock()
		if s.BotService != nil && s.BotService.IsStarted() {
			s.BotService.Stop()
		}
		s.botMu.Unlock()
	}()

	return s, nil
}

// OpenBrowser opens the specified URL in the user's default web browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform for auto-open browser: %s", runtime.GOOS)
	}
	return cmd.Start()
}
