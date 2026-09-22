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
