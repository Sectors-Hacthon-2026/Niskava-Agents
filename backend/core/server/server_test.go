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

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
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

func TestChatSession_Export_And_Search(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_export.db")

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

	// 1. Setup Session and Messages
	sessionID := "TEST-EXP-001"
	_ = database.CreateChatSession(&db.ChatSession{
		ID:     sessionID,
		Title:  "Riset Komoditas ANTM",
		Model:  "hermes",
		Status: "IDLE",
	})

	thought := "Menghitung volume MA20 dan Z-Score saham ANTM..."
	_ = database.SaveChatMessage(&db.ChatMessage{
		ID:        "MSG-EXP-01",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Analisis volume lonjakan ANTM hari ini",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	_ = database.SaveChatMessage(&db.ChatMessage{
		ID:        "MSG-EXP-02",
		SessionID: sessionID,
		Role:      "assistant",
		Thought:   &thought,
		Content:   "Ditemukan anomali lonjakan volume 3.12σ pada perdagangan terakhir.",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})

	// 2. Test Markdown Export
	resp, err := client.Get(srv.URL + "/api/chat/sessions/" + sessionID + "/export?format=markdown")
	if err != nil {
		t.Fatalf("GET export markdown failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var mdBuf bytes.Buffer
	_, _ = mdBuf.ReadFrom(resp.Body)
	resp.Body.Close()
	mdStr := mdBuf.String()

	if !strings.Contains(mdStr, "# Laporan Riset Pasar: Riset Komoditas ANTM") {
		t.Errorf("expected title in export markdown, got: %s", mdStr[:100])
	}
	if !strings.Contains(mdStr, "DISCLAIMER") {
		t.Error("expected non-advisory disclaimer in export markdown")
	}
	if !strings.Contains(mdStr, "Proses Berpikir Analitis") || !strings.Contains(mdStr, thought) {
		t.Error("expected chain-of-thought in export markdown")
	}
	if !strings.Contains(mdStr, "Ditemukan anomali lonjakan volume 3.12σ") {
		t.Error("expected assistant message content in export markdown")
	}

	// 3. Test JSON Export
	resp, err = client.Get(srv.URL + "/api/chat/sessions/" + sessionID + "/export?format=json")
	if err != nil {
		t.Fatalf("GET export json failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for json export, got %d", resp.StatusCode)
	}
	var jsonExport map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&jsonExport)
	resp.Body.Close()

	if jsonExport["session"] == nil || jsonExport["messages"] == nil {
		t.Errorf("expected session and messages in json export, got: %+v", jsonExport)
	}
	messagesArr := jsonExport["messages"].([]interface{})
	if len(messagesArr) != 2 {
		t.Errorf("expected 2 messages in json export, got %d", len(messagesArr))
	}

	// 4. Test Global Message Search
	resp, err = client.Get(srv.URL + "/api/chat/search?q=lonjakan")
	if err != nil {
		t.Fatalf("GET /api/chat/search failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on search, got %d", resp.StatusCode)
	}
	var searchRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&searchRes)
	resp.Body.Close()

	if int(searchRes["total"].(float64)) != 2 {
		t.Errorf("expected 2 search results for 'lonjakan', got %v", searchRes["total"])
	}

	// Search non-existent
	resp, err = client.Get(srv.URL + "/api/chat/search?q=BukanSaham123")
	if err != nil {
		t.Fatalf("GET /api/chat/search non-existent failed: %v", err)
	}
	var emptySearch map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&emptySearch)
	resp.Body.Close()

	if int(emptySearch["total"].(float64)) != 0 {
		t.Errorf("expected 0 search results for nonexistent, got %v", emptySearch["total"])
	}
}

