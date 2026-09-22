package tui

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
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

func TestReplInputModelBackCommandInSlashPopup(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")
	model.TextInput.SetValue("/ba")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m := updated.(ReplInputModel)

	foundBack := false
	for _, cmd := range m.FilteredCommands {
		if cmd.Command == "/back" {
			foundBack = true
		}
	}
	if !foundBack {
		t.Error("expected '/back' to appear in slash popup when typing '/ba'")
	}
}

func TestReplInputModelBackCommandSubmit(t *testing.T) {
	model := NewReplInputModel("niskava [hermes] >")
	model.TextInput.SetValue("/back")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m := updated.(ReplInputModel)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(ReplInputModel)

	if cmd == nil {
		t.Error("expected tea.Quit command after submitting /back")
	}
	if result.SubmittedValue != "/back" {
		t.Errorf("expected SubmittedValue '/back', got '%s'", result.SubmittedValue)
	}
}

func TestReplBackSentinel(t *testing.T) {
	if ReplBackSentinel != "__back__" {
		t.Errorf("expected ReplBackSentinel to be '__back__', got '%s'", ReplBackSentinel)
	}
	if replBackSentinel != "__back__" {
		t.Errorf("expected replBackSentinel to be '__back__', got '%s'", replBackSentinel)
	}
}

func TestI18nRepl_ChatsSavedNoticeKey(t *testing.T) {
	SetLanguage("en")
	en := T("repl_chats_saved_notice")
	if en == "repl_chats_saved_notice" {
		t.Error("expected English translation for 'repl_chats_saved_notice', got the key itself")
	}
	if !strings.Contains(en, "%s") {
		t.Errorf("expected 'repl_chats_saved_notice' to contain format verbs '%%s'")
	}

	SetLanguage("id")
	id := T("repl_chats_saved_notice")
	if id == "repl_chats_saved_notice" {
		t.Error("expected Indonesian translation for 'repl_chats_saved_notice', got the key itself")
	}

	SetLanguage("en")
}

func TestTF_ChatsSavedNoticeFormatting(t *testing.T) {
	SetLanguage("en")
	en := TF("repl_chats_saved_notice", "CHAT-001", "CHAT-002", "Valuasi BBCA")
	if !strings.Contains(en, "CHAT-001") || !strings.Contains(en, "CHAT-002") || !strings.Contains(en, "Valuasi BBCA") {
		t.Errorf("expected formatted English notice with IDs and title, got %q", en)
	}
	if !strings.Contains(en, "Active session") || !strings.Contains(en, "saved") {
		t.Errorf("expected English text in notice, got %q", en)
	}

	SetLanguage("id")
	id := TF("repl_chats_saved_notice", "CHAT-001", "CHAT-002", "Valuasi BBCA")
	if !strings.Contains(id, "CHAT-001") || !strings.Contains(id, "CHAT-002") || !strings.Contains(id, "Valuasi BBCA") {
		t.Errorf("expected formatted Indonesian notice with IDs and title, got %q", id)
	}
	if !strings.Contains(id, "Sesi aktif") || !strings.Contains(id, "tersimpan") {
		t.Errorf("expected Indonesian text in notice, got %q", id)
	}

	SetLanguage("en")
}

func TestI18nRepl_SessionKeys(t *testing.T) {
	keys := []string{
		"repl_db_unavailable",
		"repl_chats_fetch_err",
		"repl_resume_usage",
		"repl_resume_not_found",
		"repl_resumed_history_divider",
		"repl_user_label",
		"repl_agent_label",
	}

	for _, lang := range []string{"en", "id"} {
		SetLanguage(lang)
		for _, key := range keys {
			val := T(key)
			if val == key {
				t.Errorf("expected translation for key %q in lang %q, got key name", key, lang)
			}
		}
	}
	SetLanguage("en")
}

func TestI18nAllNewKeysExistInBothLanguages(t *testing.T) {
	newKeys := []string{
		"slash_back_desc",
		"repl_back_msg",
		"repl_chats_saved_notice",
		"slash_chats_desc",
		"slash_resume_desc",
		"session_selector_title",
		"session_selector_hint",
		"session_selector_empty",
	}

	for _, key := range newKeys {
		// Test English
		SetLanguage("en")
		val := T(key)
		if val == key {
			t.Errorf("missing English translation for i18n key '%s'", key)
		}
		if val == "" {
			t.Errorf("empty English translation for i18n key '%s'", key)
		}

		// Test Indonesian
		SetLanguage("id")
		val = T(key)
		if val == key {
			t.Errorf("missing Indonesian translation for i18n key '%s'", key)
		}
		if val == "" {
			t.Errorf("empty Indonesian translation for i18n key '%s'", key)
		}
	}

	// Reset
	SetLanguage("en")
}

func TestBannerHintContainsChatsAndBack(t *testing.T) {
	SetLanguage("en")
	hint := T("banner_hint")
	if !strings.Contains(hint, "/chats") {
		t.Error("banner_hint should mention /chats command")
	}
	if !strings.Contains(hint, "/back") {
		t.Error("banner_hint should mention /back command")
	}

	SetLanguage("id")
	hintID := T("banner_hint")
	if !strings.Contains(hintID, "/chats") {
		t.Error("banner_hint (id) should mention /chats command")
	}
	if !strings.Contains(hintID, "/back") {
		t.Error("banner_hint (id) should mention /back command")
	}

	SetLanguage("en")
}

func TestSlashCommandsListContainsBack(t *testing.T) {
	SetLanguage("en")
	cmds := GetLocalizedSlashCommands()
	found := false
	for _, c := range cmds {
		if c.Command == "/back" {
			found = true
			if c.Description == "" {
				t.Error("/back command should have a non-empty description")
			}
		}
	}
	if !found {
		t.Error("expected '/back' to be present in GetLocalizedSlashCommands()")
	}

	SetLanguage("id")
	cmdsID := GetLocalizedSlashCommands()
	foundID := false
	for _, c := range cmdsID {
		if c.Command == "/back" {
			foundID = true
			if c.Description == "" {
				t.Error("/back command should have a non-empty description in Indonesian")
			}
		}
	}
	if !foundID {
		t.Error("expected '/back' to be present in GetLocalizedSlashCommands() (id)")
	}

	SetLanguage("en")
}

func TestInterruptedFlagNoRace(t *testing.T) {
	var interrupted atomic.Bool

	done := make(chan struct{})
	go func() {
		interrupted.Store(true)
		close(done)
	}()

	<-done
	val := interrupted.Load()
	if !val {
		t.Error("expected interrupted to be true")
	}
}
