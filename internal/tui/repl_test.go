package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReplInputModelSlashPopupFiltering(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")

	// 1. Initially SlashActive should be false
	if model.SlashActive {
		t.Errorf("expected SlashActive to be false initially")
	}

	// 2. Type '/'
	model.TextInput.SetValue("/")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := updated.(ReplInputModel)

	if !m.SlashActive {
		t.Errorf("expected SlashActive to be true after typing '/'")
	}
	if len(m.FilteredCommands) != len(defaultSlashCommands) {
		t.Errorf("expected all slash commands to be present for '/'")
	}

	// 3. Type '/re'
	m.TextInput.SetValue("/re")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(ReplInputModel)

	if !m.SlashActive {
		t.Errorf("expected SlashActive to be true for '/re'")
	}

	foundReset := false
	for _, cmd := range m.FilteredCommands {
		if cmd.Command == "/reset" {
			foundReset = true
		}
	}
	if !foundReset {
		t.Errorf("expected '/reset' to be in filtered commands for '/re'")
	}
}

func TestReplInputModelTabAutocompletion(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")

	// Set value to '/re' and activate popup
	model.TextInput.SetValue("/re")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m := updated.(ReplInputModel)

	// Press Tab to autocomplete highlighted command
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(ReplInputModel)

	if !strings.HasPrefix(m.TextInput.Value(), "/") {
		t.Errorf("expected text input value to start with slash after Tab autocomplete")
	}
}
