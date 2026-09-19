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

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"total": len(sessions),
			"data":  sessions,
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
		localVenv := filepath.Join(".venv", "bin", "python3")
		if _, err := os.Stat(localVenv); err == nil {
			pythonBin = localVenv
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
				if ok && err != nil {
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
					flusher.Flush()
				}
				return

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

				dataBytes, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, string(dataBytes))
				flusher.Flush()
			}
		}
	})

	// 5. Memory Graph JSON endpoint
	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// 6. Interactive Memory Graph View endpoint (serves full Cyber-OSINT visualizer)
	mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		pythonBin := "python3"
		localVenv := filepath.Join(".venv", "bin", "python3")
		if _, err := os.Stat(localVenv); err == nil {
			pythonBin = localVenv
		}

		dbPath := filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
		if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
			dbPath = customDB
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
		cmd.Env = append(os.Environ(), "PYTHONPATH="+wd)
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
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Niskava Agent — AI Financial Research Assistant (IDX)</title>
    <style>
        :root {
            --bg: #090D16;
            --surface: #111827;
            --surface-card: #1F2937;
            --border: #374151;
            --accent: #00E5FF;
            --accent-glow: rgba(0, 229, 255, 0.15);
            --text-main: #F9FAFB;
            --text-muted: #9CA3AF;
            --success: #10B981;
            --warning: #F59E0B;
            --danger: #EF4444;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: var(--bg); color: var(--text-main); font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; height: 100vh; overflow: hidden; }
        
        /* Sidebar */
        .sidebar { width: 300px; background: var(--surface); border-right: 1px solid var(--border); display: flex; flex-direction: column; padding: 20px; }
        .brand { font-size: 18px; font-weight: 800; color: var(--accent); letter-spacing: 0.5px; display: flex; align-items: center; gap: 8px; margin-bottom: 24px; }
        .badge-live { background: rgba(16, 185, 129, 0.2); color: var(--success); font-size: 11px; padding: 3px 8px; border-radius: 99px; border: 1px solid var(--success); }
        .section-title { font-size: 12px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; margin-bottom: 12px; letter-spacing: 0.5px; }
        .quick-prompts { display: flex; flex-direction: column; gap: 8px; margin-bottom: 24px; }
        .prompt-chip { background: var(--surface-card); border: 1px solid var(--border); padding: 10px 12px; border-radius: 8px; font-size: 13px; color: var(--text-main); cursor: pointer; text-align: left; transition: all 0.2s; }
        .prompt-chip:hover { border-color: var(--accent); background: var(--accent-glow); }
        
        /* Main Chat Area */
        .main-content { flex: 1; display: flex; flex-direction: column; background: var(--bg); }
        .header { height: 64px; border-bottom: 1px solid var(--border); background: var(--surface); display: flex; align-items: center; justify-content: space-between; padding: 0 28px; }
        .header-title { font-size: 15px; font-weight: 600; }
        .chat-container { flex: 1; overflow-y: auto; padding: 28px; display: flex; flex-direction: column; gap: 20px; }
        
        /* Messages */
        .msg { display: flex; flex-direction: column; max-width: 85%%; }
        .msg-user { align-self: flex-end; }
        .msg-user .bubble { background: #0284C7; color: #fff; border-radius: 14px 14px 2px 14px; padding: 12px 18px; font-size: 14px; line-height: 1.5; }
        .msg-agent { align-self: flex-start; }
        .msg-agent .bubble { background: var(--surface); border: 1px solid var(--border); border-radius: 14px 14px 14px 2px; padding: 18px; font-size: 14px; line-height: 1.6; }
        
        /* Thoughts & Tools Stream */
        .thought-box { background: rgba(15, 23, 42, 0.6); border-left: 3px solid var(--accent); padding: 8px 12px; font-size: 12px; color: #94A3B8; font-style: italic; margin-bottom: 10px; border-radius: 0 6px 6px 0; }
        .tool-box { background: rgba(245, 158, 11, 0.1); border: 1px dashed var(--warning); padding: 6px 10px; font-size: 12px; color: #FCD34D; font-family: monospace; border-radius: 6px; margin-bottom: 10px; }
        
        /* Badges */
        .tag-supported { background: var(--success); color: #000; font-weight: bold; padding: 2px 6px; border-radius: 4px; font-size: 11px; }
        .tag-uncertain { background: var(--warning); color: #000; font-weight: bold; padding: 2px 6px; border-radius: 4px; font-size: 11px; }
        
        /* Input Box */
        .input-area { padding: 20px 28px; background: var(--surface); border-top: 1px solid var(--border); display: flex; gap: 12px; }
        .input-box { flex: 1; background: var(--surface-card); border: 1px solid var(--border); border-radius: 10px; padding: 14px 18px; color: #fff; font-size: 14px; outline: none; }
        .input-box:focus { border-color: var(--accent); box-shadow: 0 0 0 2px var(--accent-glow); }
        .btn-send { background: var(--accent); color: #090D16; border: none; border-radius: 10px; padding: 0 24px; font-weight: 700; cursor: pointer; transition: 0.2s; }
        .btn-send:hover { opacity: 0.9; }

        .disclaimer { font-size: 11px; color: #EF4444; margin-top: 14px; border-top: 1px solid rgba(239, 68, 68, 0.2); padding-top: 8px; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="brand">
            ⚡ NISKAVA AGENT
            <span class="badge-live">ONLINE</span>
        </div>

        <div class="section-title">Contoh Riset Pasar (Quick Prompts)</div>
        <div class="quick-prompts">
            <button class="prompt-chip" onclick="sendPrompt(this.innerText)">Analisis lonjakan volume ANTM 30 hari terakhir</button>
            <button class="prompt-chip" onclick="sendPrompt(this.innerText)">Apakah ada anomali transaksi asing di BBCA minggu ini?</button>
            <button class="prompt-chip" onclick="sendPrompt(this.innerText)">Cari keterbukaan informasi dan katalis saham BUMI</button>
            <button class="prompt-chip" onclick="sendPrompt(this.innerText)">Bandingkan pergerakan saham nikel INCO dan ANTM</button>
        </div>

        <div class="section-title">Visualisasi Graf Memori</div>
        <div style="margin-bottom: 20px;">
            <a href="/graph" target="_blank" style="text-decoration:none;">
                <button class="prompt-chip" style="width:100%%; border-color:var(--accent); color:var(--accent); font-weight:700; background:rgba(0, 229, 255, 0.08);">
                    🕸️ Buka Knowledge Graph
                </button>
            </a>
        </div>

        <div class="section-title" style="margin-top: auto;">Sistem & Persistensi</div>
        <p style="font-size: 12px; color: var(--text-muted); line-height: 1.5;">
            • Model: <code>9router / hermes</code><br>
            • Database: <code>SQLite WAL Active</code><br>
            • Hukum 1 & 2 Kepatuhan Penuh
        </p>
    </div>

    <div class="main-content">
        <div class="header">
            <div class="header-title">Autonomous Financial OSINT Research Assistant (Bursa Efek Indonesia)</div>
            <div style="font-size: 13px; color: var(--text-muted);">Port: <code>%d</code></div>
        </div>

        <div class="chat-container" id="chatArea">
            <div class="msg msg-agent">
                <div class="bubble">
                    <strong>Halo! Saya Niskava Agent.</strong><br>
                    Asisten riset intelijen pasar dan pembuktian anomali saham di Bursa Efek Indonesia (IDX). Tanyakan apa saja mengenai emiten, lonjakan transaksi kuantitatif (Z-score), atau keterbukaan informasi resmi.
                </div>
            </div>
        </div>

        <div class="input-area">
            <input type="text" id="promptInput" class="input-box" placeholder="Ketik pertanyaan riset pasar saham Anda di sini..." onkeypress="handleKey(event)" />
            <button class="btn-send" onclick="submitCurrentPrompt()">Kirim ➔</button>
        </div>
    </div>

    <script>
        const chatArea = document.getElementById('chatArea');
        const promptInput = document.getElementById('promptInput');
        const sessionID = 'WEB-' + Date.now();

        function handleKey(e) {
            if (e.key === 'Enter') submitCurrentPrompt();
        }

        function sendPrompt(text) {
            promptInput.value = text;
            submitCurrentPrompt();
        }

        async function submitCurrentPrompt() {
            const prompt = promptInput.value.trim();
            if (!prompt) return;

            // Add user bubble
            const userMsg = document.createElement('div');
            userMsg.className = 'msg msg-user';
            userMsg.innerHTML = '<div class="bubble">' + escapeHtml(prompt) + '</div>';
            chatArea.appendChild(userMsg);
            promptInput.value = '';
            chatArea.scrollTop = chatArea.scrollHeight;

            // Add agent placeholder
            const agentMsg = document.createElement('div');
            agentMsg.className = 'msg msg-agent';
            const agentBubble = document.createElement('div');
            agentBubble.className = 'bubble';
            agentBubble.innerHTML = '<div class="thought-box">💭 Menghubungkan ke ReAct Agent Engine...</div>';
            agentMsg.appendChild(agentBubble);
            chatArea.appendChild(agentMsg);
            chatArea.scrollTop = chatArea.scrollHeight;

            try {
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ prompt: prompt, session_id: sessionID })
                });

                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                let accumulatedText = '';

                while (true) {
                    const { value, done } = await reader.read();
                    if (done) break;

                    const chunk = decoder.decode(value);
                    const lines = chunk.split('\n');

                    for (let line of lines) {
                        if (line.startsWith('data: ')) {
                            const dataStr = line.replace('data: ', '').trim();
                            if (dataStr.startsWith('{')) {
                                try {
                                    const ev = JSON.parse(dataStr);
                                    if (ev.event === 'agent_thought') {
                                        agentBubble.innerHTML = '<div class="thought-box">💭 ' + escapeHtml(ev.thought) + '</div>' + accumulatedText;
                                    } else if (ev.event === 'agent_tool_call') {
                                        agentBubble.innerHTML += '<div class="tool-box">⚡ [Action Tool] ' + escapeHtml(ev.tool) + '</div>';
                                    } else if (ev.event === 'agent_message_chunk') {
                                        accumulatedText += ev.chunk;
                                        agentBubble.innerHTML = accumulatedText.replace(/\n/g, '<br>');
                                    } else if (ev.event === 'finding_emitted') {
                                        accumulatedText += '<div style="margin-top:8px; padding:8px; background:rgba(16,185,129,0.1); border-left:3px solid #10B981;">' +
                                            '<span class="tag-supported">SUPPORTED</span> <strong>' + escapeHtml(ev.title) + '</strong><br>' +
                                            '<small>' + escapeHtml(ev.claim_text) + '</small></div>';
                                        agentBubble.innerHTML = accumulatedText;
                                    }
                                    chatArea.scrollTop = chatArea.scrollHeight;
                                } catch (e) {}
                            }
                        }
                    }
                }
            } catch (err) {
                agentBubble.innerHTML += '<div style="color:#EF4444; margin-top:8px;">Terjadi kendala koneksi ke server daemon.</div>';
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.innerText = text || '';
            return div.innerHTML;
        }
    </script>
</body>
</html>`, s.Port)
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
