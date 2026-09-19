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

	// 4. List chat sessions
	sessions, err := database.ListChatSessions(10)
	if err != nil {
		t.Fatalf("failed to list chat sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session summary, got %d", len(sessions))
	}
	if sessions[0].SessionID != sessionID || sessions[0].MessageCount != 2 {
		t.Errorf("unexpected session summary: %+v", sessions[0])
	}
}
