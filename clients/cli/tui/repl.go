package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	// Terminal Color Styles (Binance Dark Financial Intelligence Aesthetic)
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

// ParseTimeoutCommand parses "/timeout <arg>" and returns the resolved timeout in seconds.
// Named profiles: fast=25, balanced=60, deep=120, local=180.
// Numeric: accepted in [10, 300] inclusive.
// Returns (0, false) when arg is absent (caller should display current) or invalid.
func ParseTimeoutCommand(input string) (float64, bool) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return 0, false // no arg → show current value
	}
	arg := strings.ToLower(strings.TrimSpace(parts[1]))
	switch arg {
	case "fast":
		return 25.0, true
	case "balanced":
		return 60.0, true
	case "deep":
		return 120.0, true
	case "local":
		return 180.0, true
	}
	var secs float64
	if _, err := fmt.Sscanf(arg, "%f", &secs); err != nil {
		return 0, false
	}
	if secs < 10.0 || secs > 300.0 {
		return 0, false
	}
	return secs, true
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
	sessID := ""
	if len(initialSessionID) > 0 {
		sessID = initialSessionID[0]
	}
	return RunLiveREPLWithInitialPrompt(cfg, appDB, serverURL, sessID, "")
}

// RunLiveREPLWithInitialPrompt starts an interactive REPL pre-seeded with an initial prompt.
func RunLiveREPLWithInitialPrompt(cfg *config.Config, appDB *db.DB, serverURL string, initialSessionID string, initialPrompt string) string {
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
	if strings.TrimSpace(initialSessionID) != "" {
		sessionID = strings.TrimSpace(initialSessionID)
	}

	renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)

	if strings.TrimSpace(initialSessionID) != "" {
		renderResumedHistory(appDB, sessionID)
	}

	promptPrefix := fmt.Sprintf("niskava [%s] >", modelLabel)

	var promptHistory []string

	if strings.TrimSpace(initialPrompt) != "" {
		promptHistory = append(promptHistory, initialPrompt)
		executeChatTurn(initialPrompt, sessionID, serverURL, cfg, appDB)
	}

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

		if lower == "/doctor" {
			printHealth(cfg)
			continue
		}

		if strings.HasPrefix(lower, "/export") {
			parts := strings.Fields(input)
			format := "md"
			if len(parts) > 1 && strings.ToLower(parts[1]) == "json" {
				format = "json"
			}
			filename := fmt.Sprintf("niskava_report_%s.%s", sessionID, format)
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("slash_export_db_err")))
				continue
			}
			history, errH := appDB.GetChatHistory(sessionID, 100)
			if errH != nil || len(history) == 0 {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(TF("slash_export_empty", sessionID)))
				continue
			}
			var content string
			if format == "json" {
				var jsonMsgs []map[string]interface{}
				for _, m := range history {
					jsonMsgs = append(jsonMsgs, map[string]interface{}{
						"id":         m.ID,
						"role":       m.Role,
						"content":    m.Content,
						"created_at": m.CreatedAt,
					})
				}
				data, _ := json.MarshalIndent(map[string]interface{}{
					"session_id": sessionID,
					"messages":   jsonMsgs,
				}, "", "  ")
				content = string(data)
			} else {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%s\n\n", T("slash_export_report_title")))
				sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
				sb.WriteString(fmt.Sprintf("- **Date:** `%s`\n", time.Now().Format("2006-01-02 15:04:05 MST")))
				sb.WriteString(fmt.Sprintf("- **Model:** `%s`\n\n---\n\n", modelLabel))
				for _, m := range history {
					if m.Role == "user" {
						sb.WriteString(fmt.Sprintf("%s\n> %s\n\n", T("slash_export_user_prompt"), m.Content))
					} else if m.Role == "assistant" {
						sb.WriteString(fmt.Sprintf("%s\n%s\n\n---\n\n", T("slash_export_findings"), m.Content))
					}
				}
				sb.WriteString(T("sessions_export_disclaimer"))
				content = sb.String()
			}
			if errW := os.WriteFile(filename, []byte(content), 0644); errW != nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(TF("slash_export_write_err", errW)))
			} else {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(TF("sessions_export_success", filename)))
			}
			continue
		}

		if strings.HasPrefix(lower, "/fork") {
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("slash_fork_db_err")))
				continue
			}
			parts := strings.SplitN(input, " ", 2)
			newTitle := T("slash_fork_default_title")
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				newTitle = strings.TrimSpace(parts[1])
			}
			newSessionID := fmt.Sprintf("CHAT-FORK-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
			errFork := appDB.ForkChatSession(sessionID, newSessionID, newTitle, "")
			if errFork != nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(TF("slash_fork_err", errFork)))
				continue
			}
			prevID := sessionID
			sessionID = newSessionID
			fmt.Print("\033[H\033[2J")
			renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
			fmt.Println(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(TF("slash_fork_success", prevID, sessionID, newTitle)))
			renderResumedHistory(appDB, sessionID)
			continue
		}

		if strings.HasPrefix(lower, "/search") {
			parts := strings.SplitN(input, " ", 2)
			if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(T("slash_search_usage")))
				continue
			}
			kw := strings.TrimSpace(parts[1])
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("repl_db_unavailable")))
				continue
			}
			results, errSearch := appDB.SearchChatMessages(kw, 10)
			if errSearch != nil || len(results) == 0 {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(TF("sessions_search_empty", kw)))
				continue
			}
			fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(TF("slash_search_title_repl", kw)))
			for idx, r := range results {
				snip := r.Content
				if len(snip) > 100 {
					snip = snip[:97] + "..."
				}
				fmt.Printf("  %d. [%s] [%s] %s: %s\n", idx+1, r.SessionID, r.CreatedAt, r.Role, snip)
			}
			fmt.Println()
			continue
		}

		if lower == "/anomalies" {
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("repl_db_unavailable")))
				continue
			}
			anomalies, errA := appDB.GetAnomaliesByInvestigation(sessionID)
			if errA != nil || len(anomalies) == 0 {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(T("slash_anomalies_empty")))
				continue
			}
			fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(T("slash_anomalies_title")))
			for idx, a := range anomalies {
				fmt.Println(TF("slash_anomalies_item", idx+1, a.AnomalyDate, a.MetricType, a.MetricValue, a.BaselineValue, a.ZScore, a.Description))
			}
			fmt.Println()
			continue
		}

		if lower == "/skills" {
			fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(T("slash_skills_title")))
			skills := []struct{ Name, SOP, DescKey string }{
				{"market_anomaly_recon", "SOP-01", "skill_desc_sop01"},
				{"event_causality_audit", "SOP-02", "skill_desc_sop02"},
				{"insider_bandarmology", "SOP-03", "skill_desc_sop03"},
				{"financial_health_stress", "SOP-04", "skill_desc_sop04"},
				{"mining_commodity_divergence", "SOP-05", "skill_desc_sop05"},
				{"peer_valuation_benchmark", "SOP-06", "skill_desc_sop06"},
			}
			for _, s := range skills {
				fmt.Printf("  • %-28s [%s] : %s\n", lipgloss.NewStyle().Bold(true).Foreground(ColorThought).Render(s.Name), s.SOP, T(s.DescKey))
			}
			fmt.Println()
			continue
		}

		if strings.HasPrefix(lower, "/cache") {
			if appDB == nil {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(T("repl_db_unavailable")))
				continue
			}
			parts := strings.Fields(input)
			if len(parts) > 1 && strings.ToLower(parts[1]) == "clean" {
				n, errC := appDB.CleanExpiredCache()
				if errC != nil {
					fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(TF("slash_cache_clean_err", errC)))
				} else {
					fmt.Println(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(TF("slash_cache_clean_success", n)))
				}
			} else {
				stats, errS := appDB.GetSectorsCacheStats()
				if errS != nil {
					fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(TF("slash_cache_stats_err", errS)))
				} else {
					fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(T("slash_cache_stats_title")))
					fmt.Printf("%s\n", TF("slash_cache_stats_total", stats.TotalEntries))
					fmt.Printf("%s\n", TF("slash_cache_stats_perm", stats.PermanentEntries))
					fmt.Printf("%s\n", TF("slash_cache_stats_expired", stats.ExpiredEntries))
					fmt.Println(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("slash_cache_stats_hint")))
				}
			}
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
			_ = config.SaveConfig(cfg)

			fmt.Print("\033[H\033[2J")
			renderBanner(modelLabel, serverURL, sessionID, cfg.Storage.DBPath)
			activeInfo := GetActiveLanguageInfo()
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(TF("repl_lang_switched", activeInfo.FlagSymbol, activeInfo.NativeName, activeInfo.Code)))
			continue
		}

		if strings.HasPrefix(lower, "/timeout") {
			secs, ok := ParseTimeoutCommand(input)
			if !ok && len(strings.Fields(input)) < 2 {
				current := cfg.Preferences.LLMTimeoutSecs
				if current <= 0 {
					current = 60.0
				}
				fmt.Println(lipgloss.NewStyle().Foreground(ColorAccent).Render(
					fmt.Sprintf(T("slash_timeout_current"), current),
				))
				continue
			}
			if !ok {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorWarning).Render(T("slash_timeout_invalid")))
				continue
			}
			cfg.Preferences.LLMTimeoutSecs = secs
			_ = config.SaveConfig(cfg)
			profile := "custom"
			switch secs {
			case 25.0:
				profile = "fast"
			case 60.0:
				profile = "balanced"
			case 120.0:
				profile = "deep"
			case 180.0:
				profile = "local"
			}
			fmt.Println(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(
				fmt.Sprintf(T("slash_timeout_set"), secs, profile),
			))
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
	modelLabel := "hermes"
	if cfg != nil {
		if cfg.Auth.OpenAIModel != "" {
			modelLabel = cfg.Auth.OpenAIModel
		} else if cfg.Auth.GeminiModel != "" {
			modelLabel = cfg.Auth.GeminiModel
		}
	}

	// Record User Message and ensure ChatSession metadata exists in SQLite if running standalone subprocess mode
	if !usingDaemon && appDB != nil {
		_, errSess := appDB.GetChatSession(sessionID)
		if errSess != nil {
			title := prompt
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			sess := &db.ChatSession{
				ID:        sessionID,
				Title:     title,
				Model:     modelLabel,
				Status:    "IDLE",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			}
			_ = appDB.CreateChatSession(sess)
		}

		userMsg := &db.ChatMessage{
			ID:        fmt.Sprintf("MSG-%d", time.Now().UnixNano()),
			SessionID: sessionID,
			Role:      "user",
			Content:   prompt,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		_ = appDB.SaveChatMessage(userMsg)
		_ = appDB.TouchChatSession(sessionID, prompt)
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
					if usingDaemon {
						go func(sURL, sID string) {
							abortURL := fmt.Sprintf("%s/api/chat/sessions/%s/abort", sURL, sID)
							req, _ := http.NewRequest("POST", abortURL, nil)
							if req != nil {
								req.Header.Set("Content-Type", "application/json")
							}
							client := &http.Client{Timeout: 1500 * time.Millisecond}
							_, _ = client.Do(req)
						}(serverURL, sessionID)
					}
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
			PythonBin:    pythonBin,
			WorkDir:      wd,
			DBPath:       cfg.Storage.DBPath,
			Prompt:       prompt,
			SessionID:    sessionID,
			Offline:      cfg.Preferences.OfflineMode,
			Language:     cfg.Preferences.Language,
			EnvOverrides: cfg.BuildSubprocessEnv(),
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
					_ = appDB.TouchChatSession(sessionID, assistantResponse.String())
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
				vMsg := strings.TrimPrefix(TF("thinking_synthesize", modelLabel), "  ⠋ ")
				statusText.Store(vMsg)

			case ipc.EventAgentToolCall:
				fmt.Print("\r\033[K")
				argsJSON := ""
				if ev.Args != nil {
					argsJSON = fmt.Sprintf(" %v", ev.Args)
				}
				toolName := ev.Tool
				if toolName == "search_osint" {
					toolName = "search_news"
				}
				fmt.Printf("⚡ %s%s\n", toolCallStyle.Render("[TOOL CALL: "+toolName+"]"), argsJSON)
				tMsg := strings.TrimPrefix(TF("tool_executing", toolName), "  ⠋ ")
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
	fmt.Println("  • /timeout [arg]: " + T("slash_timeout_desc"))
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
