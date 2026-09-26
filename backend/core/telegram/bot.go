package telegram

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"gopkg.in/telebot.v3"
)

// SessionManager defines session lifecycle controls needed by Telegram bot.
type SessionManager interface {
	Register(sessionID string, cancel context.CancelFunc) bool
	Unregister(sessionID string)
	Abort(sessionID string) bool
	IsBusy(sessionID string) bool
}

// DefaultSessionManager is an in-memory session manager implementation.
type DefaultSessionManager struct {
	mu     sync.Mutex
	active map[string]context.CancelFunc
}

// NewDefaultSessionManager creates a new DefaultSessionManager.
func NewDefaultSessionManager() *DefaultSessionManager {
	return &DefaultSessionManager{
		active: make(map[string]context.CancelFunc),
	}
}

// Register registers an active session with its cancel function.
func (sm *DefaultSessionManager) Register(sessionID string, cancel context.CancelFunc) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.active[sessionID]; exists {
		return false
	}
	sm.active[sessionID] = cancel
	return true
}

// Unregister removes a session from tracking.
func (sm *DefaultSessionManager) Unregister(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.active, sessionID)
}

// Abort triggers cancellation and unregisters the session.
func (sm *DefaultSessionManager) Abort(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if cancel, exists := sm.active[sessionID]; exists {
		cancel()
		delete(sm.active, sessionID)
		return true
	}
	return false
}

// IsBusy returns whether a session is currently executing.
func (sm *DefaultSessionManager) IsBusy(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, exists := sm.active[sessionID]
	return exists
}

// BotService orchestrates the Telegram Bot integration for Niskava Agent.
type BotService struct {
	bot         *telebot.Bot
	db          *db.DB
	sm          SessionManager
	cfg         *config.Config
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	started     bool
	mu          sync.Mutex
	rateLimiter *UserRateLimiter
}

// NewBotService validates Telegram settings and initializes the BotService.
func NewBotService(cfg *config.Config, database *db.DB, sm SessionManager) (*BotService, error) {
	if cfg == nil || strings.TrimSpace(cfg.Telegram.BotToken) == "" {
		return nil, fmt.Errorf("telegram bot_token cannot be empty")
	}

	if sm == nil {
		sm = NewDefaultSessionManager()
	}

	pref := telebot.Settings{
		Token:   cfg.Telegram.BotToken,
		Poller:  &telebot.LongPoller{Timeout: 10 * time.Second},
		Offline: cfg.Preferences.OfflineMode || os.Getenv("TELEGRAM_OFFLINE") == "1",
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BotService{
		bot:         bot,
		db:          database,
		sm:          sm,
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancel,
		rateLimiter: NewUserRateLimiter(6, time.Minute),
	}, nil
}

// Bot returns the underlying telebot.Bot instance.
func (s *BotService) Bot() *telebot.Bot {
	return s.bot
}

// SessionManager returns the session manager instance.
func (s *BotService) SessionManager() SessionManager {
	return s.sm
}

// IsStarted returns true if the bot service is running.
func (s *BotService) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// BotUsername returns the Telegram bot's username if connected, or empty string.
func (s *BotService) BotUsername() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bot != nil && s.bot.Me != nil {
		return s.bot.Me.Username
	}
	return ""
}

// Token returns the active Telegram bot token used by this instance.
func (s *BotService) Token() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bot != nil {
		return s.bot.Token
	}
	return ""
}

// Start registers all endpoint handlers and starts the bot in a managed background goroutine.
func (s *BotService) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	if s.ctx == nil || s.ctx.Err() != nil {
		s.ctx, s.cancel = context.WithCancel(context.Background())
	}
	s.started = true

	s.registerHandlers()

	_ = s.bot.SetCommands([]telebot.Command{
		{Text: "new", Description: "Mulai sesi riset baru"},
		{Text: "export", Description: "Unduh laporan riset (.md)"},
		{Text: "status", Description: "Status sesi aktif & model AI"},
		{Text: "stop", Description: "Batalkan analisis yang sedang berjalan"},
		{Text: "help", Description: "Panduan & batasan non-advisory"},
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.bot.Start()
	}()
	s.mu.Unlock()

	return nil
}

// Stop gracefully shuts down the Telegram bot and terminates pending routines.
func (s *BotService) Stop() {
	s.cancel()

	s.mu.Lock()
	if s.started {
		s.bot.Stop()
		s.started = false
	}
	s.mu.Unlock()

	s.wg.Wait()
}

// SendTestMessage sends a test notification to the specified chat ID.
func (s *BotService) SendTestMessage(chatID int64, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bot == nil {
		return fmt.Errorf("bot instance is not initialized")
	}
	_, err := s.bot.Send(&telebot.Chat{ID: chatID}, text)
	return err
}
