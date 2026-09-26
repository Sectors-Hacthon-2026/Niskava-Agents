package cli

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
)

func TestSessionsCmd_Types(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	// Seed one chat session and one investigation
	now := time.Now().Format(time.RFC3339)
	_ = tmpDB.CreateChatSession(&db.ChatSession{
		ID:           "CHAT-TEST-01",
		Title:        "Riset Saham ANTM",
		MessageCount: 3,
		UpdatedAt:    now,
	})
	_ = tmpDB.CreateInvestigation(&db.Investigation{
		ID:        "INV-TEST-01",
		Ticker:    "ANTM",
		Status:    "COMPLETED",
		StartedAt: now,
	})

	var buf bytes.Buffer
	err = printFormattedSessions(tmpDB, "chat", 10, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("CHAT-TEST-01")) {
		t.Fatalf("expected CHAT-TEST-01 in output, got: %s", out)
	}
	if bytes.Contains(buf.Bytes(), []byte("INV-TEST-01")) {
		t.Fatalf("did not expect INV-TEST-01 in chat-only output")
	}
}

func TestSessionsCmd_InvestigationType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	now := time.Now().Format(time.RFC3339)
	_ = tmpDB.CreateInvestigation(&db.Investigation{
		ID:        "INV-TEST-99",
		Ticker:    "BBCA",
		Status:    "COMPLETED",
		StartedAt: now,
	})

	var buf bytes.Buffer
	err = printFormattedSessions(tmpDB, "investigation", 10, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("INV-TEST-99")) {
		t.Fatalf("expected INV-TEST-99 in output, got: %s", out)
	}
}

func TestSessionsSubcommands_DeleteSearchExport(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_sub.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	oldDB := appDB
	appDB = tmpDB
	defer func() { appDB = oldDB }()

	sessID := "CHAT-SUBTEST-01"
	err = tmpDB.CreateChatSession(&db.ChatSession{
		ID:        sessID,
		Title:     "Pencarian Antam",
		UpdatedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = tmpDB.SaveChatMessage(&db.ChatMessage{
		SessionID: sessID,
		Role:      "user",
		Content:   "Apakah saham ANTM mengalami anomaly volume?",
	})
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	// Test Search subcommand execution
	results, err := tmpDB.SearchChatMessages("anomaly", 10)
	if err != nil || len(results) == 0 {
		t.Fatalf("expected search result for 'anomaly', got err=%v, count=%d", err, len(results))
	}

	// Test Export subcommand functionality
	exportFile := filepath.Join(t.TempDir(), "report.md")
	exportOutFlag = exportFile
	exportFormatFlag = "md"
	err = sessionsExportCmd.RunE(sessionsExportCmd, []string{sessID})
	if err != nil {
		t.Fatalf("sessions export failed: %v", err)
	}

	// Test Delete subcommand functionality
	err = sessionsDeleteCmd.RunE(sessionsDeleteCmd, []string{sessID})
	if err != nil {
		t.Fatalf("sessions delete failed: %v", err)
	}

	// Verify session is deleted
	history, _ := tmpDB.GetChatHistory(sessID, 10)
	if len(history) != 0 {
		t.Fatalf("expected empty chat history after deletion, got %d msgs", len(history))
	}
}

