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

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
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

// Start launches the background HTTP server on the specified port (or auto-finds free port).
func Start(ctx context.Context, requestedPort int, database *db.DB) (*Server, error) {
	mux := http.NewServeMux()

	s := &Server{
		Port:           requestedPort,
		DB:             database,
		SessionManager: NewSessionManager(),
	}

	// Helper for CORS preflight and headers
	enableCORS := func(w http.ResponseWriter, r *http.Request) bool {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
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
			md.WriteString("> Niskava Agent adalah platform intelijen dan OSINT pasar modal otonom untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi atau broker berizin. Seluruh temuan, skor anomali, dan korelasi bukti disajikan secara deskriptif untuk tujuan riset dan verifikasi fakta, serta BUKAN merupakan rekomendasi beli/jual atau target harga investasi.\n\n")

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

		// Disable write timeout for this streaming SSE connection
		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

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

		pythonBin := ipc.FindPythonBinary()

		dbPath := ""
		if s.DB != nil && s.DB.Path != "" {
			dbPath = s.DB.Path
		} else {
			dbPath = filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
			if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
				dbPath = customDB
			}
		}

		wd, _ := os.Getwd()

		chatLang := os.Getenv("NISKAVA_LANG")
		if chatLang == "" {
			chatLang = "id"
		}

		runnerParams := ipc.RunnerParams{
			PythonBin: pythonBin,
			WorkDir:   wd,
			DBPath:    dbPath,
			Prompt:    req.Prompt,
			SessionID: sessionID,
			Offline:   os.Getenv("NISKAVA_OFFLINE") == "1",
			Language:  chatLang,
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
		w.Header().Set("Content-Type", "application/json")
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		nodes, edges, err := database.GetMemoryGraph(sessionID)
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

	// 6. Interactive Memory Graph View endpoint (serves full Cyber-OSINT visualizer)
	mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		pythonBin := ipc.FindPythonBinary()

		dbPath := ""
		if s.DB != nil && s.DB.Path != "" {
			dbPath = s.DB.Path
		} else {
			dbPath = filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
			if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
				dbPath = customDB
			}
		}

		sessionID := r.URL.Query().Get("session_id")
		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("niskava_graph_%d.html", time.Now().UnixNano()))

		args := []string{"-m", "engine.runner", "--db-path", dbPath, "--export-graph-html", tmpFile}
		if sessionID != "" {
			args = append(args, "--session", sessionID)
		}

		wd, _ := os.Getwd()
		cmd := exec.CommandContext(r.Context(), pythonBin, args...)
		cmd.Dir = wd
		pythonPath := filepath.Join(wd, "backend") + string(filepath.ListSeparator) + wd
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		cmd.Env = append(os.Environ(), "PYTHONPATH="+pythonPath)
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
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 0, // 0 disables global write deadline, essential for long-running SSE streaming sessions
	}

	go func() {
		_ = s.httpServer.Serve(listener)
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
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
