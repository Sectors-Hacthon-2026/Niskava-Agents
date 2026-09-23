package telegram

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"gopkg.in/telebot.v3"
)

var (
	btnNew    = telebot.InlineButton{Unique: "btn_new_session", Text: "🔄 Sesi Baru"}
	btnExport = telebot.InlineButton{Unique: "btn_export_session", Text: "📑 Export Dokumen"}
	btnStatus = telebot.InlineButton{Unique: "btn_status_session", Text: "ℹ️ Status"}
)

// inlineActionsMarkup creates an inline keyboard markup with common quick action buttons.
func inlineActionsMarkup() *telebot.ReplyMarkup {
	return &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{btnNew, btnExport, btnStatus},
		},
	}
}

// registerHandlers registers all commands, middleware, and message listeners on the bot.
func (s *BotService) registerHandlers() {
	// Global authentication middleware
	s.bot.Use(s.authMiddleware())

	// Basic informational & control commands
	s.bot.Handle("/start", s.handleStart)
	s.bot.Handle("/help", s.handleStart)
	s.bot.Handle("/new", s.handleNewSession)
	s.bot.Handle("/reset", s.handleNewSession)
	s.bot.Handle("/stop", s.handleStop)
	s.bot.Handle("/status", s.handleStatus)
	s.bot.Handle("/export", s.handleExport)

	// Callback handlers for inline keyboard actions
	s.bot.Handle(&btnNew, s.handleCallbackNew)
	s.bot.Handle(&btnExport, s.handleCallbackExport)
	s.bot.Handle(&btnStatus, s.handleCallbackStatus)

	// Freeform conversation & investigation queries
	s.bot.Handle(telebot.OnText, s.handleTextMessage)
}

// isAuthorized checks if the given Telegram user is permitted to interact with the bot.
func (s *BotService) isAuthorized(sender *telebot.User) bool {
	if s.cfg == nil || len(s.cfg.Telegram.AllowedUsers) == 0 {
		return true
	}
	if sender == nil {
		return false
	}

	senderIDStr := strconv.FormatInt(sender.ID, 10)
	senderUsername := strings.ToLower(strings.TrimPrefix(sender.Username, "@"))

	for _, allowed := range s.cfg.Telegram.AllowedUsers {
		cleanAllowed := strings.TrimSpace(allowed)
		if cleanAllowed == "" {
			continue
		}
		// Match numerical User ID
		if cleanAllowed == senderIDStr {
			return true
		}
		// Match @username
		cleanAllowedUsername := strings.ToLower(strings.TrimPrefix(cleanAllowed, "@"))
		if senderUsername != "" && senderUsername == cleanAllowedUsername {
			return true
		}
	}

	return false
}

// authMiddleware enforces user access restrictions when AllowedUsers is configured.
func (s *BotService) authMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if !s.isAuthorized(c.Sender()) {
				return c.Reply("⛔ Akses ditolak. Akun Telegram Anda belum terdaftar dalam daftar pengguna terotorisasi Niskava Agent.")
			}
			return next(c)
		}
	}
}

// handleStart responds to /start and /help commands with welcome instructions.
func (s *BotService) handleStart(c telebot.Context) error {
	welcomeMsg := `🔍 *Niskava Agent — IDX Autonomous Financial Intelligence*

Selamat datang! Niskava Agent adalah platform intelijen dan OSINT pasar modal otonom khusus Bursa Efek Indonesia (IDX). Sistem menjembatani fakta kuantitatif Sectors Financial API v2 dengan bukti keterbukaan informasi dan berita emiten.

*Perintah Tersedia:*
• /start atau /help — Menampilkan panduan dan bantuan penggunaan bot
• /new atau /reset — Memulai sesi investigasi/percakapan baru
• /export — Unduh laporan riset (.md)
• /status — Menampilkan status sesi, kesibukan agent, dan konfigurasi model
• /stop — Membatalkan proses investigasi yang sedang berjalan

Silakan ketik pertanyaan atau perintah investigasi saham IDX (contoh: _"Analisis anomali volume saham ANTM dalam 30 hari terakhir"_).`

	formatted := FormatFinalResponse(welcomeMsg, "")
	return s.sendMarkdownOrPlain(c, formatted)
}

// handleNewSession creates a brand-new chat session for the current Telegram chat.
func (s *BotService) handleNewSession(c telebot.Context) error {
	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database tidak tersedia.")
	}

	sender := c.Sender()
	var userID int64
	var username string
	if sender != nil {
		userID = sender.ID
		username = sender.Username
	}

	newSessionID, err := s.db.ResetTelegramChatSession(c.Chat().ID, userID, username)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Gagal membuat sesi baru: %v", err))
	}

	msg := fmt.Sprintf("✨ *Sesi baru berhasil dibuat!*\nSession ID: `%s`\n\nRiwayat percakapan sebelumnya telah diarsipkan. Silakan ajukan pertanyaan atau emiten yang ingin dianalisis.", newSessionID)
	return s.sendMarkdownOrPlain(c, FormatFinalResponse(msg, ""))
}

// handleStop aborts any active running query in the current chat session.
func (s *BotService) handleStop(c telebot.Context) error {
	if s.db == nil || s.sm == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Layanan internal tidak tersedia.")
	}

	chat, err := s.db.GetTelegramChat(c.Chat().ID)
	if err != nil || chat == nil || chat.CurrentSessionID == "" {
		return s.sendMarkdownOrPlain(c, "ℹ️ Tidak ada sesi aktif yang ditemukan.")
	}

	aborted := s.sm.Abort(chat.CurrentSessionID)
	if aborted {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("🛑 Eksekusi investigasi untuk sesi `%s` berhasil dibatalkan.", chat.CurrentSessionID))
	}

	return s.sendMarkdownOrPlain(c, fmt.Sprintf("ℹ️ Tidak ada proses investigasi aktif yang sedang berjalan pada sesi `%s`.", chat.CurrentSessionID))
}

// handleStatus returns diagnostic status information about the active session and configuration.
func (s *BotService) handleStatus(c telebot.Context) error {
	sessionID := "Belum dibuat"
	isBusy := false

	if s.db != nil {
		chat, err := s.db.GetTelegramChat(c.Chat().ID)
		if err == nil && chat != nil && chat.CurrentSessionID != "" {
			sessionID = chat.CurrentSessionID
			if s.sm != nil {
				isBusy = s.sm.IsBusy(chat.CurrentSessionID)
			}
		}
	}

	busyText := "🟢 Idle (Siap menerima query)"
	if isBusy {
		busyText = "🟡 Memproses Analisis (Sedang berjalan)"
	}

	modelInfo := "Default"
	if s.cfg != nil {
		if s.cfg.Auth.AIProvider != "" {
			modelInfo = s.cfg.Auth.AIProvider
			if s.cfg.Auth.AIProvider == "gemini" && s.cfg.Auth.GeminiModel != "" {
				modelInfo = fmt.Sprintf("Gemini (%s)", s.cfg.Auth.GeminiModel)
			} else if s.cfg.Auth.AIProvider == "openai" && s.cfg.Auth.OpenAIModel != "" {
				modelInfo = fmt.Sprintf("OpenAI (%s)", s.cfg.Auth.OpenAIModel)
			} else if s.cfg.Auth.AIProvider == "ollama" && s.cfg.Auth.OllamaModel != "" {
				modelInfo = fmt.Sprintf("Ollama (%s)", s.cfg.Auth.OllamaModel)
			}
		}
	}

	market := "IDX"
	offlineMode := false
	if s.cfg != nil {
		if s.cfg.Preferences.DefaultMarket != "" {
			market = s.cfg.Preferences.DefaultMarket
		}
		offlineMode = s.cfg.Preferences.OfflineMode
	}

	statusMsg := fmt.Sprintf(`📊 *Status Niskava Agent*

• *Session ID:* `+"`%s`"+`
• *Status Aktivitas:* %s
• *Model AI:* %s
• *Pasar:* %s
• *Mode Offline:* %t`,
		sessionID,
		busyText,
		modelInfo,
		market,
		offlineMode,
	)

	return s.sendMarkdownOrPlain(c, FormatFinalResponse(statusMsg, ""))
}

func (s *BotService) handleCallbackNew(c telebot.Context) error {
	_ = c.Respond()
	return s.handleNewSession(c)
}

func (s *BotService) handleCallbackExport(c telebot.Context) error {
	_ = c.Respond()
	return s.handleExport(c)
}

func (s *BotService) handleCallbackStatus(c telebot.Context) error {
	_ = c.Respond()
	return s.handleStatus(c)
}

// formatExportDocument formats the session chat history into a structured Markdown research report.
func formatExportDocument(sessionID, model string, history []db.ChatMessage) string {
	var sb strings.Builder
	sb.WriteString("# Laporan Riset Pasar Niskava\n\n")
	sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
	sb.WriteString(fmt.Sprintf("- **Model AI:** %s\n", model))
	sb.WriteString(fmt.Sprintf("- **Waktu Ekspor:** %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	sb.WriteString("> **Pemberitahuan Kepatuhan (Law 2 — Non-Advisory Boundary):**\n")
	sb.WriteString("> Niskava Agent adalah platform intelijen dan OSINT pasar modal otonom IDX, BUKAN penasihat investasi atau broker terdaftar. Seluruh data, analisis anomali, dan korelasi bukti disajikan secara independen semata-mata untuk verifikasi fakta dan riset pasar modal. Tidak ada bagian dari laporan ini yang merupakan rekomendasi beli/jual atau nasihat investasi keuangan berlisensi.\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Riwayat Percakapan & Investigasi\n\n")

	for i, msg := range history {
		roleTitle := "👤 User"
		if strings.EqualFold(msg.Role, "assistant") {
			roleTitle = "🤖 Niskava Agent"
		} else if strings.EqualFold(msg.Role, "system") {
			roleTitle = "⚙️ Sistem"
		} else if strings.EqualFold(msg.Role, "tool") {
			roleTitle = "🔧 Pemanggilan Tool"
		}

		sb.WriteString(fmt.Sprintf("### %d. %s (`%s`)\n\n", i+1, roleTitle, msg.CreatedAt))

		if msg.Thought != nil && strings.TrimSpace(*msg.Thought) != "" {
			sb.WriteString("<details>\n<summary>Proses Penalaran (Chain of Thought)</summary>\n\n")
			sb.WriteString(strings.TrimSpace(*msg.Thought))
			sb.WriteString("\n\n</details>\n\n")
		}

		sb.WriteString(strings.TrimSpace(msg.Content))
		sb.WriteString("\n\n---\n\n")
	}

	sb.WriteString("*Dokumen ini digenerate secara otomatis oleh Niskava Agent.*\n")
	return sb.String()
}

// handleExport generates and sends a markdown document containing the current session's chat history.
func (s *BotService) handleExport(c telebot.Context) error {
	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database tidak tersedia.")
	}

	chat := c.Chat()
	if chat == nil {
		return fmt.Errorf("no chat found in context")
	}

	sender := c.Sender()
	var userID int64
	var username string
	if sender != nil {
		userID = sender.ID
		username = sender.Username
	}

	sessionID, err := s.db.GetOrCreateTelegramChatSession(chat.ID, userID, username)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Gagal memuat sesi percakapan: %v", err))
	}

	session, err := s.db.GetChatSession(sessionID)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Gagal memuat detail sesi: %v", err))
	}

	history, err := s.db.GetChatHistory(sessionID, 500)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Gagal memuat riwayat percakapan: %v", err))
	}

	if len(history) == 0 {
		emptyMsg := "⚠️ Belum ada riwayat percakapan untuk diekspor pada sesi ini."
		if err := c.Reply(emptyMsg); err != nil {
			return c.Send(emptyMsg)
		}
		return nil
	}

	modelInfo := "Default"
	if session != nil && session.Model != "" {
		modelInfo = session.Model
	} else if s.cfg != nil && s.cfg.Auth.AIProvider != "" {
		modelInfo = s.cfg.Auth.AIProvider
		if s.cfg.Auth.AIProvider == "gemini" && s.cfg.Auth.GeminiModel != "" {
			modelInfo = fmt.Sprintf("Gemini (%s)", s.cfg.Auth.GeminiModel)
		} else if s.cfg.Auth.AIProvider == "openai" && s.cfg.Auth.OpenAIModel != "" {
			modelInfo = fmt.Sprintf("OpenAI (%s)", s.cfg.Auth.OpenAIModel)
		} else if s.cfg.Auth.AIProvider == "ollama" && s.cfg.Auth.OllamaModel != "" {
			modelInfo = fmt.Sprintf("Ollama (%s)", s.cfg.Auth.OllamaModel)
		}
	}

	docContent := formatExportDocument(sessionID, modelInfo, history)

	doc := &telebot.Document{
		File:     telebot.FromReader(strings.NewReader(docContent)),
		FileName: fmt.Sprintf("niskava-session-%s.md", sessionID),
		MIME:     "text/markdown",
		Caption:  fmt.Sprintf("📑 Laporan Riset Sesi %s", sessionID),
	}
	return c.Send(doc)
}

// handleTextMessage processes freeform textual questions, streaming reasoning and response via IPC.
func (s *BotService) handleTextMessage(c telebot.Context) error {
	prompt := strings.TrimSpace(c.Text())
	if prompt == "" {
		return nil
	}

	sender := c.Sender()
	if s.rateLimiter != nil && sender != nil {
		allowed, retryAfter := s.rateLimiter.Allow(sender.ID)
		if !allowed {
			msg := fmt.Sprintf("⚠️ Terlalu banyak permintaan. Silakan tunggu %d detik.", int(retryAfter.Seconds())+1)
			if err := c.Reply(msg); err != nil {
				return c.Send(msg)
			}
			return nil
		}
	}

	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database tidak tersedia.")
	}

	var userID int64
	var username string
	if sender != nil {
		userID = sender.ID
		username = sender.Username
	}

	sessionID, err := s.db.GetOrCreateTelegramChatSession(c.Chat().ID, userID, username)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Gagal menyiapkan sesi percakapan: %v", err))
	}

	if s.sm != nil && s.sm.IsBusy(sessionID) {
		return s.sendMarkdownOrPlain(c, "⚠️ Sesi sedang memproses query sebelumnya. Ketik /stop untuk membatalkan.")
	}

	childCtx, cancelChild := context.WithCancel(s.ctx)
	defer cancelChild()

	if s.sm != nil {
		if !s.sm.Register(sessionID, cancelChild) {
			return s.sendMarkdownOrPlain(c, "⚠️ Sesi sedang memproses query sebelumnya. Ketik /stop untuk membatalkan.")
		}
		defer s.sm.Unregister(sessionID)
	}

	// Persist incoming user question
	userMsg := &db.ChatMessage{
		ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
		SessionID: sessionID,
		Role:      "user",
		Content:   prompt,
		Status:    "COMPLETED",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = s.db.SaveChatMessage(userMsg)

	// Background typing indicator loop
	stopTyping := make(chan struct{})
	var typingWg sync.WaitGroup
	typingWg.Add(1)
	go func() {
		defer typingWg.Done()
		_ = c.Notify(telebot.Typing)
		ticker := time.NewTicker(3500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopTyping:
				return
			case <-childCtx.Done():
				return
			case <-ticker.C:
				_ = c.Notify(telebot.Typing)
			}
		}
	}()

	var stopOnce sync.Once
	stopTypingFunc := func() {
		stopOnce.Do(func() {
			close(stopTyping)
			typingWg.Wait()
		})
	}
	defer stopTypingFunc()

	// Prepare subprocess runner configuration
	wd, _ := os.Getwd()
	lang := "id"
	var offline bool
	var pythonBin, enginePath string
	if s.cfg != nil {
		if s.cfg.Preferences.Language != "" {
			lang = s.cfg.Preferences.Language
		}
		offline = s.cfg.Preferences.OfflineMode
		pythonBin = s.cfg.Engine.PythonBin
		enginePath = s.cfg.Engine.EnginePath
	}

	runnerParams := ipc.RunnerParams{
		PythonBin:  pythonBin,
		EnginePath: enginePath,
		WorkDir:    wd,
		DBPath:     s.db.Path,
		SessionID:  sessionID,
		Prompt:     prompt,
		Offline:    offline,
		Language:   lang,
	}

	eventsChan, errChan := ipc.RunConversationStream(childCtx, runnerParams)

	var contentBuilder strings.Builder
	var thoughtBuilder strings.Builder
	var lastError error
	wasAborted := false

	for eventsChan != nil || errChan != nil {
		select {
		case <-childCtx.Done():
			wasAborted = true
			eventsChan = nil
			errChan = nil
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			if err != nil {
				lastError = err
			}
		case ev, ok := <-eventsChan:
			if !ok {
				eventsChan = nil
				continue
			}
			switch ev.Event {
			case ipc.EventAgentMessageChunk:
				contentBuilder.WriteString(ev.Chunk)
			case ipc.EventAgentMessageComplete:
				if ev.Content != "" && contentBuilder.Len() == 0 {
					contentBuilder.WriteString(ev.Content)
				}
			case ipc.EventAgentThought:
				if ev.Thought != "" {
					thoughtBuilder.WriteString(ev.Thought)
				}
			case ipc.EventSessionError:
				if ev.Error != "" {
					lastError = fmt.Errorf("%s", ev.Error)
				}
			}
			if ev.Thought != "" && thoughtBuilder.Len() == 0 {
				thoughtBuilder.WriteString(ev.Thought)
			}
		}
	}

	stopTypingFunc()

	if wasAborted {
		abortedText := "⚠️ Proses investigasi dibatalkan oleh pengguna."
		if contentBuilder.Len() > 0 {
			_ = s.db.SaveChatMessage(&db.ChatMessage{
				ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
				SessionID: sessionID,
				Role:      "assistant",
				Content:   contentBuilder.String() + " [Aborted]",
				Status:    "ABORTED",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			})
		}
		return s.sendMarkdownOrPlain(c, abortedText)
	}

	if lastError != nil && contentBuilder.Len() == 0 {
		errMsg := fmt.Sprintf("⚠️ Terjadi kesalahan saat memproses investigasi: %v", lastError)
		_ = s.db.SaveChatMessage(&db.ChatMessage{
			ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   errMsg,
			Status:    "FAILED",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return s.sendMarkdownOrPlain(c, errMsg)
	}

	finalContent := contentBuilder.String()
	finalThought := thoughtBuilder.String()
	if finalContent == "" && finalThought != "" {
		finalContent = finalThought
	}
	if finalContent == "" {
		finalContent = "Penyelidikan selesai tanpa respons teks tambahan."
	}

	finalText := FormatFinalResponse(finalContent, finalThought)
	finalText = SanitizeTelegramMarkdown(finalText)

	// Persist assistant message to database
	assistantMsg := &db.ChatMessage{
		ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   finalContent,
		Status:    "COMPLETED",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if finalThought != "" {
		assistantMsg.Thought = &finalThought
	}
	_ = s.db.SaveChatMessage(assistantMsg)
	_ = s.db.TouchChatSession(sessionID, finalContent)

	// Send formatted chunks to Telegram with fallback to plain text
	chunks := SplitMessage(finalText, DefaultMaxMessageLength)
	var validChunks []string
	for _, chunk := range chunks {
		trimmed := strings.TrimSpace(chunk)
		if trimmed != "" {
			validChunks = append(validChunks, trimmed)
		}
	}
	if len(validChunks) == 0 {
		validChunks = []string{finalText}
	}

	for i, chunk := range validChunks {
		isLast := i == len(validChunks)-1
		var sendOpts *telebot.SendOptions
		if isLast {
			sendOpts = &telebot.SendOptions{
				ParseMode:   telebot.ModeMarkdown,
				ReplyMarkup: inlineActionsMarkup(),
			}
		} else {
			sendOpts = &telebot.SendOptions{
				ParseMode: telebot.ModeMarkdown,
			}
		}

		err := c.Send(chunk, sendOpts)
		if err != nil {
			if isLast {
				_ = c.Send(chunk, &telebot.SendOptions{ReplyMarkup: inlineActionsMarkup()})
			} else {
				_ = c.Send(chunk)
			}
		}
	}

	return nil
}

// sendMarkdownOrPlain sends a message to Telegram using Markdown formatting,
// with safe fallback to plain text if Telegram fails to parse formatting tags.
func (s *BotService) sendMarkdownOrPlain(c telebot.Context, text string) error {
	chunks := SplitMessage(text, DefaultMaxMessageLength)
	for _, chunk := range chunks {
		trimmed := strings.TrimSpace(chunk)
		if trimmed == "" {
			continue
		}
		err := c.Send(trimmed, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		if err != nil {
			if sendErr := c.Send(trimmed); sendErr != nil {
				return sendErr
			}
		}
	}
	return nil
}
