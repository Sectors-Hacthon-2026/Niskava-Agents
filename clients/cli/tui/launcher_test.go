package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLauncherModelDirectHotkeys(t *testing.T) {
	model := NewLauncherModelWithHealth("http://localhost:8080", "v1.0.0", true)

	// Test direct key 'w' -> Web UI
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd == nil {
		t.Fatalf("expected tea.Quit command on direct hotkey selection")
	}
	m := updated.(LauncherModel)
	if m.Selected != "web" {
		t.Errorf("expected selected action 'web', got '%s'", m.Selected)
	}

	// Test direct key 't' -> Terminal UI
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(LauncherModel)
	if m.Selected != "terminal" {
		t.Errorf("expected selected action 'terminal', got '%s'", m.Selected)
	}

	// Test direct key 's' -> Sessions
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(LauncherModel)
	if m.Selected != "sessions" {
		t.Errorf("expected selected action 'sessions', got '%s'", m.Selected)
	}

	// Test direct key 'h' -> Help
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(LauncherModel)
	if m.Selected != "help" {
		t.Errorf("expected selected action 'help', got '%s'", m.Selected)
	}

	// Test direct key 'c' -> Health
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(LauncherModel)
	if m.Selected != "health" {
		t.Errorf("expected selected action 'health', got '%s'", m.Selected)
	}

	// Test direct key 'q' -> Setup
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(LauncherModel)
	if m.Selected != "setup" {
		t.Errorf("expected selected action 'setup', got '%s'", m.Selected)
	}
}

func TestLauncherModelViewOutput(t *testing.T) {
	model := NewLauncherModelWithHealth("http://localhost:8080", "v1.0.0", true)
	viewStr := model.View()

	if !strings.Contains(viewStr, "Multi-Interface AI Agent Runtime") {
		t.Errorf("expected view output to contain tagline 'Multi-Interface AI Agent Runtime'")
	}
	if !strings.Contains(viewStr, "[W]") || !strings.Contains(viewStr, "[T]") || !strings.Contains(viewStr, "[E]") {
		t.Errorf("expected view output to contain shortcut keys [W], [T], [E]")
	}
	if !strings.Contains(viewStr, "Server: localhost:8080") {
		t.Errorf("expected view output to contain server host in status bar")
	}
}

func TestLauncherModelNoColorFallback(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	model := NewLauncherModelWithHealth("http://localhost:8080", "v1.0.0", false)
	viewStr := model.View()

	if !strings.Contains(viewStr, "niskava v1.0.0") {
		t.Errorf("expected plain text status bar in NO_COLOR mode")
	}
	if !strings.Contains(viewStr, "[W]") {
		t.Errorf("expected shortcut indicator [W] in NO_COLOR mode")
	}
}
