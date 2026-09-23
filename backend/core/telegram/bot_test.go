package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"gopkg.in/telebot.v3"
)

func TestNewBotServiceValidation(t *testing.T) {
	// 1. Nil config
	_, err := NewBotService(nil, nil, nil)
	if err == nil {
		t.Errorf("expected error for nil config, got nil")
	}

	// 2. Empty bot token
	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = ""
	_, err = NewBotService(cfg, nil, nil)
	if err == nil {
		t.Errorf("expected error for empty bot_token, got nil")
	}

	// 3. Whitespace-only bot token
	cfg.Telegram.BotToken = "   "
	_, err = NewBotService(cfg, nil, nil)
	if err == nil {
		t.Errorf("expected error for whitespace-only bot_token, got nil")
	}
}

func TestNewBotServiceSuccessAndLifecycle(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	cfg.Preferences.OfflineMode = true

	tempDir := t.TempDir()
	database, err := db.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	sm := server.NewSessionManager()

	service, err := NewBotService(cfg, database, sm)
	if err != nil {
		t.Fatalf("failed to create bot service: %v", err)
	}

	if service.Bot() == nil {
		t.Errorf("expected non-nil telebot.Bot")
	}
	if service.SessionManager() == nil {
		t.Errorf("expected non-nil SessionManager")
	}

	// Start bot service
	if err := service.Start(); err != nil {
		t.Fatalf("failed to start bot service: %v", err)
	}

	// Starting again should be a no-op
	if err := service.Start(); err != nil {
		t.Errorf("expected second start to be no-op, got %v", err)
	}

	// Stop bot service
	service.Stop()
}

func TestIsAuthorized(t *testing.T) {
	cfg := config.DefaultConfig()
	service := &BotService{cfg: cfg}

	// 1. Open access when AllowedUsers is empty
	cfg.Telegram.AllowedUsers = nil
	if !service.isAuthorized(&telebot.User{ID: 100, Username: "anyone"}) {
		t.Errorf("expected open access when AllowedUsers is empty")
	}

	// 2. Specific allowed users configured
	cfg.Telegram.AllowedUsers = []string{"analyst_idx", "@trader_jkt", "998877"}

	// Allowed by username without @
	if !service.isAuthorized(&telebot.User{ID: 1, Username: "analyst_idx"}) {
		t.Errorf("expected analyst_idx to be authorized")
	}

	// Allowed by username case-insensitively
	if !service.isAuthorized(&telebot.User{ID: 2, Username: "Analyst_IDX"}) {
		t.Errorf("expected Analyst_IDX to be authorized case-insensitively")
	}

	// Allowed by @username in config
	if !service.isAuthorized(&telebot.User{ID: 3, Username: "trader_jkt"}) {
		t.Errorf("expected trader_jkt to be authorized")
	}
	if !service.isAuthorized(&telebot.User{ID: 4, Username: "@trader_jkt"}) {
		t.Errorf("expected @trader_jkt to be authorized")
	}

	// Allowed by numeric user ID
	if !service.isAuthorized(&telebot.User{ID: 998877, Username: "different_name"}) {
		t.Errorf("expected user ID 998877 to be authorized")
	}

	// Denied for unknown user
	if service.isAuthorized(&telebot.User{ID: 55555, Username: "stranger"}) {
		t.Errorf("expected stranger to be denied access")
	}

	// Denied for nil sender
	if service.isAuthorized(nil) {
		t.Errorf("expected nil sender to be denied access")
	}
}

func TestCommandHandlersWithMockServer(t *testing.T) {
	// Setup mock Telegram API server to handle send/reply calls
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 101,
				"date":       time.Now().Unix(),
				"chat": map[string]interface{}{
					"id":   12345,
					"type": "private",
				},
				"text": "mock response",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "mock_token"
	cfg.Preferences.OfflineMode = true

	tempDir := t.TempDir()
	database, err := db.Open(filepath.Join(tempDir, "test_handlers.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	sm := server.NewSessionManager()

	pref := telebot.Settings{
		URL:     mockServer.URL,
		Token:   cfg.Telegram.BotToken,
		Offline: true,
		Poller:  &telebot.LongPoller{Timeout: 1 * time.Second},
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		t.Fatalf("failed to create telebot with mock server: %v", err)
	}

	service := &BotService{
		bot: bot,
		db:  database,
		sm:  sm,
		cfg: cfg,
	}

	user := &telebot.User{
		ID:       12345,
		Username: "investor_test",
	}
	chat := &telebot.Chat{
		ID:   12345,
		Type: telebot.ChatPrivate,
	}
	msg := &telebot.Message{
		ID:     1,
		Sender: user,
		Chat:   chat,
		Text:   "/start",
	}

	ctx := bot.NewContext(telebot.Update{
		ID:      1,
		Message: msg,
	})

	// 1. Test /start
	if err := service.handleStart(ctx); err != nil {
		t.Errorf("handleStart failed: %v", err)
	}

	// 2. Test /new
	msg.Text = "/new"
	if err := service.handleNewSession(ctx); err != nil {
		t.Errorf("handleNewSession failed: %v", err)
	}

	// Verify session was created in DB
	chatRecord, err := database.GetTelegramChat(chat.ID)
	if err != nil || chatRecord == nil {
		t.Fatalf("expected chat record in db, got %v", err)
	}

	// 3. Test /status
	msg.Text = "/status"
	if err := service.handleStatus(ctx); err != nil {
		t.Errorf("handleStatus failed: %v", err)
	}

	// 4. Test /stop when not running
	msg.Text = "/stop"
	if err := service.handleStop(ctx); err != nil {
		t.Errorf("handleStop failed: %v", err)
	}

	// 5. Test /stop when running
	cancelCalled := false
	sm.Register(chatRecord.CurrentSessionID, func() {
		cancelCalled = true
	})
	if !sm.IsBusy(chatRecord.CurrentSessionID) {
		t.Errorf("expected session to be busy")
	}
	if err := service.handleStop(ctx); err != nil {
		t.Errorf("handleStop with active session failed: %v", err)
	}
	if !cancelCalled {
		t.Errorf("expected abort cancel func to be triggered")
	}
	if sm.IsBusy(chatRecord.CurrentSessionID) {
		t.Errorf("expected session to no longer be busy after stop")
	}

	// 6. Test handleTextMessage when session is busy
	sm.Register(chatRecord.CurrentSessionID, func() {})
	msg.Text = "Analisis saham BBCA"
	if err := service.handleTextMessage(ctx); err != nil {
		t.Errorf("handleTextMessage when busy failed: %v", err)
	}
	sm.Unregister(chatRecord.CurrentSessionID)
}

func TestAuthMiddleware(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 102,
				"chat": map[string]interface{}{
					"id": 555,
				},
				"text": "blocked",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "mock_token"
	cfg.Telegram.AllowedUsers = []string{"authorized_user"}
	cfg.Preferences.OfflineMode = true

	pref := telebot.Settings{
		URL:     mockServer.URL,
		Token:   cfg.Telegram.BotToken,
		Offline: true,
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}

	service := &BotService{
		bot: bot,
		cfg: cfg,
	}

	middleware := service.authMiddleware()
	handlerExecuted := false
	dummyHandler := func(c telebot.Context) error {
		handlerExecuted = true
		return nil
	}

	// Test unauthorized sender
	unauthorizedCtx := bot.NewContext(telebot.Update{
		Message: &telebot.Message{
			Chat:   &telebot.Chat{ID: 555},
			Sender: &telebot.User{ID: 999, Username: "hacker"},
		},
	})
	wrapped := middleware(dummyHandler)
	if err := wrapped(unauthorizedCtx); err != nil {
		t.Errorf("expected nil error on unauthorized reply, got %v", err)
	}
	if handlerExecuted {
		t.Errorf("expected handler NOT to execute for unauthorized user")
	}

	// Test authorized sender
	handlerExecuted = false
	authorizedCtx := bot.NewContext(telebot.Update{
		Message: &telebot.Message{
			Chat:   &telebot.Chat{ID: 555},
			Sender: &telebot.User{ID: 100, Username: "authorized_user"},
		},
	})
	if err := wrapped(authorizedCtx); err != nil {
		t.Errorf("expected nil error for authorized user, got %v", err)
	}
	if !handlerExecuted {
		t.Errorf("expected handler to execute for authorized user")
	}
}

func TestHandleExport(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 201,
				"date":       time.Now().Unix(),
				"chat": map[string]interface{}{
					"id":   5555,
					"type": "private",
				},
				"text": "export response",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "mock_token"
	cfg.Preferences.OfflineMode = true

	tempDir := t.TempDir()
	database, err := db.Open(filepath.Join(tempDir, "test_export.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	pref := telebot.Settings{
		URL:     mockServer.URL,
		Token:   cfg.Telegram.BotToken,
		Offline: true,
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		t.Fatalf("failed to create telebot: %v", err)
	}

	service := &BotService{
		bot: bot,
		db:  database,
		cfg: cfg,
	}

	user := &telebot.User{
		ID:       5555,
		Username: "investor_export",
	}
	chat := &telebot.Chat{
		ID:   5555,
		Type: telebot.ChatPrivate,
	}
	msg := &telebot.Message{
		ID:     1,
		Sender: user,
		Chat:   chat,
		Text:   "/export",
	}
	ctx := bot.NewContext(telebot.Update{
		ID:      1,
		Message: msg,
	})

	// 1. Export on brand new session with no history
	if err := service.handleExport(ctx); err != nil {
		t.Errorf("handleExport on empty session failed: %v", err)
	}

	// 2. Add chat messages to session
	sessionID, err := database.GetOrCreateTelegramChatSession(chat.ID, user.ID, user.Username)
	if err != nil {
		t.Fatalf("failed to get/create session: %v", err)
	}

	userMsg := &db.ChatMessage{
		ID:        "msg-1",
		SessionID: sessionID,
		Role:      "user",
		Content:   "Analisis anomali volume saham ANTM",
		Status:    "COMPLETED",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	thought := "Mengecek Z-Score volume ANTM selama 30 hari..."
	assistantMsg := &db.ChatMessage{
		ID:        "msg-2",
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "Ditemukan lonjakan volume dengan Z-Score 3.2 pada saham ANTM.",
		Thought:   &thought,
		Status:    "COMPLETED",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = database.SaveChatMessage(userMsg)
	_ = database.SaveChatMessage(assistantMsg)

	// 3. Export on session with history (sends document attachment)
	if err := service.handleExport(ctx); err != nil {
		t.Errorf("handleExport with messages failed: %v", err)
	}

	// Verify formatExportDocument directly
	history, _ := database.GetChatHistory(sessionID, 100)
	doc := formatExportDocument(sessionID, "gemini-2.5-flash", history)
	if !strings.Contains(doc, "# Laporan Riset Pasar Niskava") {
		t.Errorf("expected doc to contain header")
	}
	if !strings.Contains(doc, "Pemberitahuan Kepatuhan (Law 2") {
		t.Errorf("expected doc to contain Law 2 disclaimer")
	}
	if !strings.Contains(doc, "Analisis anomali volume saham ANTM") {
		t.Errorf("expected doc to contain user message")
	}
	if !strings.Contains(doc, "Ditemukan lonjakan volume") {
		t.Errorf("expected doc to contain assistant message")
	}
	if !strings.Contains(doc, "Mengecek Z-Score volume ANTM") {
		t.Errorf("expected doc to contain assistant thought process")
	}
}

func TestRateLimiterRejectionInHandleTextMessage(t *testing.T) {
	var lastRequestBody []byte
	var mu sync.Mutex

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		lastRequestBody = body
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 301,
				"date":       time.Now().Unix(),
				"chat": map[string]interface{}{
					"id":   7777,
					"type": "private",
				},
				"text": "rate limit response",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "mock_token"
	cfg.Preferences.OfflineMode = true

	tempDir := t.TempDir()
	database, err := db.Open(filepath.Join(tempDir, "test_ratelimit.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	pref := telebot.Settings{
		URL:     mockServer.URL,
		Token:   cfg.Telegram.BotToken,
		Offline: true,
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		t.Fatalf("failed to create telebot: %v", err)
	}

	limiter := NewUserRateLimiter(6, time.Minute)
	service := &BotService{
		bot:         bot,
		db:          database,
		cfg:         cfg,
		rateLimiter: limiter,
	}

	user := &telebot.User{
		ID:       7777,
		Username: "fast_user",
	}
	chat := &telebot.Chat{
		ID:   7777,
		Type: telebot.ChatPrivate,
	}
	msg := &telebot.Message{
		ID:     1,
		Sender: user,
		Chat:   chat,
		Text:   "Pertanyaan ke-7",
	}
	ctx := bot.NewContext(telebot.Update{
		ID:      1,
		Message: msg,
	})

	// Fire 6 requests through the limiter
	for i := 1; i <= 6; i++ {
		allowed, _ := limiter.Allow(user.ID)
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i)
		}
	}

	// 7th request must be rejected by limiter
	allowed, retryAfter := limiter.Allow(user.ID)
	if allowed {
		t.Fatalf("expected 7th request to be disallowed")
	}
	if retryAfter <= 0 {
		t.Fatalf("expected retryAfter > 0, got %v", retryAfter)
	}

	// Now execute handleTextMessage when quota is exhausted
	if err := service.handleTextMessage(ctx); err != nil {
		t.Errorf("expected handleTextMessage to return nil (handled rejection), got %v", err)
	}

	mu.Lock()
	bodyStr := string(lastRequestBody)
	mu.Unlock()

	if !strings.Contains(bodyStr, "Terlalu banyak permintaan") {
		t.Errorf("expected rejection warning in telegram reply, got body: %s", bodyStr)
	}
}

func TestCallbackHandlers(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 401,
				"date":       time.Now().Unix(),
				"chat": map[string]interface{}{
					"id":   8888,
					"type": "private",
				},
				"text": "callback response",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	cfg := config.DefaultConfig()
	cfg.Telegram.BotToken = "mock_token"
	cfg.Preferences.OfflineMode = true

	tempDir := t.TempDir()
	database, err := db.Open(filepath.Join(tempDir, "test_callbacks.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer database.Close()

	pref := telebot.Settings{
		URL:     mockServer.URL,
		Token:   cfg.Telegram.BotToken,
		Offline: true,
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		t.Fatalf("failed to create telebot: %v", err)
	}

	service := &BotService{
		bot: bot,
		db:  database,
		cfg: cfg,
	}

	user := &telebot.User{
		ID:       8888,
		Username: "callback_tester",
	}
	chat := &telebot.Chat{
		ID:   8888,
		Type: telebot.ChatPrivate,
	}
	parentMsg := &telebot.Message{
		ID:   1,
		Chat: chat,
	}

	cbNewCtx := bot.NewContext(telebot.Update{
		Callback: &telebot.Callback{
			ID:      "cb-1",
			Sender:  user,
			Message: parentMsg,
			Data:    "btn_new_session",
		},
	})
	if err := service.handleCallbackNew(cbNewCtx); err != nil {
		t.Errorf("handleCallbackNew failed: %v", err)
	}

	cbStatusCtx := bot.NewContext(telebot.Update{
		Callback: &telebot.Callback{
			ID:      "cb-2",
			Sender:  user,
			Message: parentMsg,
			Data:    "btn_status_session",
		},
	})
	if err := service.handleCallbackStatus(cbStatusCtx); err != nil {
		t.Errorf("handleCallbackStatus failed: %v", err)
	}

	cbExportCtx := bot.NewContext(telebot.Update{
		Callback: &telebot.Callback{
			ID:      "cb-3",
			Sender:  user,
			Message: parentMsg,
			Data:    "btn_export_session",
		},
	})
	if err := service.handleCallbackExport(cbExportCtx); err != nil {
		t.Errorf("handleCallbackExport failed: %v", err)
	}
}
