package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/server"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Terminal Color Styles (Bloomberg / Cyber-OSINT Aesthetic)
	bannerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00E5FF")).
				Background(lipgloss.Color("#0F172A")).
				Padding(0, 2)

	bannerSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	promptBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E5FF"))

	userBubbleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#38BDF8")).
			Foreground(lipgloss.Color("#F8FAFC")).
			Padding(0, 1).
			MarginTop(1)

	thoughtStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B")).
			Italic(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	observationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))

	replAnomalyBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Foreground(lipgloss.Color("#FCA5A5")).
			Padding(0, 1).
			MarginTop(1)

	supportedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#10B981")).
				Padding(0, 1)

	uncertainBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#F59E0B")).
				Padding(0, 1)

	contradictedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#EF4444")).
				Padding(0, 1)
)

// RunLiveREPL starts an interactive, conversational research assistant session.
func RunLiveREPL(cfg *config.Config, appDB *db.DB, serverURL string) {
	reader := bufio.NewReader(os.Stdin)

	// Determine active model display
	modelLabel := "hermes"
	if cfg.Auth.AIProvider == "openai" && cfg.Auth.OpenAIModel != "" {
		modelLabel = cfg.Auth.OpenAIModel
	} else if cfg.Auth.AIProvider == "gemini" && cfg.Auth.GeminiModel != "" {
		modelLabel = cfg.Auth.GeminiModel
	}

	sessionID := fmt.Sprintf("CHAT-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)

	renderBanner(modelLabel, serverURL, sessionID)

	promptPrefix := fmt.Sprintf("niskava [%s] >", modelLabel)

	for {
		fmt.Printf("\n%s ", promptBoxStyle.Render(promptPrefix))
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nKeluar dari sesi Live Assistant.")
			break
		}

		input = strings.TrimSpace(input)
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
			renderBanner(modelLabel, serverURL, sessionID)
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

		// Check if input is just a ticker code (e.g. ANTM) and formulate prompt
		userPrompt := input
		cleanInput := strings.ToUpper(input)
		if len(cleanInput) >= 4 && len(cleanInput) <= 5 && !strings.Contains(cleanInput, " ") {
			userPrompt = fmt.Sprintf("Investigasi anomali transaksi kuantitatif dan keterbukaan informasi bursa untuk emiten %s selama 30 hari terakhir.", cleanInput)
		}

		// Execute conversational research turn
		executeChatTurn(userPrompt, sessionID, cfg, appDB)
	}
}

func renderBanner(modelLabel, serverURL, sessionID string) {
	fmt.Println()
	fmt.Println("=============================================================================")
	fmt.Printf(" %s\n", bannerTitleStyle.Render("NISKAVA AGENT — CONVERSATIONAL FINANCIAL OSINT ASSISTANT"))
	fmt.Printf(" %s · Daemon: %s · Sesi: %s\n",
		bannerSubStyle.Render("Model: "+modelLabel),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Render(serverURL),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7")).Render(sessionID),
	)
	fmt.Println(" Tanyakan apa saja mengenai pasar saham IDX, anomali transaksi, atau katalis berita.")
	fmt.Println(" Ketik /help untuk melihat perintah utilitas sistem.")
	fmt.Println("=============================================================================")
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

	// Display User Card
	fmt.Printf("\n%s\n", userBubbleStyle.Render("👤 "+prompt))

	pythonBin := cfg.Engine.PythonBin
	if pythonBin == "python3" {
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
