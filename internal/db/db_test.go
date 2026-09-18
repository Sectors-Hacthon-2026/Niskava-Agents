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
