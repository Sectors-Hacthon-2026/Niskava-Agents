package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/server"
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

var defaultSlashCommands = []SlashCommand{
	{Command: "/help", Description: "Panduan lengkap perintah & instruksi sistem"},
	{Command: "/reset", Description: "Mulai sesi obrolan baru & bersihkan memory graph"},
	{Command: "/clear", Description: "Bersihkan layar terminal & tampilkan ulang banner HUD"},
	{Command: "/web", Description: "Buka dashboard visual Web Workspace di browser"},
	{Command: "/sessions", Description: "Inspeksi riwayat sesi investigasi & audit trail dari SQLite"},
	{Command: "/health", Description: "Periksa status daemon server, database, & provider AI"},
	{Command: "/exit", Description: "Keluar dari sesi Live REPL kembali ke menu utama"},
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
	ti.Placeholder = "Ketik pertanyaan riset pasar atau / untuk perintah..."
	ti.Focus()

	return ReplInputModel{
		TextInput:        ti,
		PromptPrefix:     promptPrefix,
		SlashCommands:    defaultSlashCommands,
		FilteredCommands: defaultSlashCommands,
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
	modelLabel := "hermes"
	if cfg.Auth.AIProvider == "openai" && cfg.Auth.OpenAIModel != "" {
		modelLabel = cfg.Auth.OpenAIModel
	} else if cfg.Auth.AIProvider == "gemini" && cfg.Auth.GeminiModel != "" {
		modelLabel = cfg.Auth.GeminiModel
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
			sessionID = fmt.Sprintf("CHAT-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
			fmt.Printf("\n[✓] Sesi direset. Sesi percakapan baru: %s\n", sessionID)
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

		if lower == "/sessions" {
			printSessions(appDB)
			continue
		}

		// Execute conversational research turn with verbatim user prompt
		executeChatTurn(input, sessionID, cfg, appDB)
	}
}

func renderBanner(modelLabel, serverURL, sessionID, dbPath string) {
	fmt.Print(RenderHUDHeader(modelLabel, serverURL, dbPath, sessionID))
	helpHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("  [PETUNJUK: Ketik /help untuk panduan perintah, /reset untuk reset chat, /clear untuk bersihkan layar, /exit untuk keluar]")
	fmt.Printf("\n%s\n", helpHint)
}

func executeChatTurn(prompt, sessionID string, cfg *config.Config, appDB *db.DB) {
	// 1. Record User Message in SQLite
	userMsg := &db.ChatMessage{
		ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
		SessionID: sessionID,
		Role:      "user",
		Content:   prompt,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = appDB.SaveChatMessage(userMsg)

	// Display User Card (Pure Text, No Emoji Icon)
	fmt.Printf("\n%s\n", userBubbleStyle.Render("USER > "+prompt))

	pythonBin := cfg.Engine.PythonBin
	if pythonBin == "python3" && runtime.GOOS == "windows" {
		pythonBin = "python"
	}
	if runtime.GOOS == "windows" {
		localVenv := filepath.Join(".venv", "Scripts", "python.exe")
		if _, err := os.Stat(localVenv); err == nil {
			pythonBin = localVenv
		}
	} else if pythonBin == "python3" {
		localVenv := filepath.Join(".venv", "bin", "python3")
		if _, err := os.Stat(localVenv); err == nil {
			pythonBin = localVenv
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wd, _ := os.Getwd()

	runnerParams := ipc.RunnerParams{
		PythonBin: pythonBin,
		WorkDir:   wd,
		DBPath:    cfg.Storage.DBPath,
		Prompt:    prompt,
		SessionID: sessionID,
		Offline:   cfg.Preferences.OfflineMode,
	}

	eventsChan, errChan := ipc.RunSubprocess(ctx, runnerParams)

	var (
		assistantResponse strings.Builder
		lastThought       string
	)

	fmt.Println()

	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				fmt.Printf("\n[Error Subprocess]: %v\n", err)
			}
			return

		case ev, ok := <-eventsChan:
			if !ok {
				// Process finished
				renderFinalMarkdown(assistantResponse.String())

				// Save assistant response in SQLite
				if assistantResponse.Len() > 0 {
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
				return
			}

			switch ev.Event {
			case ipc.EventAgentThought:
				lastThought = ev.Thought
				fmt.Printf("💭 %s\n", thoughtStyle.Render(ev.Thought))

			case ipc.EventAgentToolCall:
				argsJSON := ""
				if ev.Args != nil {
					argsJSON = fmt.Sprintf(" %v", ev.Args)
				}
				fmt.Printf("⚡ %s%s\n", toolCallStyle.Render("[TOOL CALL: "+ev.Tool+"]"), argsJSON)

			case ipc.EventAgentObservation:
				fmt.Printf("🔎 %s\n", observationStyle.Render(ev.Summary))

			case ipc.EventAnomalyDetected:
				anomalyText := fmt.Sprintf(
					"🚨 [ANOMALI TERDETEKSI] %s | Ticker: %s | Z-Score: %.2fσ | Metric: %.2f (Baseline: %.2f)",
					ev.MetricType, ev.Ticker, ev.ZScore, ev.MetricValue, ev.BaselineValue,
				)
				fmt.Println(replAnomalyBoxStyle.Render(anomalyText))

			case ipc.EventFindingEmitted:
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

			case ipc.EventAgentMessageComplete:
				if assistantResponse.Len() == 0 {
					assistantResponse.WriteString(ev.Content)
				}

			case ipc.EventSessionError:
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
	fmt.Println("\nDAFTAR PERINTAH NISKAVA LIVE ASSISTANT:")
	fmt.Println("  <PROMPT BEBAS>       Tanyakan pertanyaan riset pasar saham (contoh: 'Kenapa saham ANTM naik kemarin?')")
	fmt.Println("  <KODE EMITEN>        Ketik langsung 4 huruf kode emiten untuk analisis cepat (contoh: ANTM, BBCA, BUMI)")
	fmt.Println("  /reset               Mulai sesi percakapan baru (bersihkan konteks obrolan)")
	fmt.Println("  /sessions            Lihat riwayat sesi investigasi & audit trail dari SQLite lokal")
	fmt.Println("  /web                 Buka dashboard visual Web Workspace di browser")
	fmt.Println("  /health              Periksa status database, API keys, dan provider AI")
	fmt.Println("  /clear               Bersihkan layar terminal")
	fmt.Println("  /exit, quit          Keluar dari sesi REPL")
}

func printHealth(cfg *config.Config) {
	fmt.Println("\nSTATUS KESEHATAN SISTEM:")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("• Database Path  : %s\n", cfg.Storage.DBPath)
	fmt.Printf("• Python Runtime : %s\n", cfg.Engine.PythonBin)

	secKeyStatus := "Terpasang (Live Ready)"
	if cfg.Auth.SectorsAPIKey == "" {
		secKeyStatus = "Belum Terpasang (Mode Offline Aktif)"
	}
	fmt.Printf("• Sectors API Key: %s\n", secKeyStatus)

	if cfg.Auth.AIProvider == "openai" || cfg.Auth.OpenAIAPIKey != "" {
		providerName := "9router / OpenAI Compatible"
		if cfg.Auth.OpenAIBaseURL != "" {
			providerName = fmt.Sprintf("9router (%s)", cfg.Auth.OpenAIBaseURL)
		}
		fmt.Printf("• AI Provider    : %s\n", providerName)
		fmt.Printf("• Active Model   : %s\n", cfg.Auth.OpenAIModel)
		fmt.Printf("• Model API Key  : Terpasang (Live Ready)\n")
	} else {
		gemKeyStatus := "Terpasang"
		if cfg.Auth.GeminiAPIKey == "" {
			gemKeyStatus = "Belum Terpasang (Simulasi Cerdas Aktif)"
		}
		fmt.Printf("• AI Provider    : Google Gemini (%s)\n", cfg.Auth.GeminiModel)
		fmt.Printf("• Gemini API Key : %s\n", gemKeyStatus)
	}
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
}

func printSessions(appDB *db.DB) {
	investigations, err := appDB.ListInvestigations(20)
	if err != nil {
		fmt.Printf("Gagal membaca database: %v\n", err)
		return
	}

	if len(investigations) == 0 {
		fmt.Println("Belum ada sesi investigasi tersimpan.")
		return
	}

	fmt.Println("\nRIWAYAT SESI INVESTIGASI TERSIMPAN (SQLITE):")
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
