// Package server provides the background REST/SSE daemon server for Niskava Agent.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
)

// Server encapsulates the background HTTP server instance.
type Server struct {
	httpServer *http.Server
	Port       int
	DB         *db.DB
	URL        string
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

	// 3. Root / Web UI placeholder
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
    <title>Niskava Agent — Market Intelligence Dashboard</title>
    <style>
        body { background: #0B0F19; color: #F1F5F9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 40px 20px; }
        .container { max-width: 900px; margin: 0 auto; background: #111827; border: 1px solid #1E293B; border-radius: 12px; padding: 32px; }
        .badge { background: #0284C7; color: #fff; padding: 4px 10px; border-radius: 6px; font-weight: bold; font-size: 13px; }
        h1 { color: #38BDF8; margin-top: 12px; font-size: 24px; }
        p { color: #94A3B8; line-height: 1.6; }
        .card { background: #1E293B; border-radius: 8px; padding: 16px; margin-top: 20px; border-left: 4px solid #10B981; }
        .disclaimer { margin-top: 30px; padding: 16px; background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); border-radius: 8px; color: #FCA5A5; font-size: 13px; }
        code { background: #0F172A; padding: 2px 6px; border-radius: 4px; color: #38BDF8; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        <span class="badge">LOCAL DAEMON ACTIVE</span>
        <h1>Niskava Agent — IDX Market Intelligence</h1>
        <p>Server lokal berjalan di port <code>http://localhost:%d</code>. Endpoint API dan Server-Sent Events (SSE) aktif untuk visualisasi anomali transaksi dan keterbukaan informasi bursa.</p>
        
        <div class="card">
            <h3 style="margin: 0 0 8px 0; color: #F8FAFC;">Status Sistem & Persistensi</h3>
            <p style="margin: 0;">Database: <code>~/.niskava/niskava.db</code> (SQLite WAL Mode Active)<br>API Endpoint: <a href="/api/sessions" style="color: #38BDF8;">/api/sessions</a> | <a href="/api/health" style="color: #38BDF8;">/api/health</a></p>
        </div>

        <div class="disclaimer">
            <strong>DISCLAIMER FINANSIAL (NON-ADVISORY - LAW 2 & ATURAN 12):</strong><br>
            Niskava Agent adalah platform intelijen pasar dan OSINT otonom, BUKAN penasihat investasi berlisensi. Platform TIDAK PERNAH memberikan rekomendasi BELI/JUAL atau target harga saham apa pun.
        </div>
    </div>
</body>
</html>`, s.Port)
	})

	// Find free port if requested port is taken
	addr := fmt.Sprintf("127.0.0.1:%d", requestedPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Fallback to random available port
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
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Run in background goroutine
	go func() {
		_ = s.httpServer.Serve(listener)
	}()

	// Graceful shutdown on context cancellation
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
