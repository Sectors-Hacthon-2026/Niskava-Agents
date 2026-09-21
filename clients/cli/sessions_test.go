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
