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

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
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

func TestGraphStatsEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_graph_stats.db")

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

	resp, err := http.Get(srv.URL + "/api/graph/stats")
	if err != nil {
		t.Fatalf("failed to get graph stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}
	if _, ok := stats["total_nodes"]; !ok {
		t.Errorf("expected 'total_nodes' in stats")
	}
}

func TestSettingsEndpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	cfg := config.DefaultConfig()
	cfg.Auth.AIProvider = "gemini"
	cfg.Auth.GeminiModel = "gemini-2.0-flash"
	cfg.Auth.SectorsAPIKey = "sectors_secret_sample"
	_ = config.SaveConfig(cfg, cfgPath)

	srv, err := Start(ctx, 0, nil, cfg)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// 1. GET /api/settings
	resp, err := http.Get(srv.URL + "/api/settings")
	if err != nil {
		t.Fatalf("GET /api/settings failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var view map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&view); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	auth := view["auth"].(map[string]interface{})
	if auth["sectors_api_key"] == "sectors_secret_sample" {
		t.Fatalf("expected secret to be masked")
	}

	// 2. PATCH /api/settings
	patchBody := `{"preferences":{"language":"id"},"auth":{"ai_provider":"ollama","ollama_model":"qwen2.5"}}`
	patchReq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/settings", strings.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatalf("PATCH /api/settings failed: %v", err)
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", patchResp.StatusCode)
	}

	// Verify server config was hot-reloaded
	if srv.Config.Preferences.Language != "id" {
		t.Errorf("expected hot-reloaded language 'id', got '%s'", srv.Config.Preferences.Language)
	}
	if srv.Config.Auth.AIProvider != "ollama" {
		t.Errorf("expected hot-reloaded ai_provider 'ollama', got '%s'", srv.Config.Auth.AIProvider)
	}
}

func TestTestConnectionEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.Preferences.OfflineMode = true

	srv, err := Start(ctx, 0, nil, cfg)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// 1. Unknown target
	resp, err := http.Post(srv.URL+"/api/settings/test-connection", "application/json", strings.NewReader(`{"target":"invalid_target"}`))
	if err != nil {
		t.Fatalf("POST test-connection failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown target, got %d", resp.StatusCode)
	}

	// 2. Sectors target in offline mode
	resp2, err := http.Post(srv.URL+"/api/settings/test-connection", "application/json", strings.NewReader(`{"target":"sectors"}`))
	if err != nil {
		t.Fatalf("POST test-connection failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for sectors test, got %d", resp2.StatusCode)
	}

	var data map[string]interface{}
	_ = json.NewDecoder(resp2.Body).Decode(&data)
	if data["success"] != true {
		t.Errorf("expected success true in offline mode, got %+v", data)
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

func TestGraphDataEndpointAlias(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_graph.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	srv, err := Start(ctx, 0, database)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(srv.URL + "/api/graph/data")
	if err != nil {
		t.Fatalf("failed to call /api/graph/data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK from /api/graph/data alias, got %d", resp.StatusCode)
	}
}

func TestInvestigationEndpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_inv.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// 1. Seed test investigation, anomaly, and finding
	invID := "INV-TEST-REST-01"
	err = database.CreateInvestigation(&db.Investigation{
		ID:            invID,
		Ticker:        "ANTM",
		Market:        "IDX",
		TimeframeDays: 30,
		Status:        "COMPLETED",
		StartedAt:     "2026-09-20T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("failed to seed investigation: %v", err)
	}

	err = database.CreateAnomaly(&db.Anomaly{
		ID:              "ANOM-01",
		InvestigationID: invID,
		AnomalyDate:     "2026-09-18",
		MetricType:      "volume_z_score",
		MetricValue:     15000000,
		BaselineValue:   5000000,
		ZScore:          3.45,
		Description:     "Volume spike detected",
	})
	if err != nil {
		t.Fatalf("failed to seed anomaly: %v", err)
	}

	err = database.CreateFinding(&db.Finding{
		ID:                 "FIND-01",
		InvestigationID:    invID,
		Title:              "Lonjakan Volume ANTM",
		ClaimText:          "Volume transaksi meningkat tajam mendahului rilis eksplorasi",
		VerificationStatus: "SUPPORTED",
		ConfidenceScore:    0.95,
		CausalityStatus:    "PRECEDED_ANNOUNCEMENT",
	})
	if err != nil {
		t.Fatalf("failed to seed finding: %v", err)
	}

	// 2. Start test server
	srv, err := Start(ctx, 0, database)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	// 3. Test GET /api/investigations/{id} -> 200 OK
	resp, err := client.Get(srv.URL + "/api/investigations/" + invID)
	if err != nil {
		t.Fatalf("GET /api/investigations/%s failed: %v", invID, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var invData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&invData); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	resp.Body.Close()

	invObj, ok := invData["investigation"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'investigation' key in response, got %+v", invData)
	}
	if invObj["ticker"] != "ANTM" || invObj["status"] != "COMPLETED" {
		t.Errorf("unexpected investigation payload: %+v", invObj)
	}

	// 4. Test GET /api/investigations/{id}/anomalies -> 200 OK
	respAnom, err := client.Get(srv.URL + "/api/investigations/" + invID + "/anomalies")
	if err != nil {
		t.Fatalf("GET /api/investigations/%s/anomalies failed: %v", invID, err)
	}
	if respAnom.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respAnom.StatusCode)
	}
	var anomData map[string]interface{}
	if err := json.NewDecoder(respAnom.Body).Decode(&anomData); err != nil {
		t.Fatalf("failed to decode anomalies response: %v", err)
	}
	respAnom.Body.Close()

	if int(anomData["total"].(float64)) != 1 {
		t.Errorf("expected 1 anomaly, got %v", anomData["total"])
	}
	anomList, ok := anomData["anomalies"].([]interface{})
	if !ok || len(anomList) != 1 {
		t.Fatalf("expected 1 anomaly item, got %+v", anomData["anomalies"])
	}

	// 5. Test GET /api/investigations/{id}/findings -> 200 OK
	respFind, err := client.Get(srv.URL + "/api/investigations/" + invID + "/findings")
	if err != nil {
		t.Fatalf("GET /api/investigations/%s/findings failed: %v", invID, err)
	}
	if respFind.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respFind.StatusCode)
	}
	var findData map[string]interface{}
	if err := json.NewDecoder(respFind.Body).Decode(&findData); err != nil {
		t.Fatalf("failed to decode findings response: %v", err)
	}
	respFind.Body.Close()

	if int(findData["total"].(float64)) != 1 {
		t.Errorf("expected 1 finding, got %v", findData["total"])
	}
	findList, ok := findData["findings"].([]interface{})
	if !ok || len(findList) != 1 {
		t.Fatalf("expected 1 finding item, got %+v", findData["findings"])
	}

	// 6. Test GET /api/investigations/NON_EXISTENT -> 404
	respMissing, err := client.Get(srv.URL + "/api/investigations/NON_EXISTENT")
	if err != nil {
		t.Fatalf("GET missing investigation failed: %v", err)
	}
	respMissing.Body.Close()
	if respMissing.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", respMissing.StatusCode)
	}

	// 7. Test GET /api/investigations/{id}/unknown_sub_resource -> 404
	respUnknown, err := client.Get(srv.URL + "/api/investigations/" + invID + "/unknown")
	if err != nil {
		t.Fatalf("GET unknown sub-resource failed: %v", err)
	}
	respUnknown.Body.Close()
	if respUnknown.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for unknown sub-resource, got %d", respUnknown.StatusCode)
	}

	// 8. Test GET /api/investigations -> 200 OK (list investigations)
	respList, err := client.Get(srv.URL + "/api/investigations")
	if err != nil {
		t.Fatalf("GET /api/investigations list failed: %v", err)
	}
	if respList.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK from /api/investigations list, got %d", respList.StatusCode)
	}
	var listData map[string]interface{}
	_ = json.NewDecoder(respList.Body).Decode(&listData)
	respList.Body.Close()
	if int(listData["total"].(float64)) != 1 {
		t.Errorf("expected total 1 investigation, got %v", listData["total"])
	}
}

func TestSSEErrorIsValidJSON(t *testing.T) {
	testCases := []struct {
		name   string
		errMsg string
	}{
		{"simple error", "engine subprocess exited with error: exit status 1"},
		{"error with quotes", `connection refused to "localhost:8080"`},
		{"error with newline", "line1\nline2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errJSON, jsonErr := json.Marshal(map[string]interface{}{
				"event":      "session_error",
				"session_id": "TEST-123",
				"error":      tc.errMsg,
			})
			if jsonErr != nil {
				t.Fatalf("json.Marshal failed: %v", jsonErr)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(errJSON, &parsed); err != nil {
				t.Fatalf("could not unmarshal marshaled error JSON: %v", err)
			}
			if parsed["error"] != tc.errMsg {
				t.Errorf("expected error=%q, got %q", tc.errMsg, parsed["error"])
			}
			if parsed["event"] != "session_error" {
				t.Errorf("expected event='session_error', got %v", parsed["event"])
			}
		})
	}
}

func TestServerDoesNotEnforceReadTimeoutOnSSE(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := Start(ctx, 0, nil)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.httpServer.Close()

	if srv.httpServer.ReadTimeout != 0 {
		t.Errorf("expected httpServer.ReadTimeout to be 0 for SSE longevity, got %v", srv.httpServer.ReadTimeout)
	}
	if srv.httpServer.ReadHeaderTimeout == 0 {
		t.Errorf("expected httpServer.ReadHeaderTimeout to be configured to protect Slowloris, got 0")
	}
}

func TestTelegramEndpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_tg.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	cfg := config.DefaultConfig()
	cfg.Preferences.OfflineMode = true
	cfg.Telegram.BotToken = ""
	cfg.Telegram.Enabled = false
	configPath := filepath.Join(tempDir, "config.yaml")

	srv, err := Start(ctx, 0, database, cfg)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	srv.ConfigPath = configPath
	defer srv.httpServer.Close()

	time.Sleep(50 * time.Millisecond)

	// 1. Initial GET /api/settings/telegram -> NOT_CONFIGURED
	resp, err := http.Get(srv.URL + "/api/settings/telegram")
	if err != nil {
		t.Fatalf("GET /api/settings/telegram failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var view map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&view); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if view["status"] != "NOT_CONFIGURED" {
		t.Errorf("expected status 'NOT_CONFIGURED', got %v", view["status"])
	}
	if view["has_token"] != false {
		t.Errorf("expected has_token false, got %v", view["has_token"])
	}

	// 2. PATCH /api/settings/telegram -> update token and allowed_users
	patchPayload := map[string]interface{}{
		"bot_token":     "123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		"enabled":       false,
		"allowed_users": []string{"analyst_1", "998877"},
	}
	body, _ := json.Marshal(patchPayload)
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/settings/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /api/settings/telegram failed: %v", err)
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from PATCH, got %d", patchResp.StatusCode)
	}
	var patchResult map[string]interface{}
	if err := json.NewDecoder(patchResp.Body).Decode(&patchResult); err != nil {
		t.Fatalf("failed to decode PATCH response: %v", err)
	}
	if patchResult["status"] != "STOPPED" {
		t.Errorf("expected status 'STOPPED' after adding token, got %v", patchResult["status"])
	}
	if patchResult["has_token"] != true {
		t.Errorf("expected has_token true, got %v", patchResult["has_token"])
	}

	// 3. POST /api/telegram/start -> start bot daemon
	startReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/telegram/start", nil)
	startResp, err := http.DefaultClient.Do(startReq)
	if err != nil {
		t.Fatalf("POST /api/telegram/start failed: %v", err)
	}
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from POST start, got %d", startResp.StatusCode)
	}
	var startResult map[string]interface{}
	if err := json.NewDecoder(startResp.Body).Decode(&startResult); err != nil {
		t.Fatalf("failed to decode start response: %v", err)
	}
	if startResult["status"] != "RUNNING" {
		t.Errorf("expected status 'RUNNING', got %v", startResult["status"])
	}

	// Verify with GET /api/settings/telegram -> RUNNING
	getResp, err := http.Get(srv.URL + "/api/settings/telegram")
	if err != nil {
		t.Fatalf("GET /api/settings/telegram failed: %v", err)
	}
	defer getResp.Body.Close()
	var getView map[string]interface{}
	_ = json.NewDecoder(getResp.Body).Decode(&getView)
	if getView["status"] != "RUNNING" {
		t.Errorf("expected status 'RUNNING' in GET, got %v", getView["status"])
	}
	if getView["enabled"] != true {
		t.Errorf("expected enabled true after start, got %v", getView["enabled"])
	}

	// 4. POST /api/telegram/test -> test ping in offline mode
	testPayload := map[string]interface{}{"chat_id": 12345}
	testBody, _ := json.Marshal(testPayload)
	testReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/telegram/test", bytes.NewReader(testBody))
	testReq.Header.Set("Content-Type", "application/json")
	testResp, err := http.DefaultClient.Do(testReq)
	if err != nil {
		t.Fatalf("POST /api/telegram/test failed: %v", err)
	}
	defer testResp.Body.Close()
	if testResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from test ping, got %d", testResp.StatusCode)
	}
	var testResult map[string]interface{}
	_ = json.NewDecoder(testResp.Body).Decode(&testResult)
	if testResult["mock"] != true {
		t.Errorf("expected mock true in offline mode test, got %v", testResult["mock"])
	}

	// 5. POST /api/telegram/stop -> stop bot daemon
	stopReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/telegram/stop", nil)
	stopResp, err := http.DefaultClient.Do(stopReq)
	if err != nil {
		t.Fatalf("POST /api/telegram/stop failed: %v", err)
	}
	defer stopResp.Body.Close()
	if stopResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from POST stop, got %d", stopResp.StatusCode)
	}
	var stopResult map[string]interface{}
	if err := json.NewDecoder(stopResp.Body).Decode(&stopResult); err != nil {
		t.Fatalf("failed to decode stop response: %v", err)
	}
	if stopResult["status"] != "STOPPED" {
		t.Errorf("expected status 'STOPPED', got %v", stopResult["status"])
	}

	// Verify with GET /api/settings/telegram -> STOPPED
	getResp2, err := http.Get(srv.URL + "/api/settings/telegram")
	if err != nil {
		t.Fatalf("GET /api/settings/telegram failed: %v", err)
	}
	defer getResp2.Body.Close()
	var getView2 map[string]interface{}
	_ = json.NewDecoder(getResp2.Body).Decode(&getView2)
	if getView2["status"] != "STOPPED" {
		t.Errorf("expected status 'STOPPED' in GET, got %v", getView2["status"])
	}
	if getView2["enabled"] != false {
		t.Errorf("expected enabled false after stop, got %v", getView2["enabled"])
	}
}

func TestSystemEndpoints(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_srv_sys.db")

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

	// 1. GET /api/system/sectors-usage
	resp, err := http.Get(srv.URL + "/api/system/sectors-usage")
	if err != nil {
		t.Fatalf("GET /api/system/sectors-usage failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var usage map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		t.Fatalf("failed to decode usage: %v", err)
	}
	if usage["status"] != "HEALTHY" {
		t.Errorf("expected status 'HEALTHY', got %v", usage["status"])
	}

	// 2. GET /api/system/diagnostics
	resp2, err := http.Get(srv.URL + "/api/system/diagnostics")
	if err != nil {
		t.Fatalf("GET /api/system/diagnostics failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp2.StatusCode)
	}
	var diag map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&diag); err != nil {
		t.Fatalf("failed to decode diag: %v", err)
	}
	if diag["status"] != "OK" {
		t.Errorf("expected status 'OK', got %v", diag["status"])
	}
	if diag["go_version"] == "" {
		t.Errorf("expected non-empty go_version")
	}

	// 3. POST /api/system/cache/clean
	resp3, err := http.Post(srv.URL+"/api/system/cache/clean", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/system/cache/clean failed: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp3.StatusCode)
	}
	var clean map[string]interface{}
	if err := json.NewDecoder(resp3.Body).Decode(&clean); err != nil {
		t.Fatalf("failed to decode clean response: %v", err)
	}
	if clean["status"] != "ok" {
		t.Errorf("expected clean status 'ok', got %v", clean["status"])
	}
}
