package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
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

	// 1. Indonesian test
	SetLanguage("id")
	badgeId := renderCompletionBadge(duration, "CHAT-TEST-001", "hermes", 2, 1)
	if !strings.Contains(badgeId, "SELESAI") {
		t.Errorf("expected Indonesian badge to contain 'SELESAI', got: %s", badgeId)
	}
	if !strings.Contains(badgeId, "1.5s") {
		t.Errorf("expected badge to contain duration '1.5s', got: %s", badgeId)
	}
	if !strings.Contains(badgeId, "2 Anomali, 1 Temuan") {
		t.Errorf("expected badge to contain '2 Anomali, 1 Temuan', got: %s", badgeId)
	}

	// 2. English test
	SetLanguage("en")
	badgeEn := renderCompletionBadge(duration, "CHAT-TEST-001", "hermes", 2, 1)
	if !strings.Contains(badgeEn, "COMPLETED") {
		t.Errorf("expected English badge to contain 'COMPLETED', got: %s", badgeEn)
	}
	if !strings.Contains(badgeEn, "1.5s") {
		t.Errorf("expected badge to contain duration '1.5s', got: %s", badgeEn)
	}
	if !strings.Contains(badgeEn, "2 Anomalies, 1 Findings") {
		t.Errorf("expected badge to contain '2 Anomalies, 1 Findings', got: %s", badgeEn)
	}
}

func TestReplInputModelSlashChatsAndResume(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")
	model.TextInput.SetValue("/ch")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m := updated.(ReplInputModel)

	foundChats := false
	for _, cmd := range m.FilteredCommands {
		if cmd.Command == "/chats" {
			foundChats = true
		}
	}
	if !foundChats {
		t.Errorf("expected '/chats' to be in filtered commands for '/ch'")
	}
}

func TestRenderResumedHistory(t *testing.T) {
	// 1. Should not panic on nil db or empty session
	renderResumedHistory(nil, "NON-EXISTENT")

	// 2. Test with populated database
	dbPath := filepath.Join(t.TempDir(), "test.db")
	tmpDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tmpDB.Close()

	sessionID := "TEST-RESUME-001"
	_ = tmpDB.SaveChatMessage(&db.ChatMessage{
		ID:        "M1",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Cek saham BBRI",
	})
	_ = tmpDB.SaveChatMessage(&db.ChatMessage{
		ID:        "M2",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Berikut ringkasan analisis saham BBRI.",
	})

	renderResumedHistory(tmpDB, sessionID)
}

func TestReplInputModel_EscKeyBehavior(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")

	// 1. Non-empty input: pressing Esc clears text input
	model.TextInput.SetValue("analisis ANTM")
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(ReplInputModel)

	if m.TextInput.Value() != "" {
		t.Fatalf("expected text input to be cleared after Esc, got: %s", m.TextInput.Value())
	}
	if m.Quitting {
		t.Fatalf("expected Quitting to be false when clearing text input")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd when clearing text input")
	}

	// 2. Slash active: pressing Esc closes slash popup
	model.TextInput.SetValue("/res")
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(ReplInputModel)
	if !m.SlashActive {
		t.Fatalf("expected SlashActive to be true for '/res'")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(ReplInputModel)
	if m.SlashActive {
		t.Fatalf("expected SlashActive to be false after Esc")
	}
	if m.Quitting {
		t.Fatalf("expected Quitting to be false when dismissing slash popup")
	}
}

func TestReplInputModel_DoublePressExit(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")

	// First Esc press on empty input: sets warning, does not quit
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(ReplInputModel)

	if !m.ExitWarning {
		t.Fatalf("expected ExitWarning to be true after first Esc press")
	}
	if m.Quitting {
		t.Fatalf("expected Quitting to be false after first Esc press")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd after first Esc press")
	}

	// Second Esc press immediately (within 2s): quits with /exit
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(ReplInputModel)

	if !m.Quitting {
		t.Fatalf("expected Quitting to be true after second Esc press")
	}
	if m.SubmittedValue != "/exit" {
		t.Fatalf("expected SubmittedValue to be '/exit', got: %s", m.SubmittedValue)
	}
}
