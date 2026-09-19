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
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/ipc"
)

// Server encapsulates the background HTTP server instance.
type Server struct {
	httpServer *http.Server
	Port       int
	DB         *db.DB
	URL        string
}

// ChatRequest represents the JSON payload for /api/chat.
type ChatRequest struct {
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id,omitempty"`
}

// Start launches the background HTTP server on the specified port (or auto-finds free port).
func Start(ctx context.Context, requestedPort int, database *db.DB) (*Server, error) {
	mux := http.NewServeMux()

	s := &Server{
		Port: requestedPort,
		DB:   database,
	}

	// 1. Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"app":       "Niskava Agent",
			"version":   "1.0.0",
			"market":    "IDX",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 2. Sessions list endpoint
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessions, err := database.ListInvestigations(50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		chatSessions, _ := database.ListChatSessions(50)

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"total":          len(sessions) + len(chatSessions),
			"data":           sessions,
			"investigations": sessions,
			"chat_sessions":  chatSessions,
		})
	})

	// 3. Chat History endpoint
	mux.HandleFunc("/api/chat/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id": sessionID,
			"total":      len(history),
			"messages":   history,
		})
	})

	// 3b. Reset / Delete Chat History endpoint
	mux.HandleFunc("/api/chat/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost && r.Method != http.MethodDelete {
			http.Error(w, `{"error": "Method not allowed. Use POST or DELETE."}`, http.StatusMethodNotAllowed)
			return
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}

		sessionID := r.URL.Query().Get("session_id")
		if sessionID != "" {
			_ = database.DeleteChatSession(sessionID)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"deleted_session": sessionID,
			})
			return
		}

		_ = database.ClearAllChatHistory()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"cleared_all": true,
		})
	})

	// 4. Conversational Chat SSE Streaming endpoint
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
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

		// Save User Message
		if database != nil {
			_ = database.SaveChatMessage(&db.ChatMessage{
				ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
				SessionID: sessionID,
				Role:      "user",
				Content:   req.Prompt,
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

		pythonBin := "python3"
		if runtime.GOOS == "windows" {
			pythonBin = "python"
			localVenv := filepath.Join(".venv", "Scripts", "python.exe")
			if _, err := os.Stat(localVenv); err == nil {
				pythonBin = localVenv
			}
		} else {
			localVenv := filepath.Join(".venv", "bin", "python3")
			if _, err := os.Stat(localVenv); err == nil {
				pythonBin = localVenv
			}
		}

		dbPath := filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
		if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
			dbPath = customDB
		}

		wd, _ := os.Getwd()

		runnerParams := ipc.RunnerParams{
			PythonBin: pythonBin,
			WorkDir:   wd,
			DBPath:    dbPath,
			Prompt:    req.Prompt,
			SessionID: sessionID,
			Offline:   os.Getenv("NISKAVA_OFFLINE") == "1",
		}

		eventsChan, errChan := ipc.RunSubprocess(r.Context(), runnerParams)

		var assistantResponse strings.Builder

		for {
			select {
			case <-r.Context().Done():
				return

			case err, ok := <-errChan:
				if !ok {
					errChan = nil
					continue
				}
				if err != nil {
					errJSON, _ := json.Marshal(map[string]string{"error": err.Error(), "event": "error"})
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(errJSON))
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
							CreatedAt: time.Now().UTC().Format(time.RFC3339),
						})
					}
					return
				}

				if ev.Event == ipc.EventAgentMessageChunk {
					assistantResponse.WriteString(ev.Chunk)
				}
				if ev.Event == ipc.EventAgentMessageComplete && assistantResponse.Len() == 0 {
					assistantResponse.WriteString(ev.Content)
				}

				dataBytes, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, string(dataBytes))
				flusher.Flush()
			}
		}
	})

	// 4b. Export Investigation Session endpoint
	mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "markdown"
		}
		if database == nil {
			http.Error(w, `{"error": "database not initialized"}`, http.StatusInternalServerError)
			return
		}
		history, err := database.GetChatHistory(sessionID, 100)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		if format == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"niskava-report-%s.json\"", sessionID))
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"session_id":  sessionID,
				"exported_at": time.Now().UTC().Format(time.RFC3339),
				"messages":    history,
			})
			return
		}
		// Markdown export
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"niskava-report-%s.md\"", sessionID))
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Laporan Riset Niskava Agent (Sesi: %s)\n\n", sessionID))
		sb.WriteString(fmt.Sprintf("**Waktu Ekspor:** %s\n\n", time.Now().Format("2006-01-02 15:04:05 MST")))
		sb.WriteString("---\n\n")
		for _, msg := range history {
			roleLabel := "Pengguna"
			if msg.Role == "assistant" {
				roleLabel = "Niskava Agent"
			}
			sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", roleLabel, msg.Content))
		}
		sb.WriteString("\n---\n*Pemberitahuan Kepatuhan Hukum: Laporan ini dihasilkan secara otomatis oleh Niskava Agent untuk tujuan riset informasi, edukasi, dan intelijen data terbuka (OSINT). Niskava bukan merupakan penasihat investasi berlisensi.*\n")
		_, _ = w.Write([]byte(sb.String()))
	})

	// 5. Interactive AI Assistant Web Workspace
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(RenderWorkspaceHTML(s.Port)))
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
		WriteTimeout: 60 * time.Second,
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
