package db

import (
	"path/filepath"
	"testing"
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
		ID:        sessionID,
		Title:     "Analisis Saham ANTM",
		Model:     "hermes",
		Status:    "IDLE",
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

