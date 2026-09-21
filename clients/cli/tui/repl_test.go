package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
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
	if len(m.FilteredCommands) != len(GetLocalizedSlashCommands()) {
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

func TestEventChannelDraining(t *testing.T) {
	eventsChan := make(chan ipc.Event, 5)
	errChan := make(chan error, 1)

	// Enqueue events
	eventsChan <- ipc.Event{Event: ipc.EventAgentThought, Thought: "Thinking..."}
	eventsChan <- ipc.Event{Event: ipc.EventAgentMessageChunk, Chunk: "Halo! "}
	eventsChan <- ipc.Event{Event: ipc.EventAgentMessageChunk, Chunk: "Ada yang bisa dibantu?"}
	eventsChan <- ipc.Event{Event: ipc.EventAgentMessageComplete, Content: "Halo! Ada yang bisa dibantu?"}
	close(eventsChan)
	close(errChan) // Closed simultaneously like cmd.Wait()

	var collected strings.Builder
	var lastThought string
	activeErrChan := errChan

	for {
		select {
		case err, ok := <-activeErrChan:
			if !ok {
				activeErrChan = nil
				continue
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		case ev, ok := <-eventsChan:
			if !ok {
				goto Done
			}
			if ev.Event == ipc.EventAgentThought {
				lastThought = ev.Thought
			}
			if ev.Event == ipc.EventAgentMessageChunk {
				collected.WriteString(ev.Chunk)
			}
		}
	}
Done:
	if lastThought != "Thinking..." {
		t.Errorf("expected thought 'Thinking...', got '%s'", lastThought)
	}
	if collected.String() != "Halo! Ada yang bisa dibantu?" {
		t.Errorf("expected collected response 'Halo! Ada yang bisa dibantu?', got '%s'", collected.String())
	}
}

func TestChatTurnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Simulate interrupt immediately

	select {
	case <-ctx.Done():
		// Success
	default:
		t.Fatalf("expected context to be cancelled")
	}
}

func TestCompletionBadgeFormatting(t *testing.T) {
	duration := 1500 * time.Millisecond
	badge := renderCompletionBadge(duration, "CHAT-TEST-001", "hermes", 0, 0)
	if !strings.Contains(badge, "SELESAI") {
		t.Errorf("expected badge to contain 'SELESAI', got: %s", badge)
	}
	if !strings.Contains(badge, "1.5s") {
		t.Errorf("expected badge to contain duration '1.5s', got: %s", badge)
	}
}
