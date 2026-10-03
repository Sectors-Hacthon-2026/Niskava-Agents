package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSessionGatewayModel_NavigationAndSelect(t *testing.T) {
	model := NewSessionGatewayModel()

	if model.Cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", model.Cursor)
	}

	// Move down to option 2 (investigations)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(SessionGatewayModel)
	if m.Cursor != 1 {
		t.Fatalf("expected cursor at 1 after KeyDown, got %d", m.Cursor)
	}

	// Press Enter to select investigations
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SessionGatewayModel)
	if m.Selected != "investigations" {
		t.Fatalf("expected selected to be 'investigations', got %s", m.Selected)
	}
}

func TestSessionGatewayModel_DirectHotkeys(t *testing.T) {
	// Key '1' selects chats
	model1 := NewSessionGatewayModel()
	updated1, _ := model1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m1 := updated1.(SessionGatewayModel)
	if m1.Selected != "chats" {
		t.Fatalf("expected '1' to select 'chats', got %s", m1.Selected)
	}

	// Key '2' selects investigations
	model2 := NewSessionGatewayModel()
	updated2, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m2 := updated2.(SessionGatewayModel)
	if m2.Selected != "investigations" {
		t.Fatalf("expected '2' to select 'investigations', got %s", m2.Selected)
	}
}

func TestSessionGatewayModel_Cancel(t *testing.T) {
	model := NewSessionGatewayModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(SessionGatewayModel)
	if !m.Canceled {
		t.Fatalf("expected Canceled to be true upon Esc")
	}
	if m.Selected != "" {
		t.Fatalf("expected empty Selected upon Esc, got %s", m.Selected)
	}
}
