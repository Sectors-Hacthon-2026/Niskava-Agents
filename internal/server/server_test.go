package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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

func TestMemoryGraphEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_mem.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	srv, err := Start(ctx, 0, database)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(srv.URL + "/api/graph")
	if err != nil {
		t.Fatalf("failed to GET /api/graph: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if _, ok := data["nodes"]; !ok {
		t.Errorf("expected 'nodes' in graph response, got %+v", data)
	}
	if _, ok := data["edges"]; !ok {
		t.Errorf("expected 'edges' in graph response, got %+v", data)
	}
}

func TestChatSessions_REST_Endpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_chat.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	srv, err := Start(ctx, 0, database)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	// 1. GET /api/chat/sessions (initially empty)
	resp, err := client.Get(srv.URL + "/api/chat/sessions")
	if err != nil {
		t.Fatalf("GET /api/chat/sessions failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var listRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&listRes)
	resp.Body.Close()
	if int(listRes["total"].(float64)) != 0 {
		t.Fatalf("expected 0 sessions initially, got %v", listRes["total"])
	}

	// 2. POST /api/chat/sessions (create session)
	createPayload := []byte(`{"id":"TEST-WEB-001","title":"Analisis ANTM","model":"hermes"}`)
	resp, err = client.Post(srv.URL+"/api/chat/sessions", "application/json", bytes.NewReader(createPayload))
	if err != nil {
		t.Fatalf("POST /api/chat/sessions failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. GET /api/chat/sessions/TEST-WEB-001
	resp, err = client.Get(srv.URL + "/api/chat/sessions/TEST-WEB-001")
	if err != nil {
		t.Fatalf("GET session failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var getRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&getRes)
	resp.Body.Close()
	sessData := getRes["session"].(map[string]interface{})
	if sessData["title"] != "Analisis ANTM" {
		t.Errorf("expected title 'Analisis ANTM', got %v", sessData["title"])
	}

	// 4. PATCH /api/chat/sessions/TEST-WEB-001 (rename & pin)
	patchReq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/chat/sessions/TEST-WEB-001", strings.NewReader(`{"title":"Analisis Nikel ANTM","is_pinned":true}`))
	patchReq.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(patchReq)
	if err != nil {
		t.Fatalf("PATCH failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on PATCH, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. POST /api/chat/sessions/TEST-WEB-001/fork
	forkReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/chat/sessions/TEST-WEB-001/fork", strings.NewReader(`{"new_id":"TEST-WEB-FORK-001","title":"Cabang ANTM"}`))
	forkReq.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(forkReq)
	if err != nil {
		t.Fatalf("POST fork failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created on fork, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify fork created
	resp, _ = client.Get(srv.URL + "/api/chat/sessions/TEST-WEB-FORK-001")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected forked session to exist, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 6. POST /api/chat/sessions/TEST-WEB-001/abort
	abortReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/chat/sessions/TEST-WEB-001/abort", nil)
	resp, err = client.Do(abortReq)
	if err != nil {
		t.Fatalf("POST abort failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on abort, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 7. POST /api/chat/sessions/TEST-WEB-FORK-001/reset
	resetReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/chat/sessions/TEST-WEB-FORK-001/reset", nil)
	resp, err = client.Do(resetReq)
	if err != nil {
		t.Fatalf("POST reset failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on reset, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 8. DELETE /api/chat/sessions/TEST-WEB-001
	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/chat/sessions/TEST-WEB-001", nil)
	resp, err = client.Do(delReq)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on DELETE, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify session is gone
	resp, _ = client.Get(srv.URL + "/api/chat/sessions/TEST-WEB-001")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found after delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

