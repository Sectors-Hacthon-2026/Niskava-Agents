package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionSelectorModel_NavigationAndSelect(t *testing.T) {
	prev := "Riset BBCA"
	sessions := []db.ChatSession{
		{
			ID:                 "CHAT-20260920-0001",
			Title:              "Riset Saham ANTM",
			MessageCount:       4,
			LastMessagePreview: "Volume melonjak 3.2x",
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		{
			ID:                 "CHAT-20260921-0002",
			Title:              "Valuasi Saham BBCA",
			MessageCount:       2,
			LastMessagePreview: prev,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
	}

	model := NewSessionSelectorModel(sessions)

	// Initial cursor should be at 0
	if model.Cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", model.Cursor)
	}

	// Move down
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(SessionSelectorModel)
	if m.Cursor != 1 {
		t.Fatalf("expected cursor at 1 after KeyDown, got %d", m.Cursor)
	}

	// Press Enter to select
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SessionSelectorModel)
	if m.SelectedSession == nil {
		t.Fatal("expected selected session, got nil")
	}
	if m.SelectedSession.ID != "CHAT-20260921-0002" {
		t.Fatalf("expected CHAT-20260921-0002, got %s", m.SelectedSession.ID)
	}
	if m.Canceled {
		t.Fatal("expected Canceled to be false")
	}
}

func TestSessionSelectorModel_Cancel(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-1", Title: "Sesi 1"},
	}
	model := NewSessionSelectorModel(sessions)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(SessionSelectorModel)
	if !m.Canceled {
		t.Fatal("expected Canceled to be true upon Esc")
	}
	if m.SelectedSession != nil {
		t.Fatal("expected nil SelectedSession upon Esc")
	}
}

func TestSessionSelectorModel_EmptyList(t *testing.T) {
	model := NewSessionSelectorModel([]db.ChatSession{})
	view := model.View()
	if view == "" {
		t.Fatal("expected non-empty view for empty sessions list")
	}
}

func TestSessionSelectorModel_SanitizeMultilinePreview(t *testing.T) {
	sessions := []db.ChatSession{
		{
			ID:                 "CHAT-20260921-3076",
			Title:              "Sesi Riset Pasar",
			MessageCount:       4,
			LastMessagePreview: "### ⚠️ Gagal Terhubung ke Provider AI\n\n - **Endpoint...",
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
	}
	model := NewSessionSelectorModel(sessions)
	view := model.View()

	if strings.Contains(view, "\n - **Endpoint") {
		t.Fatalf("expected multiline preview to be sanitized into a single line, got raw newline in view: %s", view)
	}
}

func TestSanitizePreviewText(t *testing.T) {
	raw := "### ⚠️ Gagal Terhubung ke Provider AI\n\n - **Endpoint..."
	cleaned := SanitizePreviewText(raw)
	expected := "Gagal Terhubung ke Provider AI - Endpoint..."
	if cleaned != expected {
		t.Fatalf("expected '%s', got '%s'", expected, cleaned)
	}
}

func TestSessionSelectorModel_LiveKeywordFilter(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-1", Title: "Riset Saham ANTM", LastMessagePreview: "Volume naik"},
		{ID: "CHAT-2", Title: "Valuasi Saham BBCA", LastMessagePreview: "Diskon 5%"},
	}
	model := NewSessionSelectorModel(sessions)

	model.FilterQuery = "bbca"
	filtered := model.getFilteredSessions()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 matching session for 'bbca', got %d", len(filtered))
	}
	if filtered[0].ID != "CHAT-2" {
		t.Fatalf("expected CHAT-2 for 'bbca', got %s", filtered[0].ID)
	}

	view := model.View()
	if !strings.Contains(view, "bbca") || !strings.Contains(view, "Filter:") {
		t.Fatalf("expected view to contain filter header with 'bbca', got: %s", view)
	}
}

func TestSessionSelectorModel_PinHotkey(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-1", Title: "Sesi ANTM", IsPinned: false, UpdatedAt: "2026-09-20T10:00:00Z"},
		{ID: "CHAT-2", Title: "Sesi BBCA", IsPinned: false, UpdatedAt: "2026-09-21T10:00:00Z"},
	}
	model := NewSessionSelectorModel(sessions)

	// Press Ctrl+P on first session
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m := updated.(SessionSelectorModel)

	if !m.Sessions[0].IsPinned {
		t.Fatalf("expected first session in sorted list to be pinned")
	}
	if !strings.Contains(m.View(), "📌") {
		t.Fatalf("expected view to render pin badge 📌")
	}
}

func TestSessionSelectorModel_DeleteConfirmation(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-1", Title: "Sesi ANTM"},
	}
	model := NewSessionSelectorModel(sessions)

	// Press Ctrl+D to trigger deletion dialog
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m := updated.(SessionSelectorModel)

	if !m.ConfirmDelete {
		t.Fatalf("expected ConfirmDelete to be true")
	}

	// Press 'n' to cancel deletion
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(SessionSelectorModel)
	if m.ConfirmDelete {
		t.Fatalf("expected ConfirmDelete to be false after 'n'")
	}
}

func TestSessionSelectorModel_ExportModal(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-1", Title: "Sesi ANTM", LastMessagePreview: "Hasil investigasi ANTM"},
	}
	model := NewSessionSelectorModel(sessions)

	// Press Ctrl+E to open export modal
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	m := updated.(SessionSelectorModel)

	if !m.ExportModalActive {
		t.Fatalf("expected ExportModalActive to be true")
	}

	view := m.View()
	if !strings.Contains(view, "EXPORT SESSION TRANSCRIPT") {
		t.Fatalf("expected view to render export modal title")
	}

	// Press Enter to export
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SessionSelectorModel)
	if m.ExportModalActive {
		t.Fatalf("expected ExportModalActive to be false after Enter")
	}
}

func TestExportSessionTranscript(t *testing.T) {
	session := &db.ChatSession{
		ID:                 "TEST-EXPORT-001",
		Title:              "Riset ANTM Export Test",
		LastMessagePreview: "Hasil pengamatan ANTM",
		UpdatedAt:          time.Now().Format(time.RFC3339),
	}

	// Test Markdown export (format 0)
	mdPath, err := ExportSessionTranscript(nil, session, 0)
	if err != nil {
		t.Fatalf("failed to export markdown: %v", err)
	}
	if !strings.HasSuffix(mdPath, ".md") {
		t.Fatalf("expected .md extension, got %s", mdPath)
	}

	// Test JSON export (format 1)
	jsonPath, err := ExportSessionTranscript(nil, session, 1)
	if err != nil {
		t.Fatalf("failed to export json: %v", err)
	}
	if !strings.HasSuffix(jsonPath, ".json") {
		t.Fatalf("expected .json extension, got %s", jsonPath)
	}
}

func TestControlKeysDoNotCorruptFilterQuery(t *testing.T) {
	sessions := []db.ChatSession{
		{ID: "CHAT-001", Title: "Riset Saham ANTM"},
	}
	model := NewSessionSelectorModel(sessions)

	// Simulate Ctrl+P keypress with rune 16
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP, Runes: []rune{16}})
	m := updated.(SessionSelectorModel)

	if m.FilterQuery != "" {
		t.Fatalf("expected FilterQuery to remain empty after Ctrl+P, got: %q", m.FilterQuery)
	}

	// Simulate Ctrl+D keypress with rune 4
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlD, Runes: []rune{4}})
	m = updated.(SessionSelectorModel)

	if m.FilterQuery != "" {
		t.Fatalf("expected FilterQuery to remain empty after Ctrl+D, got: %q", m.FilterQuery)
	}

	// Cancel deletion confirmation dialog so model returns to normal mode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(SessionSelectorModel)

	// Verify printable typing works
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(SessionSelectorModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(SessionSelectorModel)
	if m.FilterQuery != "an" {
		t.Fatalf("expected FilterQuery to be 'an', got: %q", m.FilterQuery)
	}
}
