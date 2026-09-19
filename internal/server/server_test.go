package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
)

func TestServerStartAndHealth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := Start(ctx, 0, nil) // 0 binds to a free port
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("failed to GET health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}
}

func TestServerWorkspaceHTML(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := Start(ctx, 0, nil)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("failed to GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	bodyStr := string(body)

	// Check core elements
	checks := []string{
		"<!DOCTYPE html>",
		"data-theme=\"dark\"",
		"[data-theme=\"light\"]",
		"prefers-color-scheme: dark",
		"NISKAVA",
		"BEI LIVE",
		"Percakapan AI",
		"Metrik & Candlestick",
		"Timeline Kejadian",
		"themeToggleBtn",
		"Pemberitahuan Kepatuhan",
	}

	for _, check := range checks {
		if !strings.Contains(bodyStr, check) {
			t.Errorf("expected HTML to contain %q", check)
		}
	}
}

func TestServerExportEndpoint(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "niskava_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	sessionID := "TEST-SESSION-001"
	_ = database.SaveChatMessage(&db.ChatMessage{
		ID:        "MSG-1",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Analisis lonjakan volume ANTM",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	_ = database.SaveChatMessage(&db.ChatMessage{
		ID:        "MSG-2",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Terdeteksi anomali volume Z-score +3.84σ pada 12 Sep 2026.",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := Start(ctx, 0, database)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Test Markdown export
	respMD, err := http.Get(srv.URL + "/api/export?session_id=" + sessionID + "&format=markdown")
	if err != nil {
		t.Fatalf("failed to GET markdown export: %v", err)
	}
	defer respMD.Body.Close()

	if respMD.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for markdown export, got %d", respMD.StatusCode)
	}

	bodyMD, _ := io.ReadAll(respMD.Body)
	if !strings.Contains(string(bodyMD), "TEST-SESSION-001") {
		t.Errorf("expected markdown export to contain session ID")
	}

	// Test JSON export
	respJSON, err := http.Get(srv.URL + "/api/export?session_id=" + sessionID + "&format=json")
	if err != nil {
		t.Fatalf("failed to GET json export: %v", err)
	}
	defer respJSON.Body.Close()

	if respJSON.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for json export, got %d", respJSON.StatusCode)
	}

	var exportData map[string]interface{}
	if err := json.NewDecoder(respJSON.Body).Decode(&exportData); err != nil {
		t.Fatalf("failed to decode json export: %v", err)
	}

	if exportData["session_id"] != sessionID {
		t.Errorf("expected session_id %s, got %v", sessionID, exportData["session_id"])
	}

	// Test /api/sessions
	respSessions, err := http.Get(srv.URL + "/api/sessions")
	if err != nil {
		t.Fatalf("failed to GET /api/sessions: %v", err)
	}
	defer respSessions.Body.Close()

	if respSessions.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/sessions, got %d", respSessions.StatusCode)
	}

	var sessionsData map[string]interface{}
	if err := json.NewDecoder(respSessions.Body).Decode(&sessionsData); err != nil {
		t.Fatalf("failed to decode /api/sessions JSON: %v", err)
	}

	if sessionsData["total"].(float64) < 1 {
		t.Errorf("expected at least 1 session, got %v", sessionsData["total"])
	}

	// Test /api/chat/history
	respHist, err := http.Get(srv.URL + "/api/chat/history?session_id=" + sessionID)
	if err != nil {
		t.Fatalf("failed to GET /api/chat/history: %v", err)
	}
	defer respHist.Body.Close()

	if respHist.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/chat/history, got %d", respHist.StatusCode)
	}

	var histData map[string]interface{}
	if err := json.NewDecoder(respHist.Body).Decode(&histData); err != nil {
		t.Fatalf("failed to decode /api/chat/history JSON: %v", err)
	}

	if histData["total"].(float64) != 2 {
		t.Errorf("expected 2 history messages, got %v", histData["total"])
	}
}
