package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInvestigationSelectorModel_NavigationAndSelect(t *testing.T) {
	investigations := []db.Investigation{
		{
			ID:        "INV-20260920-0001",
			Ticker:    "ANTM",
			Status:    "COMPLETED",
			StartedAt: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "INV-20260921-0002",
			Ticker:    "BBCA",
			Status:    "COMPLETED",
			StartedAt: time.Now().Format(time.RFC3339),
		},
	}

	model := NewInvestigationSelectorModel(investigations)

	if model.Cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", model.Cursor)
	}

	// Move down
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(InvestigationSelectorModel)
	if m.Cursor != 1 {
		t.Fatalf("expected cursor at 1 after KeyDown, got %d", m.Cursor)
	}

	// Press Enter to select
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(InvestigationSelectorModel)
	if m.SelectedSession == nil {
		t.Fatal("expected selected session, got nil")
	}
	if m.SelectedSession.ID != "INV-20260921-0002" {
		t.Fatalf("expected INV-20260921-0002, got %s", m.SelectedSession.ID)
	}
}

func TestInvestigationSelectorModel_FilterAndExport(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_inv.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	summaryText := "Investigasi saham ANTM menemukan lonjakan volume."
	inv := &db.Investigation{
		ID:          "INV-TEST-001",
		Ticker:      "ANTM",
		Status:      "COMPLETED",
		StartedAt:   time.Now().Format(time.RFC3339),
		SummaryText: &summaryText,
	}
	_ = tmpDB.CreateInvestigation(inv)

	model := NewInvestigationSelectorModelWithDB([]db.Investigation{*inv}, tmpDB)

	// Filter query
	model.FilterQuery = "antm"
	filtered := model.getFilteredInvestigations()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 match for 'antm', got %d", len(filtered))
	}

	// Export test
	outPath, err := ExportInvestigationTranscript(tmpDB, inv, 0)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !strings.HasSuffix(outPath, ".md") {
		t.Fatalf("expected .md extension, got %s", outPath)
	}
}

func TestInvestigationSelectorModel_Delete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_inv_del.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	inv := &db.Investigation{ID: "INV-DEL-001", Ticker: "BUMI", Status: "FAILED"}
	_ = tmpDB.CreateInvestigation(inv)

	model := NewInvestigationSelectorModelWithDB([]db.Investigation{*inv}, tmpDB)

	// Press Ctrl+D
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m := updated.(InvestigationSelectorModel)
	if !m.ConfirmDelete {
		t.Fatalf("expected ConfirmDelete to be true")
	}

	// Press 'y' to confirm delete
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(InvestigationSelectorModel)

	if len(m.Investigations) != 0 {
		t.Fatalf("expected 0 investigations in memory after deletion, got %d", len(m.Investigations))
	}

	reloaded, _ := tmpDB.GetInvestigation("INV-DEL-001")
	if reloaded != nil {
		t.Fatalf("expected INV-DEL-001 to be deleted from SQLite")
	}
}

func TestInvestigationSelectorModel_PinPersistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_inv_pin.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	inv1 := &db.Investigation{ID: "INV-PIN-001", Ticker: "ANTM", Status: "COMPLETED", StartedAt: "2026-09-20T10:00:00Z"}
	inv2 := &db.Investigation{ID: "INV-PIN-002", Ticker: "BBRI", Status: "COMPLETED", StartedAt: "2026-09-21T10:00:00Z"}
	_ = tmpDB.CreateInvestigation(inv1)
	_ = tmpDB.CreateInvestigation(inv2)

	invs, err := tmpDB.ListInvestigations(10)
	if err != nil || len(invs) != 2 {
		t.Fatalf("expected 2 investigations, got %d (err: %v)", len(invs), err)
	}

	model := NewInvestigationSelectorModelWithDB(invs, tmpDB)

	pinnedID := invs[0].ID

	// Press Ctrl+P on first investigation
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m := updated.(InvestigationSelectorModel)

	if !m.Investigations[0].IsPinned {
		t.Fatalf("expected first investigation in memory to be pinned")
	}

	// Verify persistence in SQLite
	reloaded, err := tmpDB.GetInvestigation(pinnedID)
	if err != nil || reloaded == nil {
		t.Fatalf("failed to reload %s: %v", pinnedID, err)
	}
	if !reloaded.IsPinned {
		t.Fatalf("EXPECTED IsPinned to be true in SQLite database for investigation after Ctrl+P")
	}
}

