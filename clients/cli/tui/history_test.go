package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReplInputModelHistoryNavigation(t *testing.T) {
	history := []string{"investigate ANTM", "check IHSG foreign flow", "analyze ASII dividends"}
	model := NewReplInputModelWithHistory("niskava [hermes] >", history)

	// Verify initial history index at end of history
	if model.HistoryIndex != 3 {
		t.Errorf("expected initial HistoryIndex to be 3, got %d", model.HistoryIndex)
	}

	// Press Up Arrow (tea.KeyUp) -> should yield last item "analyze ASII dividends"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := updated.(ReplInputModel)
	if m.TextInput.Value() != "analyze ASII dividends" {
		t.Errorf("expected Up Arrow to populate 'analyze ASII dividends', got %q", m.TextInput.Value())
	}
	if m.HistoryIndex != 2 {
		t.Errorf("expected HistoryIndex to be 2, got %d", m.HistoryIndex)
	}

	// Press Up Arrow again -> should yield "check IHSG foreign flow"
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(ReplInputModel)
	if m.TextInput.Value() != "check IHSG foreign flow" {
		t.Errorf("expected Up Arrow to populate 'check IHSG foreign flow', got %q", m.TextInput.Value())
	}
	if m.HistoryIndex != 1 {
		t.Errorf("expected HistoryIndex to be 1, got %d", m.HistoryIndex)
	}

	// Press Down Arrow (tea.KeyDown) -> should yield "analyze ASII dividends"
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(ReplInputModel)
	if m.TextInput.Value() != "analyze ASII dividends" {
		t.Errorf("expected Down Arrow to populate 'analyze ASII dividends', got %q", m.TextInput.Value())
	}

	// Press Down Arrow again -> should restore draft (empty string)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(ReplInputModel)
	if m.TextInput.Value() != "" {
		t.Errorf("expected Down Arrow at end of history to restore draft buffer, got %q", m.TextInput.Value())
	}
}
