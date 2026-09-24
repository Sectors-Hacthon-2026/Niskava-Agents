package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// ReplBackSentinel is the sentinel return value from RunLiveREPL when the user requests returning to launcher.
const ReplBackSentinel = "__back__"
const replBackSentinel = ReplBackSentinel

var (
	// Terminal Color Styles (Binance Dark Financial OSINT Aesthetic)
	promptBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	userBubbleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Foreground(ColorFg).
			Padding(0, 1).
			MarginTop(1)

	thoughtStyle = lipgloss.NewStyle().
			Foreground(ColorThought).
			Italic(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	observationStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	replAnomalyBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorDanger).
				Foreground(ColorFg).
				Padding(0, 1).
				MarginTop(1)

	supportedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorSuccess).
				Padding(0, 1)

	uncertainBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorWarning).
				Padding(0, 1)

	contradictedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorFg).
				Background(ColorDanger).
				Padding(0, 1)
)

// SlashCommand represents a registered slash command in the interactive REPL.
type SlashCommand struct {
	Command     string
	Category    string
	Description string
}

func getDefaultSlashCommands() []SlashCommand {
	return GetLocalizedSlashCommands()
}

// ReplInputModel is the Bubbletea interactive text input model with OpenCode-style slash popup and prompt history navigation.
type ReplInputModel struct {
	TextInput        textinput.Model
	PromptPrefix     string
	SlashCommands    []SlashCommand
	FilteredCommands []SlashCommand
	SlashCursor      int
	SlashActive      bool
	SubmittedValue   string
	Quitting         bool
	LastExitTime     time.Time
	ExitWarning      bool
	History          []string
	HistoryIndex     int
	DraftValue       string
	NavigatingHist   bool
}

// NewReplInputModel initializes the interactive REPL prompt input.
func NewReplInputModel(promptPrefix string) ReplInputModel {
	return NewReplInputModelWithHistory(promptPrefix, nil)
}

// NewReplInputModelWithHistory initializes prompt input with existing prompt history.
func NewReplInputModelWithHistory(promptPrefix string, history []string) ReplInputModel {
	ti := textinput.New()
	ti.Prompt = promptBoxStyle.Render(promptPrefix + " ")
	ti.Placeholder = T("prompt_placeholder")
	ti.Focus()

	cmds := GetLocalizedSlashCommands()
	return ReplInputModel{
		TextInput:        ti,
		PromptPrefix:     promptPrefix,
		SlashCommands:    cmds,
		FilteredCommands: cmds,
		History:          history,
		HistoryIndex:     len(history),
	}
}

func (m ReplInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ReplInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Priority 1: Esc menutup slash popup jika aktif
			if m.SlashActive && msg.Type == tea.KeyEsc {
				m.SlashActive = false
				m.ExitWarning = false
				m.LastExitTime = time.Time{}
				return m, nil
			}
			if msg.Type == tea.KeyEsc && strings.TrimSpace(m.TextInput.Value()) != "" {
				m.TextInput.SetValue("")
				m.TextInput.SetCursor(0)
				m.ExitWarning = false
				m.LastExitTime = time.Time{}
				m.NavigatingHist = false
				m.HistoryIndex = len(m.History)
				return m, nil
			}

			// 3-Layer Double Press Esc / Ctrl+C Safety Protection
			now := time.Now()
			if !m.LastExitTime.IsZero() && now.Sub(m.LastExitTime) <= 2*time.Second {
				m.Quitting = true
				m.SubmittedValue = "/exit"
				return m, tea.Quit
			}
			m.LastExitTime = now
			m.ExitWarning = true
			return m, nil

		case tea.KeyUp:
			if m.SlashActive && len(m.FilteredCommands) > 0 {
				if m.SlashCursor > 0 {
					m.SlashCursor--
				} else {
					m.SlashCursor = len(m.FilteredCommands) - 1
				}
				return m, nil
			}
			// OpenCode-style prompt history navigation (Up Arrow)
			if !m.SlashActive && len(m.History) > 0 {
				if !m.NavigatingHist {
					m.DraftValue = m.TextInput.Value()
					m.NavigatingHist = true
					m.HistoryIndex = len(m.History)
				}
				if m.HistoryIndex > 0 {
					m.HistoryIndex--
					m.TextInput.SetValue(m.History[m.HistoryIndex])
					m.TextInput.SetCursor(len(m.TextInput.Value()))
				}
				return m, nil
			}

		case tea.KeyDown:
			if m.SlashActive && len(m.FilteredCommands) > 0 {
				if m.SlashCursor < len(m.FilteredCommands)-1 {
					m.SlashCursor++
				} else {
					m.SlashCursor = 0
				}
				return m, nil
			}
			// OpenCode-style prompt history navigation (Down Arrow)
			if !m.SlashActive && m.NavigatingHist {
				if m.HistoryIndex < len(m.History)-1 {
					m.HistoryIndex++
					m.TextInput.SetValue(m.History[m.HistoryIndex])
					m.TextInput.SetCursor(len(m.TextInput.Value()))
				} else if m.HistoryIndex == len(m.History)-1 {
					m.HistoryIndex = len(m.History)
					m.TextInput.SetValue(m.DraftValue)
					m.TextInput.SetCursor(len(m.TextInput.Value()))
					m.NavigatingHist = false
				}
				return m, nil
			}

		case tea.KeyTab:
			if m.SlashActive && len(m.FilteredCommands) > 0 {
				selected := m.FilteredCommands[m.SlashCursor].Command
				m.TextInput.SetValue(selected)
				m.TextInput.SetCursor(len(selected))
				m.SlashActive = false
				return m, nil
			}

		case tea.KeyEnter:
			val := strings.TrimSpace(m.TextInput.Value())
			if m.SlashActive && len(m.FilteredCommands) > 0 && strings.HasPrefix(val, "/") {
				val = m.FilteredCommands[m.SlashCursor].Command
			}
			m.SubmittedValue = val
			return m, tea.Quit
		}
	}

	m.TextInput, cmd = m.TextInput.Update(msg)

	// Live filter slash commands when input starts with '/'
	val := strings.TrimSpace(m.TextInput.Value())
	if strings.HasPrefix(val, "/") && !strings.Contains(val, " ") {
		m.SlashActive = true
		m.FilteredCommands = nil
		for _, sc := range m.SlashCommands {
			if strings.HasPrefix(sc.Command, val) || strings.HasPrefix(val, sc.Command) || strings.Contains(sc.Command, strings.ToLower(val)) {
				m.FilteredCommands = append(m.FilteredCommands, sc)
			}
		}
		if len(m.FilteredCommands) == 0 {
			m.FilteredCommands = m.SlashCommands
		}
		if m.SlashCursor >= len(m.FilteredCommands) {
			m.SlashCursor = 0
		}
	} else {
		m.SlashActive = false
		m.SlashCursor = 0
		m.FilteredCommands = m.SlashCommands
	}

	return m, cmd
}

func (m ReplInputModel) View() string {
	var b strings.Builder

	// Render input prompt box
	b.WriteString("\n" + m.TextInput.View() + "\n")

	// Render long prompt drafting character counter indicator if input is long (>50 chars)
	val := strings.TrimSpace(m.TextInput.Value())
	if len(val) >= 50 && !m.SlashActive {
		countPill := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			Render(fmt.Sprintf("  ✍️  Long Prompt Active (%d chars) • [Enter to execute, Esc to clear]", len(val)))
		b.WriteString(countPill + "\n")
	}

	// Render double-press exit warning hint if active
	if m.ExitWarning && !m.LastExitTime.IsZero() && time.Since(m.LastExitTime) <= 2*time.Second {
		warningStr := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(T("repl_exit_confirm"))
		b.WriteString(warningStr + "\n")
	}

	// Render OpenCode-style Slash Autocomplete Popup Box when slash active
	if m.SlashActive && len(m.FilteredCommands) > 0 {
		popupHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
			Padding(0, 1).
			Render(T("slash_popup_header"))

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(0, 1)

		var popupLines []string
		popupLines = append(popupLines, popupHeader)

		for i, sc := range m.FilteredCommands {
			cursor := "  "
			if i == m.SlashCursor {
				cursor = "> "
			}

			cmdStr := fmt.Sprintf("%-12s", sc.Command)
			catBadge := ""
			if sc.Category != "" {
				catBadge = lipgloss.NewStyle().
					Foreground(ColorMuted).
					Render(fmt.Sprintf("[%s] ", sc.Category))
			}
			descStr := sc.Description

			if i == m.SlashCursor {
				cmdR := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(cmdStr)
				descR := lipgloss.NewStyle().Foreground(ColorFg).Render(descStr)
				popupLines = append(popupLines, fmt.Sprintf("%s%s%s%s", lipgloss.NewStyle().Foreground(ColorAccent).Render(cursor), cmdR, catBadge, descR))
			} else {
				cmdR := lipgloss.NewStyle().Foreground(ColorAccent).Render(cmdStr)
				descR := lipgloss.NewStyle().Foreground(ColorMuted).Render(descStr)
				popupLines = append(popupLines, fmt.Sprintf("  %s%s%s", cmdR, catBadge, descR))
			}
		}

		b.WriteString(boxStyle.Render(strings.Join(popupLines, "\n")) + "\n")
	}

	return b.String()
}

// renderResumedHistory displays past user and assistant turns when resuming an earlier session.
func renderResumedHistory(appDB *db.DB, sessionID string) {
	if appDB == nil {
		return
	}
	history, err := appDB.GetChatHistory(sessionID, 50)
	if err != nil || len(history) == 0 {
		return
	}

	fmt.Println()
	divider := lipgloss.NewStyle().Foreground(ColorMuted).Render(TF("repl_resumed_history_divider", len(history)))
	fmt.Println(divider)

	for _, msg := range history {
		if msg.Role == "user" {
			userBox := userBubbleStyle.Render(TF("repl_user_label", msg.Content))
			fmt.Println(userBox)
		} else if msg.Role == "assistant" {
			fmt.Println("\n" + lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render(T("repl_agent_label")))
			renderFinalMarkdown(msg.Content)
		}
	}

	fmt.Println(lipgloss.NewStyle().Foreground(ColorMuted).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
}

// RunLiveREPL starts an interactive, conversational research assistant session.
// If initialSessionID is provided and non-empty, it resumes that session directly.
// Returns replBackSentinel ("__back__") if user typed /back to return to launcher, or "" if user exited.
func RunLiveREPL(cfg *config.Config, appDB *db.DB, serverURL string, initialSessionID ...string) string {
	// Determine active model display
	modelLabel := cfg.Auth.OpenAIModel
	if modelLabel == "" {
		if cfg.Auth.GeminiModel != "" {
			modelLabel = cfg.Auth.GeminiModel
		} else {
			modelLabel = "hermes"
		}
	}

	sessionID := fmt.Sprintf("CHAT-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
	if len(initialSessionID) > 0 && strings.TrimSpace(initialSessionID[0]) != "" {
		sessionID = strings.TrimSpace(initialSessionID[0])
	}

	renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)

	if len(initialSessionID) > 0 && strings.TrimSpace(initialSessionID[0]) != "" {
		renderResumedHistory(appDB, sessionID)
	}

	promptPrefix := fmt.Sprintf("niskava [%s] >", modelLabel)

	var promptHistory []string

	for {
		// Run interactive Bubbletea prompt input with live OpenCode slash popup and prompt history
		inputModel := NewReplInputModelWithHistory(promptPrefix, promptHistory)
		p := tea.NewProgram(inputModel)
		m, err := p.Run()
		if err != nil {
			fmt.Printf("\n%s\n", T("repl_exit_msg"))
			return ""
		}

		input := strings.TrimSpace(m.(ReplInputModel).SubmittedValue)
		if input == "" {
			continue
		}

		// Save non-slash prompts to history
		if !strings.HasPrefix(input, "/") {
			if len(promptHistory) == 0 || promptHistory[len(promptHistory)-1] != input {
				promptHistory = append(promptHistory, input)
			}
		}

		// Handle Slash Commands
		lower := strings.ToLower(input)
		if lower == "/exit" || lower == "exit" || lower == "quit" || lower == ":q" {
			fmt.Println(T("repl_exit_msg"))
			return replBackSentinel
		}

		if lower == "/back" || lower == "back" {
			fmt.Print(T("repl_back_msg"))
			return replBackSentinel
		}

		if lower == "/help" {
			printHelp()
			continue
		}

		if lower == "/clear" || lower == "clear" {
			fmt.Print("\033[H\033[2J")
			renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
			continue
		}

		if lower == "/reset" {
			if appDB != nil {
				_ = appDB.ClearMemoryGraph()
			}
			sessionID = fmt.Sprintf("CHAT-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
			fmt.Printf(T("repl_session_reset"), sessionID)
			continue
		}

		if lower == "/graph" {
			graphURL := fmt.Sprintf("%s/graph", serverURL)
			fmt.Printf(T("repl_open_graph"), graphURL)
			_ = server.OpenBrowser(graphURL)
			continue
		}

		if lower == "/web" {
			fmt.Printf(T("repl_open_web"), serverURL)
			_ = server.OpenBrowser(serverURL)
			continue
		}

		if lower == "/health" {
			printHealth(cfg)
			continue
		}

		if strings.HasPrefix(lower, "/lang") {
			parts := strings.Fields(input)
			if len(parts) > 1 {
				langArg := parts[1]
				SetLanguage(langArg)
			} else {
				langModel := NewLangSelectorModel()
				pLang := tea.NewProgram(langModel)
				mLang, errLang := pLang.Run()
				if errLang == nil {
					selLang := mLang.(LangSelectorModel).Selected
					if selLang != "" {
						SetLanguage(selLang)
					}
				}
			}
			cfg.Preferences.Language = ActiveLanguage

			fmt.Print("\033[H\033[2J")
			renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
			activeInfo := GetActiveLanguageInfo()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(TF("repl_lang_switched", activeInfo.FlagSymbol, activeInfo.NativeName, activeInfo.Code)))
			continue
		}

		if lower == "/sessions" {
			printSessions(appDB)
			continue
		}

		if lower == "/chats" {
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("repl_db_unavailable")))
				continue
			}
			chatList, _, errList := appDB.ListChatSessions(30, 0, "")
			if errList != nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(TF("repl_chats_fetch_err", errList)))
				continue
			}
			selector := NewSessionSelectorModel(chatList)
			pSel := tea.NewProgram(selector)
			mSel, errRun := pSel.Run()
			if errRun == nil {
				res := mSel.(SessionSelectorModel)
				if !res.Canceled && res.SelectedSession != nil {
					prevSessionID := sessionID
					sessionID = res.SelectedSession.ID
					fmt.Print("\033[H\033[2J")
					renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
					title := res.SelectedSession.Title
					if title == "" {
						title = res.SelectedSession.ID
					}
					fmt.Print(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(TF("repl_chats_saved_notice", prevSessionID, sessionID, title)))
					renderResumedHistory(appDB, sessionID)
				}
			}
			continue
		}

		if strings.HasPrefix(lower, "/resume") {
			parts := strings.Fields(input)
			if len(parts) < 2 {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(T("repl_resume_usage")))
				continue
			}
			targetID := strings.TrimSpace(parts[1])
			if appDB != nil {
				sess, errGet := appDB.GetChatSession(targetID)
				if errGet != nil || sess == nil {
					fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(TF("repl_resume_not_found", targetID)))
					continue
				}
				prevSessionID := sessionID
				sessionID = sess.ID
				fmt.Print("\033[H\033[2J")
				renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
				title := sess.Title
				if title == "" {
					title = sess.ID
				}
				fmt.Print(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(TF("repl_chats_saved_notice", prevSessionID, sessionID, title)))
				renderResumedHistory(appDB, sessionID)
			}
			continue
		}

		// Execute conversational research turn with verbatim user prompt
		executeChatTurn(input, sessionID, serverURL, cfg, appDB)
	}
}

func renderBanner(modelLabel, serverURL, sessionID, dbPath string) {
	fmt.Print(RenderHUDHeader(modelLabel, serverURL, dbPath, sessionID))
	helpHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(T("banner_hint"))
	fmt.Printf("\n%s\n", helpHint)
}

func startLiveSpinner(ctx context.Context, getStatus func() string) func() {
	spinnerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-spinnerCtx.Done():
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				frame := frames[idx%len(frames)]
				idx++
				status := getStatus()
				if status != "" {
					spinnerLine := lipgloss.NewStyle().Foreground(ColorAccent).Render(fmt.Sprintf("  %s %s", frame, status))
					fmt.Print("\r\033[K" + spinnerLine + "\r")
				}
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

func executeChatTurn(prompt, sessionID, serverURL string, cfg *config.Config, appDB *db.DB) {
	usingDaemon := IsDaemonAlive(serverURL)

	// Record User Message in SQLite if running standalone subprocess mode
	if !usingDaemon && appDB != nil {
		userMsg := &db.ChatMessage{
			ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
			SessionID: sessionID,
			Role:      "user",
			Content:   prompt,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		_ = appDB.SaveChatMessage(userMsg)
	}

	// Sleek session divider (avoids redundant duplicate user input box)
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Foreground(ColorAccent).Render(TF("repl_session_banner", sessionID)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	var interrupted atomic.Bool
	lastSignalTime := time.Time{}
	go func() {
		for {
			select {
			case <-sigChan:
				now := time.Now()
				if !lastSignalTime.IsZero() && now.Sub(lastSignalTime) <= 2*time.Second {
					interrupted.Store(true)
					fmt.Println("\n" + lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render(strings.TrimSpace(T("repl_execution_cancelled"))))
					cancel()
					return
				}
				lastSignalTime = now
				fmt.Print("\r\033[K" + lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render(T("repl_interrupt_confirm")) + "\r")
			case <-ctx.Done():
				return
			}
		}
	}()

	var eventsChan <-chan ipc.Event
	var errChan <-chan error

	if usingDaemon {
		eventsChan, errChan = StreamChatViaSSE(ctx, serverURL, sessionID, prompt)
	} else {
		pythonBin := ""
		if cfg != nil {
			pythonBin = cfg.Engine.PythonBin
		}
		pythonBin = ipc.ResolvePythonBin(pythonBin)

		wd, _ := os.Getwd()
		runnerParams := ipc.RunnerParams{
			PythonBin: pythonBin,
			WorkDir:   wd,
			DBPath:    cfg.Storage.DBPath,
			Prompt:    prompt,
			SessionID: sessionID,
			Offline:   cfg.Preferences.OfflineMode,
			Language:  cfg.Preferences.Language,
		}
		eventsChan, errChan = ipc.RunSubprocess(ctx, runnerParams)
	}

	var (
		assistantResponse strings.Builder
		lastThought       string
		totalAnomalies    int
		totalFindings     int
	)

	turnStart := time.Now()
	modelLabel := cfg.Auth.OpenAIModel
	if modelLabel == "" {
		if cfg.Auth.GeminiModel != "" {
			modelLabel = cfg.Auth.GeminiModel
		} else {
			modelLabel = "hermes"
		}
	}

	var statusText atomic.Value
	initStatus := strings.TrimPrefix(TF("thinking_init", modelLabel), "  ⠋ ")
	statusText.Store(initStatus)

	stopSpinner := startLiveSpinner(ctx, func() string {
		if val := statusText.Load(); val != nil {
			return val.(string)
		}
		return ""
	})
	defer stopSpinner()

	activeErrChan := errChan
	for {
		select {
		case <-ctx.Done():
			stopSpinner()
			if interrupted.Load() {
				fmt.Print("\r\033[K")
				fmt.Print(T("repl_execution_cancelled"))
				return
			}

		case err, ok := <-activeErrChan:
			if !ok {
				activeErrChan = nil
				continue
			}
			if err != nil {
				stopSpinner()
				fmt.Print("\r\033[K")
				fmt.Printf("\n[Subprocess Error]: %v\n", err)
				return
			}

		case ev, ok := <-eventsChan:
			if !ok {
				stopSpinner()
				// Process finished: clear spinner and render final markdown
				fmt.Print("\r\033[K")
				renderFinalMarkdown(assistantResponse.String())

				// Save assistant response in SQLite if running standalone subprocess
				if !usingDaemon && appDB != nil && assistantResponse.Len() > 0 {
					asstMsg := &db.ChatMessage{
						ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
						SessionID: sessionID,
						Role:      "assistant",
						Content:   assistantResponse.String(),
						Thought:   &lastThought,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
					}
					_ = appDB.SaveChatMessage(asstMsg)
				}

				// Render official completion badge with timing & statistics
				fmt.Print(renderCompletionBadge(time.Since(turnStart), sessionID, modelLabel, totalAnomalies, totalFindings))
				return
			}

			switch ev.Event {
			case ipc.EventAgentThought:
				lastThought = ev.Thought
				fmt.Print("\r\033[K")
				fmt.Printf("💭 %s\n", thoughtStyle.Render(ev.Thought))
				vMsg := strings.TrimPrefix(TF("thinking_verify", modelLabel), "  ⠋ ")
				statusText.Store(vMsg)

			case ipc.EventAgentToolCall:
				fmt.Print("\r\033[K")
				argsJSON := ""
				if ev.Args != nil {
					argsJSON = fmt.Sprintf(" %v", ev.Args)
				}
				fmt.Printf("⚡ %s%s\n", toolCallStyle.Render("[TOOL CALL: "+ev.Tool+"]"), argsJSON)
				tMsg := strings.TrimPrefix(TF("tool_executing", ev.Tool), "  ⠋ ")
				statusText.Store(tMsg)

			case ipc.EventAgentObservation:
				fmt.Print("\r\033[K")
				fmt.Printf("🔎 %s\n", observationStyle.Render(ev.Summary))
				sMsg := strings.TrimPrefix(TF("thinking_synthesize", modelLabel), "  ⠋ ")
				statusText.Store(sMsg)

			case ipc.EventAnomalyDetected:
				fmt.Print("\r\033[K")
				totalAnomalies++
				anomalyText := TF(
					"repl_anomaly_alert",
					ev.MetricType, ev.Ticker, ev.ZScore, ev.MetricValue, ev.BaselineValue,
				)
				fmt.Println(replAnomalyBoxStyle.Render(anomalyText))

			case ipc.EventFindingEmitted:
				fmt.Print("\r\033[K")
				totalFindings++
				badge := supportedBadgeStyle.Render("[SUPPORTED]")
				if ev.VerificationStat == "UNCERTAIN" {
					badge = uncertainBadgeStyle.Render("[UNCERTAIN]")
				} else if ev.VerificationStat == "CONTRADICTED" {
					badge = contradictedBadgeStyle.Render("[CONTRADICTED]")
				}
				fmt.Printf("\n%s %s (Confidence: %.0f%%)\n", badge, lipgloss.NewStyle().Bold(true).Render(ev.Title), ev.ConfidenceScore*100)
				fmt.Printf("   %s\n", ev.ClaimText)

			case ipc.EventAgentMessageChunk:
				assistantResponse.WriteString(ev.Chunk)
				words := len(strings.Fields(assistantResponse.String()))
				fmt.Print("\r\033[K" + lipgloss.NewStyle().Foreground(ColorAccent).Italic(true).Render(TF("thinking_drafting", modelLabel, words)) + "\r")

			case ipc.EventAgentMessageComplete:
				if assistantResponse.Len() == 0 {
					assistantResponse.WriteString(ev.Content)
				}

			case ipc.EventSessionComplete:
				if ev.TotalAnomalies > 0 {
					totalAnomalies = ev.TotalAnomalies
				}
				if ev.TotalFindings > 0 {
					totalFindings = ev.TotalFindings
				}

			case ipc.EventSessionError:
				fmt.Print("\r\033[K")
				if assistantResponse.Len() == 0 {
					errBox := lipgloss.NewStyle().
						Border(lipgloss.RoundedBorder()).
						BorderForeground(ColorDanger).
						Padding(0, 1).
						Foreground(ColorFg).
						Render(fmt.Sprintf("❌ [SESSION ERROR]: %s", ev.Error))
					assistantResponse.WriteString(errBox)
				}
			}
		}
	}
}

func renderCompletionBadge(duration time.Duration, sessionID, model string, anomalies, findings int) string {
	sep := lipgloss.NewStyle().Foreground(ColorMuted).Render("─────────────────────────────────────────────────────────────────────────────")
	badge := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(T("badge_completed"))
	detail := TF("badge_completed_detail", duration.Seconds(), model, sessionID)
	if anomalies > 0 || findings > 0 {
		detail += TF("badge_completed_counts", anomalies, findings)
	}
	return fmt.Sprintf("\n%s\n%s %s\n%s\n", sep, badge, detail, sep)
}

func renderFinalMarkdown(markdownContent string) {
	if strings.TrimSpace(markdownContent) == "" {
		return
	}

	fmt.Println()
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(95),
	)
	if err == nil {
		out, renderErr := renderer.Render(markdownContent)
		if renderErr == nil {
			fmt.Println(out)
			return
		}
	}

	// Fallback plain print if glamour encounters error
	fmt.Println(markdownContent)
}

func printHelp() {
	fmt.Println(T("help_header"))
	fmt.Println(T("help_prompt_desc"))
	fmt.Println(T("help_ticker_desc"))
	fmt.Println(T("help_graph_desc"))
	fmt.Println(T("help_reset_desc"))
	fmt.Println("  • /chats        : " + T("slash_chats_desc"))
	fmt.Println("  • /resume <id>  : " + T("slash_resume_desc"))
	fmt.Println(T("help_sessions_desc"))
	fmt.Println(T("help_web_desc"))
	fmt.Println(T("help_health_desc"))
	fmt.Println(T("help_lang_desc"))
	fmt.Println(T("help_clear_desc"))
	fmt.Println(T("help_exit_desc"))
}

func printHealth(cfg *config.Config) {
	PrintHealthDiagnostics(cfg, "")
}

func printSessions(appDB *db.DB) {
	investigations, err := appDB.ListInvestigations(20)
	if err != nil {
		fmt.Printf("Failed to query database: %v\n", err)
		return
	}

	if len(investigations) == 0 {
		fmt.Println(T("sessions_empty"))
		return
	}

	fmt.Println(T("sessions_header"))
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%-22s %-8s %-12s %-20s %s\n", "SESSION ID", "TICKER", "STATUS", "STARTED AT", "SUMMARY")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")

	for _, inv := range investigations {
		summary := "-"
		if inv.SummaryText != nil && *inv.SummaryText != "" {
			summary = *inv.SummaryText
			if len(summary) > 35 {
				summary = summary[:32] + "..."
			}
		}

		dateStr := inv.StartedAt
		if len(dateStr) > 19 {
			dateStr = strings.Replace(dateStr[:19], "T", " ", 1)
		}

		fmt.Printf("%-22s %-8s %-12s %-20s %s\n", inv.ID, inv.Ticker, inv.Status, dateStr, summary)
	}
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
}
