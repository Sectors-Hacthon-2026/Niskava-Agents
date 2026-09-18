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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunLiveREPL starts a persistent, interactive terminal session.
func RunLiveREPL(cfg *config.Config, appDB *db.DB, serverURL string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=============================================================================")
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5FF")).Render(" NISKAVA LIVE CLI — INTERACTIVE OSINT INVESTIGATOR"))
	fmt.Println(" Ketik kode saham IDX (contoh: ANTM, BBCA, GOTO, BUMI) atau /help untuk bantuan.")
	fmt.Printf(" Server Web aktif di: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Render(serverURL))
	fmt.Println("=============================================================================")

	promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5FF"))

	for {
		fmt.Printf("\n%s ", promptStyle.Render("niskava [LIVE] >"))
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nKeluar dari sesi live.")
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

		// Extract target ticker
		targetTicker := input
		if strings.HasPrefix(lower, "/investigate ") {
			targetTicker = strings.TrimSpace(input[len("/investigate "):])
		}

		targetTicker = strings.ToUpper(targetTicker)
		if len(targetTicker) < 4 || len(targetTicker) > 5 {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Ticker tidak valid. Masukkan 4-5 huruf kode saham BEI (contoh: ANTM, BBRI)."))
			continue
		}

		// Execute live investigation within the REPL
		executeInvestigation(targetTicker, cfg, appDB)
	}
}

func executeInvestigation(ticker string, cfg *config.Config, appDB *db.DB) {
	sessionID := fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), ticker)

	// Save session record
	inv := &db.Investigation{
		ID:            sessionID,
		Ticker:        ticker,
		Market:        cfg.Preferences.DefaultMarket,
		TimeframeDays: 30,
		Status:        "RUNNING",
		StartedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	_ = appDB.CreateInvestigation(inv)

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
		Ticker:    ticker,
		Days:      30,
		SessionID: sessionID,
		Offline:   cfg.Preferences.OfflineMode,
	}

	eventsChan, errChan := ipc.RunSubprocess(ctx, runnerParams)

	model := NewModel(ticker, 30, cfg.Storage.DBPath, eventsChan, errChan)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running investigation: %v\n", err)
	}
}

func printHelp() {
	fmt.Println("\nDAFTAR PERINTAH NISKAVA LIVE REPL:")
	fmt.Println("  <TICKER>             Langsung jalankan investigasi saham (contoh: ANTM, BBCA, BUMI)")
	fmt.Println("  /investigate <CODE>  Perintah eksplisit investigasi emiten")
	fmt.Println("  /sessions            Lihat riwayat sesi investigasi dari database SQLite lokal")
	fmt.Println("  /web                 Buka dashboard visual Web Workspace di browser")
	fmt.Println("  /health              Periksa status database, API keys, dan lingkungan Python")
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

	gemKeyStatus := "Terpasang"
	if cfg.Auth.GeminiAPIKey == "" {
		gemKeyStatus = "Belum Terpasang (Simulasi Cerdas Aktif)"
	}
	fmt.Printf("• Gemini API Key : %s\n", gemKeyStatus)
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
	fmt.Printf("%-20s %-8s %-12s %-20s %s\n", "SESSION ID", "TICKER", "STATUS", "STARTED AT", "SUMMARY")
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

		fmt.Printf("%-20s %-8s %-12s %-20s %s\n", inv.ID, inv.Ticker, inv.Status, dateStr, summary)
	}
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
}
