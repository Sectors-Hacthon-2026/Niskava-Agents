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
