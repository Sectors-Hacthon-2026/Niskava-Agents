package db

import (
	"path/filepath"
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
		Model:  "hermes",
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
		Model:  "hermes",
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
		"osint_cache", "telegram_chats",
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

	cleaned, err := database.CleanExpiredCache()
	if err != nil {
		t.Fatalf("CleanExpiredCache failed: %v", err)
	}
	if cleaned != 1 {
		t.Fatalf("expected 1 cleaned item, got %d", cleaned)
	}
}
