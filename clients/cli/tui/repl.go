package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

var (
	// Terminal Color Styles (Light Green / Matrix OSINT Aesthetic)
	promptBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4ADE80"))

	userBubbleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22C55E")).
			Foreground(lipgloss.Color("#F8FAFC")).
			Padding(0, 1).
			MarginTop(1)

	thoughtStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#86EFAC")).
			Italic(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FACC15")).
			Bold(true)

	observationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4ADE80"))

	replAnomalyBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#EF4444")).
				Foreground(lipgloss.Color("#FCA5A5")).
				Padding(0, 1).
				MarginTop(1)

	supportedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#052E16")).
				Background(lipgloss.Color("#22C55E")).
				Padding(0, 1)

	uncertainBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#FACC15")).
				Padding(0, 1)

	contradictedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#EF4444")).
				Padding(0, 1)
)

// SlashCommand represents a registered slash command in the interactive REPL.
type SlashCommand struct {
	Command     string
	Description string
}

func getDefaultSlashCommands() []SlashCommand {
	return GetLocalizedSlashCommands()
}

// ReplInputModel is the Bubbletea interactive text input model with OpenCode-style slash popup.
type ReplInputModel struct {
	TextInput        textinput.Model
	PromptPrefix     string
	SlashCommands    []SlashCommand
	FilteredCommands []SlashCommand
	SlashCursor      int
	SlashActive      bool
	SubmittedValue   string
	Quitting         bool
}

// NewReplInputModel initializes the interactive REPL prompt input.
func NewReplInputModel(promptPrefix string) ReplInputModel {
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
			if m.SlashActive && msg.Type == tea.KeyEsc {
				m.SlashActive = false
				return m, nil
			}
			m.Quitting = true
			m.SubmittedValue = "/exit"
			return m, tea.Quit

		case tea.KeyUp:
			if m.SlashActive && len(m.FilteredCommands) > 0 {
				if m.SlashCursor > 0 {
					m.SlashCursor--
				} else {
					m.SlashCursor = len(m.FilteredCommands) - 1
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
	if strings.HasPrefix(val, "/") {
		m.SlashActive = true
		m.FilteredCommands = nil
		for _, sc := range m.SlashCommands {
			if strings.HasPrefix(sc.Command, val) || strings.Contains(sc.Command, strings.ToLower(val)) {
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

	// Render OpenCode-style Slash Autocomplete Popup Box when slash active
	if m.SlashActive && len(m.FilteredCommands) > 0 {
		popupHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#052E16")).
			Background(lipgloss.Color("#22C55E")).
			Padding(0, 1).
			Render("SLASH COMMANDS (Gunakan ↑/↓ untuk memilih, Tab/Enter untuk melengkapi)")

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22C55E")).
			Padding(0, 1)

		var popupLines []string
		popupLines = append(popupLines, popupHeader)

		for i, sc := range m.FilteredCommands {
			cursor := "  "
			if i == m.SlashCursor {
				cursor = "▶ "
			}

			cmdStr := fmt.Sprintf("%-12s", sc.Command)
			descStr := sc.Description

			if i == m.SlashCursor {
				cmdR := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render(cmdStr)
				descR := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8FAFC")).Render(descStr)
				popupLines = append(popupLines, fmt.Sprintf("%s%s %s", lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF87")).Render(cursor), cmdR, descR))
			} else {
				cmdR := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Render(cmdStr)
				descR := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(descStr)
				popupLines = append(popupLines, fmt.Sprintf("  %s %s", cmdR, descR))
			}
		}

		b.WriteString(boxStyle.Render(strings.Join(popupLines, "\n")) + "\n")
	}

	return b.String()
}

// RunLiveREPL starts an interactive, conversational research assistant session.
func RunLiveREPL(cfg *config.Config, appDB *db.DB, serverURL string) {
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

	renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)

	promptPrefix := fmt.Sprintf("niskava [%s] >", modelLabel)

	for {
		// Run interactive Bubbletea prompt input with live OpenCode slash popup
		inputModel := NewReplInputModel(promptPrefix)
		p := tea.NewProgram(inputModel)
		m, err := p.Run()
		if err != nil {
			fmt.Println("\nKeluar dari sesi Live Assistant.")
			break
		}

		input := strings.TrimSpace(m.(ReplInputModel).SubmittedValue)
		if input == "" {
			continue
		}

		// Handle Slash Commands
		lower := strings.ToLower(input)
		if lower == "/exit" || lower == "exit" || lower == "quit" || lower == ":q" {
			fmt.Println("Keluar dari sesi Live REPL.")
			break
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
			fmt.Printf("\n[✓] Sesi direset dan memory graph dibersihkan. Sesi percakapan baru: %s\n", sessionID)
			continue
		}

		if lower == "/graph" {
			graphURL := fmt.Sprintf("%s/graph", serverURL)
			fmt.Printf("Membuka visualisasi Memory Knowledge Graph di browser (%s)...\n", graphURL)
			_ = server.OpenBrowser(graphURL)
			continue
		}

		if lower == "/web" {
			fmt.Printf("Membuka web workspace di browser (%s)...\n", serverURL)
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
			if ActiveLanguage == "en" {
				fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22C55E")).Render(fmt.Sprintf("  [✓] Language preference switched to %s %s (%s).", activeInfo.FlagSymbol, activeInfo.NativeName, activeInfo.Code)))
			} else {
				fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22C55E")).Render(fmt.Sprintf("  [✓] Preferensi bahasa berhasil diubah ke %s %s (%s).", activeInfo.FlagSymbol, activeInfo.NativeName, activeInfo.Code)))
			}
			continue
		}

		if lower == "/sessions" {
			printSessions(appDB)
			continue
		}

		// Execute conversational research turn with verbatim user prompt
		executeChatTurn(input, sessionID, serverURL, cfg, appDB)
	}
}

func renderBanner(modelLabel, serverURL, sessionID, dbPath string) {
	fmt.Print(RenderHUDHeader(modelLabel, serverURL, dbPath, sessionID))
	helpHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(T("banner_hint"))
	fmt.Printf("\n%s\n", helpHint)
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
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render("─── Sesi Investigasi Aktif: "+sessionID+" ─────────────────────────────"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	interrupted := false
	go func() {
		select {
		case <-sigChan:
			interrupted = true
			fmt.Println("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FACC15")).Bold(true).Render("[!] Eksekusi dibatalkan oleh pengguna."))
			cancel()
		case <-ctx.Done():
		}
	}()

	var eventsChan <-chan ipc.Event
	var errChan <-chan error

	if usingDaemon {
		eventsChan, errChan = StreamChatViaSSE(ctx, serverURL, sessionID, prompt)
	} else {
		pythonBin := cfg.Engine.PythonBin
		if pythonBin == "python3" {
			localVenv := filepath.Join(".venv", "bin", "python3")
			if _, err := os.Stat(localVenv); err == nil {
				pythonBin = localVenv
			}
		}

		wd, _ := os.Getwd()
		runnerParams := ipc.RunnerParams{
			PythonBin: pythonBin,
			WorkDir:   wd,
			DBPath:    cfg.Storage.DBPath,
			Prompt:    prompt,
			SessionID: sessionID,
			Offline:   cfg.Preferences.OfflineMode,
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

	fmt.Print("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true).Render(fmt.Sprintf("  ⠋ [%s] Menginisialisasi analisis & merencanakan investigasi...", modelLabel)) + "\r")

	activeErrChan := errChan
	for {
		select {
		case <-ctx.Done():
			if interrupted {
				fmt.Print("\r\033[K")
				return
			}

		case err, ok := <-activeErrChan:
			if !ok {
				activeErrChan = nil
				continue
			}
			if err != nil {
				fmt.Print("\r\033[K")
				fmt.Printf("\n[Error Subprocess]: %v\n", err)
				return
			}

		case ev, ok := <-eventsChan:
			if !ok {
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
				fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true).Render(fmt.Sprintf("  ⠋ [%s] Menjalankan verifikasi alat & analisis kuantitatif...", modelLabel)) + "\r")

			case ipc.EventAgentToolCall:
				fmt.Print("\r\033[K")
				argsJSON := ""
				if ev.Args != nil {
					argsJSON = fmt.Sprintf(" %v", ev.Args)
				}
				fmt.Printf("⚡ %s%s\n", toolCallStyle.Render("[TOOL CALL: "+ev.Tool+"]"), argsJSON)
				fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#FACC15")).Italic(true).Render("  ⠋ Mengeksekusi alat "+ev.Tool+"...") + "\r")

			case ipc.EventAgentObservation:
				fmt.Print("\r\033[K")
				fmt.Printf("🔎 %s\n", observationStyle.Render(ev.Summary))
				fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true).Render(fmt.Sprintf("  ⠋ [%s] Menyintesis temuan & menyusun respons...", modelLabel)) + "\r")

			case ipc.EventAnomalyDetected:
				fmt.Print("\r\033[K")
				totalAnomalies++
				anomalyText := fmt.Sprintf(
					"🚨 [ANOMALI TERDETEKSI] %s | Ticker: %s | Z-Score: %.2fσ | Metric: %.2f (Baseline: %.2f)",
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
				fmt.Print("\r\033[K" + lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Italic(true).Render(fmt.Sprintf("  ⠋ [%s] Menyusun sintesis respons (%d kata)...", modelLabel, words)) + "\r")

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
						BorderForeground(lipgloss.Color("#EF4444")).
						Padding(0, 1).
						Foreground(lipgloss.Color("#FCA5A5")).
						Render(fmt.Sprintf("❌ [ERROR SESSION]: %s", ev.Error))
					assistantResponse.WriteString(errBox)
				}
			}
		}
	}
}

func renderCompletionBadge(duration time.Duration, sessionID, model string, anomalies, findings int) string {
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#1F5C3F")).Render("─────────────────────────────────────────────────────────────────────────────")
	badge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22C55E")).Render("✔ [SELESAI]")
	detail := fmt.Sprintf("Analisis tuntas dalam %.1fs • Model: %s • Sesi: %s", duration.Seconds(), model, sessionID)
	if anomalies > 0 || findings > 0 {
		detail += fmt.Sprintf(" • (%d Anomali, %d Temuan)", anomalies, findings)
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
	fmt.Println(T("help_sessions_desc"))
	fmt.Println(T("help_web_desc"))
	fmt.Println(T("help_health_desc"))
	fmt.Println(T("help_lang_desc"))
	fmt.Println(T("help_clear_desc"))
	fmt.Println(T("help_exit_desc"))
}

func printHealth(cfg *config.Config) {
	fmt.Println(T("health_header"))
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("• Database Path  : %s\n", cfg.Storage.DBPath)
	fmt.Printf("• Python Runtime : %s\n", cfg.Engine.PythonBin)

	secKeyStatus := T("health_installed")
	if cfg.Auth.SectorsAPIKey == "" {
		secKeyStatus = T("health_not_installed")
	}
	fmt.Printf("• Sectors API Key: %s\n", secKeyStatus)

	activeModel := cfg.Auth.OpenAIModel
	if activeModel == "" {
		if cfg.Auth.GeminiModel != "" {
			activeModel = cfg.Auth.GeminiModel
		} else {
			activeModel = "hermes"
		}
	}

	baseURL := cfg.Auth.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "OpenAI-Compatible Standard"
	}

	keyStatus := T("health_installed")
	if cfg.Auth.OpenAIAPIKey == "" && cfg.Auth.GeminiAPIKey == "" {
		keyStatus = T("health_not_installed")
	}

	fmt.Printf("• Inference Engine: Universal ReAct (%s)\n", baseURL)
	fmt.Printf("• Active Model   : %s\n", activeModel)
	fmt.Printf("• Model API Key  : %s\n", keyStatus)
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
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
