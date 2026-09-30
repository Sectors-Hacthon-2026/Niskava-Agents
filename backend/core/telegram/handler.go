package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"gopkg.in/telebot.v3"
)

var (
	btnNew    = telebot.InlineButton{Unique: "btn_new_session", Text: "🔄 New Session"}
	btnExport = telebot.InlineButton{Unique: "btn_export_session", Text: "📑 Export Document"}
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
				return c.Reply("⛔ Access denied. Your Telegram account is not authorized to access Niskava Agent.")
			}
			return next(c)
		}
	}
}

// handleStart responds to /start and /help commands with welcome instructions.
func (s *BotService) handleStart(c telebot.Context) error {
	welcomeMsg := `🔍 *Niskava Agent — IDX Autonomous Financial Intelligence*

Welcome! Niskava Agent is an autonomous financial intelligence platform built for the Indonesia Stock Exchange (IDX). It bridges quantitative facts from Sectors Financial API v2 with qualitative corporate disclosures and market filings.

*Available Commands:*
• /start or /help — Display usage guide and instructions
• /new or /reset — Start a fresh research/investigation session
• /export — Download session research report (.md)
• /status — View session activity, model configuration, and status
• /stop — Abort currently executing investigation

Feel free to ask any question or command for IDX equities (e.g., _"Analyze volume anomalies for ANTM over the last 30 days"_).`

	formatted := FormatFinalResponse(welcomeMsg, "")
	return s.sendMarkdownOrPlain(c, formatted)
}

// handleNewSession creates a brand-new chat session for the current Telegram chat.
func (s *BotService) handleNewSession(c telebot.Context) error {
	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database is currently unavailable.")
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
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Failed to create a new session: %v", err))
	}

	msg := fmt.Sprintf("✨ *New session created successfully!*\nSession ID: `%s`\n\nPrevious conversation history has been archived. You may now submit your queries or tickers to analyze.", newSessionID)
	return s.sendMarkdownOrPlain(c, FormatFinalResponse(msg, ""))
}

// handleStop aborts any active running query in the current chat session.
func (s *BotService) handleStop(c telebot.Context) error {
	if s.db == nil || s.sm == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Internal service is unavailable.")
	}

	chat, err := s.db.GetTelegramChat(c.Chat().ID)
	if err != nil || chat == nil || chat.CurrentSessionID == "" {
		return s.sendMarkdownOrPlain(c, "ℹ️ No active session found.")
	}

	aborted := s.sm.Abort(chat.CurrentSessionID)
	if aborted {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("🛑 Investigation execution for session `%s` has been cancelled.", chat.CurrentSessionID))
	}

	return s.sendMarkdownOrPlain(c, fmt.Sprintf("ℹ️ No active running investigation was found in session `%s`.", chat.CurrentSessionID))
}

// handleStatus returns diagnostic status information about the active session and configuration.
func (s *BotService) handleStatus(c telebot.Context) error {
	sessionID := "Not initialized"
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

	busyText := "🟢 Idle (Ready for queries)"
	if isBusy {
		busyText = "🟡 Processing (Running analysis)"
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

	statusMsg := fmt.Sprintf(`📊 *Niskava Agent Status*

• *Session ID:* `+"`%s`"+`
• *Activity Status:* %s
• *AI Model:* %s
• *Market:* %s
• *Offline Mode:* %t`,
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
	sb.WriteString("# Niskava Market Intelligence Research Report\n\n")
	sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
	sb.WriteString(fmt.Sprintf("- **AI Model:** %s\n", model))
	sb.WriteString(fmt.Sprintf("- **Export Timestamp:** %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	sb.WriteString("> **Compliance Notice (Law 2 — Non-Advisory Boundary):**\n")
	sb.WriteString("> Niskava Agent is an autonomous market intelligence and research platform for the Indonesia Stock Exchange (IDX), NOT a registered investment advisor or broker-dealer. All data, anomaly detections, and evidence correlations are provided independently for fact verification and capital market research purposes only. No part of this document constitutes a buy/sell recommendation or licensed financial advice.\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Conversation & Investigation History\n\n")

	for i, msg := range history {
		roleTitle := "👤 User"
		if strings.EqualFold(msg.Role, "assistant") {
			roleTitle = "🤖 Niskava Agent"
		} else if strings.EqualFold(msg.Role, "system") {
			roleTitle = "⚙️ System"
		} else if strings.EqualFold(msg.Role, "tool") {
			roleTitle = "🔧 Tool Execution"
		}

		sb.WriteString(fmt.Sprintf("### %d. %s (`%s`)\n\n", i+1, roleTitle, msg.CreatedAt))

		if msg.Thought != nil && strings.TrimSpace(*msg.Thought) != "" {
			sb.WriteString("<details>\n<summary>Reasoning Process (Chain of Thought)</summary>\n\n")
			sb.WriteString(strings.TrimSpace(*msg.Thought))
			sb.WriteString("\n\n</details>\n\n")
		}

		sb.WriteString(strings.TrimSpace(msg.Content))
		sb.WriteString("\n\n---\n\n")
	}

	sb.WriteString("*This document was generated automatically by Niskava Agent.*\n")
	return sb.String()
}

// handleExport generates and sends a markdown document containing the current session's chat history.
func (s *BotService) handleExport(c telebot.Context) error {
	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database is currently unavailable.")
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
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Failed to load chat session: %v", err))
	}

	session, err := s.db.GetChatSession(sessionID)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Failed to load session details: %v", err))
	}

	history, err := s.db.GetChatHistory(sessionID, 500)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Failed to load chat history: %v", err))
	}

	if len(history) == 0 {
		emptyMsg := "⚠️ No conversation history found to export for this session."
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
		Caption:  fmt.Sprintf("📑 Research Report Session %s", sessionID),
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
			msg := fmt.Sprintf("⚠️ Too many requests. Please wait %d seconds.", int(retryAfter.Seconds())+1)
			if err := c.Reply(msg); err != nil {
				return c.Send(msg)
			}
			return nil
		}
	}

	if s.db == nil {
		return s.sendMarkdownOrPlain(c, "⚠️ Database is currently unavailable.")
	}

	var userID int64
	var username string
	if sender != nil {
		userID = sender.ID
		username = sender.Username
	}

	sessionID, err := s.db.GetOrCreateTelegramChatSession(c.Chat().ID, userID, username)
	if err != nil {
		return s.sendMarkdownOrPlain(c, fmt.Sprintf("⚠️ Failed to initialize chat session: %v", err))
	}

	if s.sm != nil && s.sm.IsBusy(sessionID) {
		return s.sendMarkdownOrPlain(c, "⚠️ Session is currently processing a previous query. Type /stop to cancel.")
	}

	childCtx, cancelChild := context.WithCancel(s.ctx)
	defer cancelChild()

	if s.sm != nil {
		if !s.sm.Register(sessionID, cancelChild) {
			return s.sendMarkdownOrPlain(c, "⚠️ Session is currently processing a previous query. Type /stop to cancel.")
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

	// Auto-update session title from the first user message so each session is
	// identifiable in the Web Workspace sidebar by its actual research topic.
	if s.db != nil {
		if sess, err := s.db.GetChatSession(sessionID); err == nil && sess != nil && sess.MessageCount == 0 {
			topic := strings.TrimSpace(prompt)
			topic = strings.TrimLeft(topic, "#> *`-_")
			topic = strings.TrimSpace(topic)
			if len(topic) > 45 {
				topic = topic[:42] + "..."
			}
			if topic != "" {
				_ = s.db.UpdateChatSession(sessionID, &topic, nil, nil)
			}
		}
	}

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
	lang := "en"
	var offline bool
	var pythonBin, enginePath string
	envOverrides := make(map[string]string)
	if s.cfg != nil {
		if s.cfg.Preferences.Language != "" {
			lang = s.cfg.Preferences.Language
		}
		offline = s.cfg.Preferences.OfflineMode
		pythonBin = s.cfg.Engine.PythonBin
		enginePath = s.cfg.Engine.EnginePath
		envOverrides = s.cfg.BuildSubprocessEnv()
	}

	runnerParams := ipc.RunnerParams{
		PythonBin:    pythonBin,
		EnginePath:   enginePath,
		WorkDir:      wd,
		DBPath:       s.db.Path,
		SessionID:    sessionID,
		Prompt:       prompt,
		Offline:      offline,
		Language:     lang,
		EnvOverrides: envOverrides,
	}

	eventsChan, errChan := ipc.RunConversationStream(childCtx, runnerParams)

	var contentBuilder strings.Builder
	var thoughtBuilder strings.Builder
	var generatedPdfs []string
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
			case ipc.EventPdfReportReady:
				if ev.PdfPath != "" {
					generatedPdfs = append(generatedPdfs, ev.PdfPath)
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
		abortedText := "⚠️ Investigation process was cancelled by the user."
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
		errMsg := fmt.Sprintf("⚠️ An error occurred while processing the investigation: %v", lastError)
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
		finalContent = "Investigation completed without additional text response."
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

	// Dispatch any physical PDF reports generated during the turn
	for _, pdfPath := range generatedPdfs {
		cleanPath := strings.TrimSpace(pdfPath)
		if cleanPath == "" {
			continue
		}
		if _, statErr := os.Stat(cleanPath); statErr != nil {
			continue
		}
		doc := &telebot.Document{
			File:     telebot.FromDisk(cleanPath),
			FileName: filepath.Base(cleanPath),
			MIME:     "application/pdf",
			Caption:  fmt.Sprintf("📑 Research Report PDF: %s", filepath.Base(cleanPath)),
		}
		_ = c.Send(doc)
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
