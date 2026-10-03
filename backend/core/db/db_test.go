package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSQLiteMigrationsAndOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// 1. Create an investigation session
	inv := &Investigation{
		ID:            "INV-2026-TEST-001",
		Ticker:        "ANTM",
		Market:        "IDX",
		TimeframeDays: 30,
		Status:        "PENDING",
	}
	if err := database.CreateInvestigation(inv); err != nil {
		t.Fatalf("failed to create investigation: %v", err)
	}

	// 2. Fetch the investigation
	fetched, err := database.GetInvestigation("INV-2026-TEST-001")
	if err != nil {
		t.Fatalf("failed to fetch investigation: %v", err)
	}
	if fetched == nil {
		t.Fatalf("expected investigation record, got nil")
	}
	if fetched.Ticker != "ANTM" || fetched.Status != "PENDING" {
		t.Errorf("unexpected record values: %+v", fetched)
	}

	// 3. Update status to COMPLETED
	summary := "Anomaly investigation completed successfully with 1 likely catalyst."
	if err := database.UpdateInvestigationStatus("INV-2026-TEST-001", "COMPLETED", &summary); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	updated, err := database.GetInvestigation("INV-2026-TEST-001")
	if err != nil {
		t.Fatalf("failed to fetch updated investigation: %v", err)
	}
	if updated.Status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %s", updated.Status)
	}
	if updated.SummaryText == nil || *updated.SummaryText != summary {
		t.Errorf("expected summary text '%s', got %v", summary, updated.SummaryText)
	}
}

func TestChatMessagesPersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_chat.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	sessionID := "CHAT-TEST-001"

	// 1. Save user message
	userMsg := &ChatMessage{
		ID:        "MSG-1",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Analisis saham ANTM",
		CreatedAt: "2026-09-18T10:00:00Z",
	}
	if err := database.SaveChatMessage(userMsg); err != nil {
		t.Fatalf("failed to save user message: %v", err)
	}

	// 2. Save assistant message with thought
	thought := "Menganalisis emiten ANTM via ReAct loop..."
	asstMsg := &ChatMessage{
		ID:        "MSG-2",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Berikut adalah laporan analisis...",
		Thought:   &thought,
		CreatedAt: "2026-09-18T10:00:05Z",
	}
	if err := database.SaveChatMessage(asstMsg); err != nil {
		t.Fatalf("failed to save assistant message: %v", err)
	}

	// 3. Retrieve history
	history, err := database.GetChatHistory(sessionID, 10)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}
	if history[0].Role != "user" || history[1].Role != "assistant" {
		t.Errorf("unexpected message roles in history: %+v", history)
	}
	if history[1].Thought == nil || *history[1].Thought != thought {
		t.Errorf("expected thought '%s', got %v", thought, history[1].Thought)
	}
}

func TestMemoryGraphPersistenceAndClear(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_mem_graph.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Seed nodes, investigation, and edges directly
	_, err = database.conn.Exec(`
		INSERT INTO investigations (id, ticker, market, timeframe_days, status, started_at)
		VALUES ('S1', 'ANTM', 'IDX', 30, 'COMPLETED', CURRENT_TIMESTAMP);
		INSERT INTO memory_nodes (id, label, node_type, last_observed_at)
		VALUES ('ticker:antm', 'ANTM', 'TICKER', '2026-09-18T10:00:00Z'),
		       ('user:default', 'User', 'USER', '2026-09-18T10:00:00Z');
		INSERT INTO memory_edges (source_id, target_id, relation, weight, session_id, last_observed_at)
		VALUES ('user:default', 'ticker:antm', 'INVESTIGATED', 1.0, 'S1', '2026-09-18T10:00:00Z');
	`)
	if err != nil {
		t.Fatalf("failed to seed memory records: %v", err)
	}

	// 1. GetMemoryGraph
	nodes, edges, err := database.GetMemoryGraph("")
	if err != nil {
		t.Fatalf("failed to get memory graph: %v", err)
	}
	if len(nodes) != 2 || len(edges) != 1 {
		t.Errorf("expected 2 nodes and 1 edge, got %d nodes and %d edges", len(nodes), len(edges))
	}
	if edges[0].Relation != "INVESTIGATED" {
		t.Errorf("expected relation INVESTIGATED, got %s", edges[0].Relation)
	}

	// 2. Clear by session
	if err := database.ClearMemoryGraph("S1"); err != nil {
		t.Fatalf("failed to clear memory graph for session S1: %v", err)
	}
	nodesAfter, edgesAfter, err := database.GetMemoryGraph("")
	if err != nil {
		t.Fatalf("failed to get memory graph after clear: %v", err)
	}
	if len(edgesAfter) != 0 {
		t.Errorf("expected 0 edges after clear, got %d", len(edgesAfter))
	}
	if len(nodesAfter) != 0 {
		t.Errorf("expected 0 orphaned nodes after clear, got %d", len(nodesAfter))
	}
}

func TestChatSessionsCompleteLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sessions_lifecycle.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// 1. Create a session explicitly
	sessionID := "CHAT-2026-TEST-A"
	sess := &ChatSession{
		ID:     sessionID,
		Title:  "Analisis Saham ANTM",
		Model:  "gemini-2.0-flash",
		Status: "IDLE",
	}
	if err := database.CreateChatSession(sess); err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	// 2. Fetch session
	fetched, err := database.GetChatSession(sessionID)
	if err != nil {
		t.Fatalf("failed to get chat session: %v", err)
	}
	if fetched == nil || fetched.Title != "Analisis Saham ANTM" {
		t.Fatalf("expected title 'Analisis Saham ANTM', got %+v", fetched)
	}

	// 3. Save messages and verify TouchChatSession updates preview & count
	userMsg := &ChatMessage{
		ID:        "M1",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Kenapa volume ANTM melonjak?",
	}
	if err := database.SaveChatMessage(userMsg); err != nil {
		t.Fatalf("failed to save user msg: %v", err)
	}

	asstMsg := &ChatMessage{
		ID:        "M2",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Terdeteksi anomali volume Z-score 3.12 dengan inflow asing.",
	}
	if err := database.SaveChatMessage(asstMsg); err != nil {
		t.Fatalf("failed to save asst msg: %v", err)
	}

	// Also add a memory edge with this session_id to test cascade
	_, err = database.conn.Exec(`
		INSERT INTO memory_nodes (id, label, node_type, last_observed_at) VALUES ('ticker:antm', 'ANTM', 'TICKER', CURRENT_TIMESTAMP);
		INSERT INTO memory_edges (source_id, target_id, relation, session_id, last_observed_at) VALUES ('ticker:antm', 'ticker:antm', 'SELF', 'CHAT-2026-TEST-A', CURRENT_TIMESTAMP);
	`)
	if err != nil {
		t.Fatalf("failed to seed memory edge: %v", err)
	}

	// Check updated stats
	updatedSess, err := database.GetChatSession(sessionID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}
	if updatedSess.MessageCount != 2 {
		t.Errorf("expected message count 2, got %d", updatedSess.MessageCount)
	}

	// 4. Update session title & pinning
	newTitle := "Analisis Saham Nikel ANTM & INCO"
	isPinned := true
	status := "BUSY"
	if err := database.UpdateChatSession(sessionID, &newTitle, &isPinned, &status); err != nil {
		t.Fatalf("failed to update session: %v", err)
	}

	checkedSess, _ := database.GetChatSession(sessionID)
	if !checkedSess.IsPinned || checkedSess.Title != newTitle || checkedSess.Status != "BUSY" {
		t.Errorf("update session failed, got: %+v", checkedSess)
	}

	// 5. List sessions with search and pagination
	sessions, total, err := database.ListChatSessions(10, 0, "Nikel")
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	if total != 1 || len(sessions) != 1 {
		t.Errorf("expected 1 session matching 'Nikel', got total %d, returned %d", total, len(sessions))
	}

	// 6. Fork session (OpenCode pattern)
	forkedID := "CHAT-2026-FORK-B"
	forkTitle := "Eksplorasi Cabang ANTM"
	if err := database.ForkChatSession(sessionID, forkedID, forkTitle, ""); err != nil {
		t.Fatalf("failed to fork chat session: %v", err)
	}

	forkedSess, err := database.GetChatSession(forkedID)
	if err != nil || forkedSess == nil {
		t.Fatalf("failed to get forked session: %v", err)
	}
	if forkedSess.Title != forkTitle || forkedSess.ParentSessionID == nil || *forkedSess.ParentSessionID != sessionID {
		t.Errorf("unexpected forked session attributes: %+v", forkedSess)
	}

	forkedHistory, err := database.GetChatHistory(forkedID, 10)
	if err != nil || len(forkedHistory) != 2 {
		t.Errorf("expected 2 messages copied into forked session, got %d", len(forkedHistory))
	}

	// 7. Clear history for forked session
	if err := database.ClearSessionHistory(forkedID); err != nil {
		t.Fatalf("failed to clear session history: %v", err)
	}
	clearedHistory, _ := database.GetChatHistory(forkedID, 10)
	if len(clearedHistory) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(clearedHistory))
	}

	// 8. Delete source session and verify cascade delete of messages & scoped edges
	if err := database.DeleteChatSession(sessionID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
	deletedCheck, _ := database.GetChatSession(sessionID)
	if deletedCheck != nil {
		t.Errorf("expected session to be nil after delete, got %+v", deletedCheck)
	}
	messagesAfterDelete, _ := database.GetChatHistory(sessionID, 10)
	if len(messagesAfterDelete) != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", len(messagesAfterDelete))
	}
}

func TestDB_Path_SelfHealing_And_SearchMessages(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_healing_search.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// 1. Verify db.Path is populated with absolute path
	if db1.Path == "" {
		t.Fatal("expected db.Path to be populated, got empty string")
	}
	absExpected, _ := filepath.Abs(dbPath)
	if db1.Path != absExpected {
		t.Errorf("expected db.Path=%s, got %s", absExpected, db1.Path)
	}

	// 2. Create a session and set it to BUSY to simulate a running session when server terminates
	sessionID := "SESS-ZOMBIE-001"
	err = db1.CreateChatSession(&ChatSession{
		ID:     sessionID,
		Title:  "Analisis ANTM",
		Model:  "gemini-2.0-flash",
		Status: "BUSY",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add messages for search testing
	_ = db1.SaveChatMessage(&ChatMessage{
		ID:        "MSG-01",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Bagaimana rumor Rights Issue ANTM?",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	_ = db1.SaveChatMessage(&ChatMessage{
		ID:        "MSG-02",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Berdasarkan keterbukaan informasi, belum ada dokumen rights issue resmi untuk emiten ANTM.",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})

	// Close db1
	db1.Close()

	// 3. Re-open database (simulate daemon restart) and verify self-healing reset BUSY -> IDLE
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to re-open database: %v", err)
	}
	defer db2.Close()

	sess, err := db2.GetChatSession(sessionID)
	if err != nil || sess == nil {
		t.Fatalf("failed to retrieve session after restart: %v", err)
	}
	if sess.Status != "IDLE" {
		t.Errorf("expected self-healing status to be IDLE, got %s", sess.Status)
	}

	// 4. Test SearchChatMessages
	results, err := db2.SearchChatMessages("Rights Issue", 10)
	if err != nil {
		t.Fatalf("SearchChatMessages failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results matching 'Rights Issue', got %d", len(results))
	}
	if len(results) > 0 && results[0].SessionTitle != "Analisis ANTM" {
		t.Errorf("expected session_title 'Analisis ANTM', got %s", results[0].SessionTitle)
	}

	// Search non-existent
	emptyResults, err := db2.SearchChatMessages("SahamTidakAdaXYZ", 10)
	if err != nil {
		t.Fatalf("SearchChatMessages failed for nonexistent: %v", err)
	}
	if len(emptyResults) != 0 {
		t.Errorf("expected 0 results, got %d", len(emptyResults))
	}

	// Search empty query
	blankResults, err := db2.SearchChatMessages("", 10)
	if err != nil || len(blankResults) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(blankResults))
	}
}

func TestAnomaliesAndFindingsByInvestigation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_anom_find.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	invID := "INV-DB-TEST-001"
	err = database.CreateInvestigation(&Investigation{
		ID:            invID,
		Ticker:        "BBCA",
		Market:        "IDX",
		TimeframeDays: 30,
		Status:        "COMPLETED",
	})
	if err != nil {
		t.Fatalf("failed to create investigation: %v", err)
	}

	// Verify empty queries return non-nil empty slices
	anoms, err := database.GetAnomaliesByInvestigation(invID)
	if err != nil {
		t.Fatalf("GetAnomaliesByInvestigation failed: %v", err)
	}
	if anoms == nil || len(anoms) != 0 {
		t.Errorf("expected empty non-nil slice, got %+v", anoms)
	}

	findings, err := database.ListFindingsByInvestigation(invID)
	if err != nil {
		t.Fatalf("ListFindingsByInvestigation failed: %v", err)
	}
	if findings == nil || len(findings) != 0 {
		t.Errorf("expected empty non-nil slice, got %+v", findings)
	}

	// Insert anomaly
	anom := &Anomaly{
		ID:              "A1",
		InvestigationID: invID,
		AnomalyDate:     "2026-09-15",
		MetricType:      "volume_z_score",
		MetricValue:     20000000,
		BaselineValue:   5000000,
		ZScore:          3.8,
		Description:     "Huge volume spike",
	}
	if err := database.CreateAnomaly(anom); err != nil {
		t.Fatalf("CreateAnomaly failed: %v", err)
	}

	// Insert finding
	finding := &Finding{
		ID:                 "F1",
		InvestigationID:    invID,
		Title:              "Akumulasi Asing BBCA",
		ClaimText:          "Inflow asing masif terdeteksi",
		VerificationStatus: "SUPPORTED",
		ConfidenceScore:    1.00,
		CausalityStatus:    "LIKELY_CATALYST",
	}
	if err := database.CreateFinding(finding); err != nil {
		t.Fatalf("CreateFinding failed: %v", err)
	}

	// Fetch again
	anomsAfter, err := database.GetAnomaliesByInvestigation(invID)
	if err != nil || len(anomsAfter) != 1 {
		t.Fatalf("expected 1 anomaly, got %d (err: %v)", len(anomsAfter), err)
	}
	if anomsAfter[0].MetricType != "volume_z_score" || anomsAfter[0].ZScore != 3.8 {
		t.Errorf("unexpected anomaly content: %+v", anomsAfter[0])
	}

	findingsAfter, err := database.ListFindingsByInvestigation(invID)
	if err != nil || len(findingsAfter) != 1 {
		t.Fatalf("expected 1 finding, got %d (err: %v)", len(findingsAfter), err)
	}
	if findingsAfter[0].Title != "Akumulasi Asing BBCA" || findingsAfter[0].ConfidenceScore != 1.00 {
		t.Errorf("unexpected finding content: %+v", findingsAfter[0])
	}
}

func TestAllSpecificationTablesCreated(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	expectedTables := []string{
		"investigations", "anomalies", "findings", "evidence_items",
		"timeline_events", "sectors_cache", "memory_nodes", "memory_edges",
		"chat_sessions", "chat_messages", "suspension_records", "insider_filings",
		"news_cache", "telegram_chats",
	}

	for _, tbl := range expectedTables {
		var count int
		err := database.conn.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&count)
		if err != nil || count == 0 {
			t.Errorf("expected table '%s' to exist in database schema", tbl)
		}
	}
}

func TestTelegramChatSessionOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telegram_test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	chatID := int64(987654321)
	userID := int64(12345)
	username := "investor_idx"

	// 1. First get or create - should create a new session
	sessID1, err := database.GetOrCreateTelegramChatSession(chatID, userID, username)
	if err != nil {
		t.Fatalf("failed to get or create session: %v", err)
	}
	if sessID1 == "" {
		t.Fatalf("expected non-empty session ID")
	}

	// Verify chat record exists
	chatRecord, err := database.GetTelegramChat(chatID)
	if err != nil {
		t.Fatalf("failed to get telegram chat: %v", err)
	}
	if chatRecord == nil || chatRecord.CurrentSessionID != sessID1 {
		t.Fatalf("expected chat record current_session_id %s, got %+v", sessID1, chatRecord)
	}

	// Verify chat_sessions record exists
	chatSession, err := database.GetChatSession(sessID1)
	if err != nil {
		t.Fatalf("failed to get chat session: %v", err)
	}
	if chatSession == nil {
		t.Fatalf("expected session %s in chat_sessions", sessID1)
	}

	// 2. Second call should return the SAME session ID
	sessID2, err := database.GetOrCreateTelegramChatSession(chatID, userID, username)
	if err != nil {
		t.Fatalf("failed on second get or create: %v", err)
	}
	if sessID2 != sessID1 {
		t.Errorf("expected same session ID %s, got %s", sessID1, sessID2)
	}

	// 3. Reset session - should generate a NEW session ID
	sessID3, err := database.ResetTelegramChatSession(chatID, userID, username)
	if err != nil {
		t.Fatalf("failed to reset session: %v", err)
	}
	if sessID3 == sessID1 {
		t.Errorf("expected new session ID after reset, got same %s", sessID3)
	}

	// Verify DB now maps chat to the new session
	chatRecordAfterReset, err := database.GetTelegramChat(chatID)
	if err != nil {
		t.Fatalf("failed to get telegram chat after reset: %v", err)
	}
	if chatRecordAfterReset.CurrentSessionID != sessID3 {
		t.Errorf("expected current session ID %s, got %s", sessID3, chatRecordAfterReset.CurrentSessionID)
	}
}

func TestFilteredMemoryGraphAndStats(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_graph_filter.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Insert test nodes
	nodes := []MemoryNode{
		{ID: "node:ANTM", Label: "ANTM", NodeType: "TICKER", LastObservedAt: time.Now().Format(time.RFC3339)},
		{ID: "node:NICKEL", Label: "Nickel Commodity", NodeType: "CATALYST", LastObservedAt: time.Now().Format(time.RFC3339)},
		{ID: "node:INCO", Label: "INCO", NodeType: "TICKER", LastObservedAt: time.Now().Format(time.RFC3339)},
	}
	for _, n := range nodes {
		if err := database.SaveMemoryNode(&n); err != nil {
			t.Fatalf("failed to save node: %v", err)
		}
	}

	// Insert test edges
	sess1 := "sess-1"
	sess2 := "sess-2"
	edges := []MemoryEdge{
		{SourceID: "node:ANTM", TargetID: "node:NICKEL", Relation: "EXPOSED_TO", Weight: 2.5, SessionID: &sess1, LastObservedAt: time.Now().Format(time.RFC3339)},
		{SourceID: "node:INCO", TargetID: "node:NICKEL", Relation: "EXPOSED_TO", Weight: 1.0, SessionID: &sess2, LastObservedAt: time.Now().Format(time.RFC3339)},
	}
	for _, e := range edges {
		if err := database.SaveMemoryEdge(&e); err != nil {
			t.Fatalf("failed to save edge: %v", err)
		}
	}

	// 1. Filter by ticker / ego network
	fNodes, fEdges, err := database.GetFilteredMemoryGraph(MemoryGraphFilter{
		Ticker: "ANTM",
		Depth:  1,
	})
	if err != nil {
		t.Fatalf("GetFilteredMemoryGraph failed: %v", err)
	}
	if len(fNodes) != 2 {
		t.Fatalf("expected 2 nodes in ANTM ego graph, got %d", len(fNodes))
	}
	if len(fEdges) != 1 {
		t.Fatalf("expected 1 edge in ANTM ego graph, got %d", len(fEdges))
	}

	// 2. Stats
	stats, err := database.GetMemoryGraphStats()
	if err != nil {
		t.Fatalf("GetMemoryGraphStats failed: %v", err)
	}
	if stats.TotalNodes != 3 || stats.TotalEdges != 2 {
		t.Fatalf("expected 3 nodes and 2 edges in stats, got %d nodes and %d edges", stats.TotalNodes, stats.TotalEdges)
	}
	if stats.NodeTypes["TICKER"] != 2 || stats.NodeTypes["CATALYST"] != 1 {
		t.Fatalf("unexpected node type breakdown: %+v", stats.NodeTypes)
	}
}

func TestSectorsCacheStatsAndClean(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cache_stats.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Insert permanent cache
	err = database.SetSectorsCache("key1", "/v2/daily/BBCA", `{"ok":true}`, nil)
	if err != nil {
		t.Fatalf("SetSectorsCache failed: %v", err)
	}

	// Insert expired cache
	past := time.Now().Add(-2 * time.Hour)
	err = database.SetSectorsCache("key2", "/v2/company/BBCA", `{"ok":true}`, &past)
	if err != nil {
		t.Fatalf("SetSectorsCache failed: %v", err)
	}

	stats, err := database.GetSectorsCacheStats()
	if err != nil {
		t.Fatalf("GetSectorsCacheStats failed: %v", err)
	}
	if stats.TotalEntries != 2 || stats.ExpiredEntries != 1 || stats.PermanentEntries != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	cachedCandles, err := database.GetCachedDailyCandles("BBCA")
	if err != nil {
		t.Fatalf("GetCachedDailyCandles failed: %v", err)
	}
	if cachedCandles != `{"ok":true}` {
		t.Fatalf("expected payload `{\"ok\":true}`, got %s", cachedCandles)
	}

	cleaned, err := database.CleanExpiredCache()
	if err != nil {
		t.Fatalf("CleanExpiredCache failed: %v", err)
	}
	if cleaned != 1 {
		t.Fatalf("expected 1 cleaned item, got %d", cleaned)
	}
}

func TestLegacyOSINTCacheMigration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration_test.db")

	// First create legacy database with osint_cache table
	rawConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open raw sqlite: %v", err)
	}
	_, err = rawConn.Exec(`
		CREATE TABLE osint_cache (
			cache_key TEXT PRIMARY KEY,
			source_type TEXT NOT NULL,
			query_or_url TEXT NOT NULL,
			content_text TEXT NOT NULL,
			metadata_json TEXT,
			expires_at TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO osint_cache (cache_key, source_type, query_or_url, content_text)
		VALUES ('test_key', 'news', 'https://example.com', 'test content');
	`)
	if err != nil {
		t.Fatalf("failed to seed legacy table: %v", err)
	}
	rawConn.Close()

	// Now open using DB wrapper Open(), which triggers migration
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed on legacy database: %v", err)
	}
	defer database.Close()

	// Verify news_cache exists and contains the migrated data
	var count int
	err = database.conn.QueryRow("SELECT count(*) FROM news_cache WHERE cache_key = 'test_key'").Scan(&count)
	if err != nil {
		t.Fatalf("query on news_cache failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row in news_cache, got %d", count)
	}
}

func TestPruneMockTestData(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "prune_test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	// Insert mock eval records
	_, _ = database.conn.Exec("INSERT INTO memory_nodes (id, label, node_type, last_observed_at) VALUES ('mock:node1', 'Mock 1', 'TICKER', CURRENT_TIMESTAMP)")
	_, _ = database.conn.Exec("INSERT INTO memory_nodes (id, label, node_type, last_observed_at) VALUES ('mock:node2', 'Mock 2', 'CATALYST_EVENT', CURRENT_TIMESTAMP)")
	_, _ = database.conn.Exec("INSERT INTO memory_edges (source_id, target_id, relation, session_id, last_observed_at) VALUES ('mock:node1', 'mock:node2', 'RELATES', 'EVAL-TEST-01', CURRENT_TIMESTAMP)")

	pruned, err := database.PruneMockTestData()
	if err != nil {
		t.Fatalf("PruneMockTestData failed: %v", err)
	}
	if pruned < 1 {
		t.Errorf("Expected at least 1 pruned edge, got %d", pruned)
	}
}

func TestListChatSessionsFiltersEmptySessions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_filter_empty.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	// Create a session with messages (should appear)
	sessWithMsg := &ChatSession{
		ID:                 "TELE-20260930-0001",
		Title:              "Telegram (@testuser)",
		Model:              "gemini-2.0-flash",
		Status:             "IDLE",
		MessageCount:       3,
		LastMessagePreview: "Analisis ANTM volume anomaly...",
	}
	if err := database.CreateChatSession(sessWithMsg); err != nil {
		t.Fatalf("failed to create session with messages: %v", err)
	}

	// Create a ghost session with zero messages (should be filtered out)
	ghostSess := &ChatSession{
		ID:                 "TELE-20260930-0002",
		Title:              "Telegram (@testuser)",
		Model:              "gemini-2.0-flash",
		Status:             "IDLE",
		MessageCount:       0,
		LastMessagePreview: "",
	}
	if err := database.CreateChatSession(ghostSess); err != nil {
		t.Fatalf("failed to create ghost session: %v", err)
	}

	sessions, total, err := database.ListChatSessions(50, 0, "")
	if err != nil {
		t.Fatalf("ListChatSessions failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1 (ghost filtered), got %d", total)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "TELE-20260930-0001" {
		t.Errorf("expected non-ghost session, got %s", sessions[0].ID)
	}
}

func TestResetTelegramChatSessionTitleIsUnique(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_unique_title.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	chatID := int64(111222333)
	userID := int64(55566677)
	username := "trader_idx"

	sessID1, err := database.ResetTelegramChatSession(chatID, userID, username)
	if err != nil {
		t.Fatalf("first reset failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	sessID2, err := database.ResetTelegramChatSession(chatID, userID, username)
	if err != nil {
		t.Fatalf("second reset failed: %v", err)
	}

	s1, err := database.GetChatSession(sessID1)
	if err != nil || s1 == nil {
		t.Fatalf("failed to get session 1: %v", err)
	}
	s2, err := database.GetChatSession(sessID2)
	if err != nil || s2 == nil {
		t.Fatalf("failed to get session 2: %v", err)
	}

	if s1.Title == s2.Title {
		t.Errorf("expected unique session titles, both got: %q", s1.Title)
	}
	if !strings.Contains(s1.Title, "@trader_idx") {
		t.Errorf("expected title to contain @username, got: %q", s1.Title)
	}
}

func TestEnsureInvestigationSessionAndAnomalyPersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_anom.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	sessionID := "CHAT-20261003-9999"
	err = database.EnsureInvestigationSession(sessionID, "BBRI")
	if err != nil {
		t.Fatalf("EnsureInvestigationSession failed: %v", err)
	}

	inv, err := database.GetInvestigation(sessionID)
	if err != nil || inv == nil {
		t.Fatalf("expected investigation record created, got: %v", err)
	}
	if inv.Ticker != "BBRI" {
		t.Errorf("expected ticker BBRI, got %s", inv.Ticker)
	}

	anom := &Anomaly{
		ID:              "ANOM-TEST-1",
		InvestigationID: sessionID,
		AnomalyDate:     "2026-09-10",
		MetricType:      "VOLUME_Z_SCORE",
		MetricValue:     5000000,
		BaselineValue:   140000,
		ZScore:          35.71,
		Description:     "Lonjakan Volume Ekstrem BBRI",
	}

	if err := database.CreateAnomaly(anom); err != nil {
		t.Fatalf("CreateAnomaly failed: %v", err)
	}

	// Idempotency check: CreateAnomaly with INSERT OR REPLACE
	if err := database.CreateAnomaly(anom); err != nil {
		t.Fatalf("CreateAnomaly repeat failed: %v", err)
	}

	anomalies, err := database.GetAnomaliesByInvestigation(sessionID)
	if err != nil {
		t.Fatalf("GetAnomaliesByInvestigation failed: %v", err)
	}
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}
	if anomalies[0].ZScore != 35.71 {
		t.Errorf("expected Z-score 35.71, got %f", anomalies[0].ZScore)
	}
}

func TestForkChatSessionClonesAnomaliesAndFindings(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_fork_anom.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	sourceID := "CHAT-SRC-100"
	_ = database.CreateChatSession(&ChatSession{ID: sourceID, Title: "Source Chat", Model: "niskava"})
	_ = database.EnsureInvestigationSession(sourceID, "BBRI")

	// Insert 3 anomalies for sourceID
	for i := 1; i <= 3; i++ {
		_ = database.CreateAnomaly(&Anomaly{
			ID:              fmt.Sprintf("ANOM-SRC-%d", i),
			InvestigationID: sourceID,
			AnomalyDate:     fmt.Sprintf("2026-09-%02d", i),
			MetricType:      "VOLUME_SURGE",
			MetricValue:     float64(i * 1000000),
			BaselineValue:   200000,
			ZScore:          float64(i) * 3.5,
			Description:     fmt.Sprintf("Anomaly %d", i),
		})
	}

	forkID := "CHAT-FORK-200"
	err = database.ForkChatSession(sourceID, forkID, "Forked Session Test", "")
	if err != nil {
		t.Fatalf("ForkChatSession failed: %v", err)
	}

	// Verify forked session has all 3 anomalies cloned
	anomForked, err := database.GetAnomaliesByInvestigation(forkID)
	if err != nil {
		t.Fatalf("GetAnomaliesByInvestigation for forked session failed: %v", err)
	}
	if len(anomForked) != 3 {
		t.Fatalf("expected 3 anomalies cloned in forked session, got %d", len(anomForked))
	}
	if anomForked[2].ZScore != 10.5 {
		t.Errorf("expected 3rd anomaly Z-score 10.5, got %f", anomForked[2].ZScore)
	}
}

func TestListLatestRadarAnomalies(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "test_radar.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	inv := &Investigation{
		ID:            "INV-RADAR-01",
		Ticker:        "ANTM",
		Market:        "IDX",
		TimeframeDays: 30,
		Status:        "COMPLETED",
	}
	if err := database.CreateInvestigation(inv); err != nil {
		t.Fatalf("failed to create investigation: %v", err)
	}

	anom := &Anomaly{
		ID:              "ANOM-RADAR-01",
		InvestigationID: inv.ID,
		AnomalyDate:     "2026-09-12",
		MetricType:      "volume_z_score",
		MetricValue:     184500000,
		BaselineValue:   48200000,
		ZScore:          3.84,
		Description:     "Volume anomaly +3.84σ detected",
	}
	if err := database.CreateAnomaly(anom); err != nil {
		t.Fatalf("failed to create anomaly: %v", err)
	}

	items, err := database.ListLatestRadarAnomalies(10, 2.0)
	if err != nil {
		t.Fatalf("ListLatestRadarAnomalies failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 anomaly item, got %d", len(items))
	}
	if items[0].Ticker != "ANTM" || items[0].ZScore != 3.84 {
		t.Fatalf("unexpected item values: %+v", items[0])
	}
}

func TestListLatestRadarAnomalies_ExcludesEvalAndTestSessions(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "test_radar_filter.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	seed := []struct {
		invID, ticker, anomID string
		z                     float64
	}{
		{"INV-REAL-01", "BBCA", "ANOM-REAL-01", 3.10},
		{"EVAL-BBRI-1789902142", "BBRI", "ANOM-EVAL-01", 35.71},
		{"TEST-ANTM-01", "ANTM", "ANOM-TEST-01", 35.71},
	}
	for _, s := range seed {
		inv := &Investigation{ID: s.invID, Ticker: s.ticker, Market: "IDX", TimeframeDays: 30, Status: "COMPLETED"}
		if err := database.CreateInvestigation(inv); err != nil {
			t.Fatalf("failed to create investigation %s: %v", s.invID, err)
		}
		anom := &Anomaly{
			ID:              s.anomID,
			InvestigationID: s.invID,
			AnomalyDate:     "2026-09-12",
			MetricType:      "VOLUME_SPIKE",
			MetricValue:     50000000,
			BaselineValue:   20000000,
			ZScore:          s.z,
			Description:     "seeded anomaly",
		}
		if err := database.CreateAnomaly(anom); err != nil {
			t.Fatalf("failed to create anomaly %s: %v", s.anomID, err)
		}
	}

	items, err := database.ListLatestRadarAnomalies(10, 2.0)
	if err != nil {
		t.Fatalf("ListLatestRadarAnomalies failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only the real anomaly, got %d items: %+v", len(items), items)
	}
	if items[0].InvestigationID != "INV-REAL-01" {
		t.Errorf("expected INV-REAL-01, got %s", items[0].InvestigationID)
	}
}

func TestGetInvestigationTimeline_MergesStoredAndDerivedEvents(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "test_timeline.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	const invID = "INV-TIMELINE-01"
	if err := database.CreateInvestigation(&Investigation{ID: invID, Ticker: "ANTM", Market: "IDX", TimeframeDays: 30, Status: "COMPLETED"}); err != nil {
		t.Fatalf("failed to create investigation: %v", err)
	}
	if err := database.CreateAnomaly(&Anomaly{
		ID: "ANOM-TL-01", InvestigationID: invID, AnomalyDate: "2026-09-18",
		MetricType: "VOLUME_SPIKE", MetricValue: 15000000, BaselineValue: 5000000, ZScore: 3.45,
		Description: "Volume spike detected",
	}); err != nil {
		t.Fatalf("failed to create anomaly: %v", err)
	}
	if _, err := database.conn.Exec(
		`INSERT INTO timeline_events (id, investigation_id, event_timestamp, event_type, headline, details) VALUES (?, ?, ?, ?, ?, ?)`,
		"TL-01", invID, "2026-09-17 16:30:00", "DISCLOSURE", "Keterbukaan informasi smelter", "IDXnet filing",
	); err != nil {
		t.Fatalf("failed to seed timeline event: %v", err)
	}

	events, err := database.GetInvestigationTimeline(invID)
	if err != nil {
		t.Fatalf("GetInvestigationTimeline failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events (1 stored + 1 derived anomaly), got %d: %+v", len(events), events)
	}
	if events[0].EventType != "DISCLOSURE" || events[1].EventType != "QUANT_ANOMALY" {
		t.Errorf("expected chronological order DISCLOSURE -> QUANT_ANOMALY, got %s -> %s", events[0].EventType, events[1].EventType)
	}
	if events[1].EventTimestamp != "2026-09-18" {
		t.Errorf("derived anomaly event must use the real anomaly date, got %q", events[1].EventTimestamp)
	}
}

func TestGetInvestigationTimeline_EmptyReturnsNonNilSlice(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "test_timeline_empty.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	events, err := database.GetInvestigationTimeline("INV-DOES-NOT-EXIST")
	if err != nil {
		t.Fatalf("GetInvestigationTimeline failed: %v", err)
	}
	if events == nil || len(events) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", events)
	}
}
