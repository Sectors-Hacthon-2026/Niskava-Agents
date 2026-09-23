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
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"gopkg.in/telebot.v3"
)

// BotService orchestrates the Telegram Bot integration for Niskava Agent.
type BotService struct {
	bot         *telebot.Bot
	db          *db.DB
	sm          *server.SessionManager
	cfg         *config.Config
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	started     bool
	mu          sync.Mutex
	rateLimiter *UserRateLimiter
}

// NewBotService validates Telegram settings and initializes the BotService.
func NewBotService(cfg *config.Config, database *db.DB, sm *server.SessionManager) (*BotService, error) {
	if cfg == nil || strings.TrimSpace(cfg.Telegram.BotToken) == "" {
		return nil, fmt.Errorf("telegram bot_token cannot be empty")
	}

	if sm == nil {
		sm = server.NewSessionManager()
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
func (s *BotService) SessionManager() *server.SessionManager {
	return s.sm
}

// Start registers all endpoint handlers and starts the bot in a managed background goroutine.
func (s *BotService) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
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
