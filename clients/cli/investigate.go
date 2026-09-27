package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	daysFlag         int
	offlineFlag      bool
	interactiveFlag  bool
	pyBinFlag        string
	enginePath       string
	invExportFmtFlag string
	invExportOutFlag string
)

var investigateCmd = &cobra.Command{
	Use:   "investigate [TICKER]",
	Short: "Run autonomous investigation on an IDX ticker (e.g. ANTM)",
	Long: `Run an autonomous 7-stage investigation pipeline on an IDX ticker (e.g. ANTM, BBCA).
Executes quantitative anomaly calculations, harvests contemporaneous news/disclosures,
and compiles evidence classified into SUPPORTED, UNCERTAIN, or CONTRADICTED findings.`,
	Example: `  # Run 30-day headless investigation:
  niskava investigate ANTM

  # Run investigation and export report to Markdown:
  niskava investigate ANTM --days 30 --export-format md --export-out ANTM_Report.md

  # Run investigation and open interactive REPL pre-focused on ticker:
  niskava investigate ANTM -i`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := strings.ToUpper(strings.TrimSpace(args[0]))
		if len(ticker) < 4 || len(ticker) > 5 {
			return fmt.Errorf("invalid IDX ticker '%s': IDX tickers are 4-5 letters (e.g. ANTM, BBCA)", ticker)
		}

		sessionID := fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), strings.ToUpper(ticker))

		// Initialize session in local SQLite database
		inv := &db.Investigation{
			ID:            sessionID,
			Ticker:        ticker,
			Market:        cfg.Preferences.DefaultMarket,
			TimeframeDays: daysFlag,
			Status:        "RUNNING",
			StartedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		if err := appDB.CreateInvestigation(inv); err != nil {
			// Non-fatal if session already exists, continue
		}

		// Determine Python binary: CLI flag > Config file / Env > Virtual env fallback
		pythonBin := pyBinFlag
		if pythonBin == "" && cfg != nil {
			pythonBin = cfg.Engine.PythonBin
		}
		pythonBin = ipc.ResolvePythonBin(pythonBin)

		isOffline := offlineFlag || cfg.Preferences.OfflineMode

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wd, _ := os.Getwd()
		resolvedRoot := ipc.ResolveRepoRoot(wd)

		customEngine := enginePath
		if customEngine == "" && cfg != nil {
			customEngine = cfg.Engine.EnginePath
		}
		resolvedEngine := ipc.ResolveEnginePath(resolvedRoot, customEngine)

		runnerParams := ipc.RunnerParams{
			PythonBin:    pythonBin,
			WorkDir:      resolvedRoot,
			EnginePath:   resolvedEngine,
			DBPath:       cfg.Storage.DBPath,
			Ticker:       ticker,
			Days:         daysFlag,
			SessionID:    sessionID,
			Offline:      isOffline,
			Language:     cfg.Preferences.Language,
			EnvOverrides: cfg.BuildSubprocessEnv(),
		}

		// If --interactive / -i flag is set, launch Live REPL immediately pre-focused on target ticker
		if interactiveFlag {
			srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
			if err != nil {
				return fmt.Errorf("failed to start background daemon for interactive mode: %w", err)
			}
			srv.ConfigPath = cfgFile
			initialPrompt := fmt.Sprintf("Lakukan investigasi anomali volume dan verifikasi bukti untuk saham %s", ticker)
			_ = tui.RunLiveREPLWithInitialPrompt(cfg, appDB, srv.URL, sessionID, initialPrompt)
			return nil
		}

		eventsChan, errChan := ipc.RunSubprocess(ctx, runnerParams)

		model := tui.NewModel(ticker, daysFlag, cfg.Storage.DBPath, eventsChan, errChan)
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running interactive TUI: %w", err)
		}

		// Export report if export flag was set
		if invExportFmtFlag != "" || invExportOutFlag != "" {
			exportFmt := strings.ToLower(strings.TrimSpace(invExportFmtFlag))
			if exportFmt == "" {
				exportFmt = "md"
			}
			outPath := invExportOutFlag
			if outPath == "" {
				outPath = fmt.Sprintf("niskava_investigation_%s.%s", sessionID, exportFmt)
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# Niskava Agent — Audit & Investigation Report (%s)\n\n", ticker))
			sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", sessionID))
			sb.WriteString(fmt.Sprintf("- **Ticker:** `%s`\n", ticker))
			sb.WriteString(fmt.Sprintf("- **Date:** `%s`\n\n---\n\n", time.Now().Format("2006-01-02 15:04:05 MST")))

			if invData, errInv := appDB.GetInvestigation(sessionID); errInv == nil && invData != nil {
				if invData.SummaryText != nil && *invData.SummaryText != "" {
					sb.WriteString(fmt.Sprintf("## ⚡ Executive Summary\n%s\n\n---\n\n", *invData.SummaryText))
				}
			}

			sb.WriteString(tui.T("sessions_export_disclaimer"))
			_ = os.WriteFile(outPath, []byte(sb.String()), 0644)
			fmt.Printf("\n✓ Investigation report exported to: %s\n", outPath)
		}

		// Option A: If running in an interactive terminal, offer CTA to transition into Live REPL
		if isTerminalInput() {
			fmt.Printf("\n [Enter / y] Lanjutkan diskusi interaktif untuk emiten %s? (y/N): ", ticker)
			var resp string
			_, _ = fmt.Scanln(&resp)
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp == "" || resp == "y" || resp == "yes" {
				srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
				if err == nil {
					srv.ConfigPath = cfgFile
					initialPrompt := fmt.Sprintf("Berdasarkan hasil audit %s yang baru saja dilakukan, analisis temuan dan berita lebih lanjut.", ticker)
					_ = tui.RunLiveREPLWithInitialPrompt(cfg, appDB, srv.URL, sessionID, initialPrompt)
				}
			}
		}

		return nil
	},
}

func isTerminalInput() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func init() {
	investigateCmd.Flags().IntVarP(&daysFlag, "days", "d", 30, "observation window in days (30, 60, or 90)")
	investigateCmd.Flags().BoolVar(&offlineFlag, "offline", false, "run in offline mock mode without calling remote APIs")
	investigateCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "run in interactive conversational investigation mode")
	investigateCmd.Flags().StringVar(&pyBinFlag, "python-bin", "", "path to python binary")
	investigateCmd.Flags().StringVar(&enginePath, "engine-path", "", "path to python engine directory")
	investigateCmd.Flags().StringVarP(&invExportFmtFlag, "export-format", "f", "", "export report format: 'md' or 'json'")
	investigateCmd.Flags().StringVarP(&invExportOutFlag, "export-out", "o", "", "output report file path (e.g. report.md)")

	investigateCmd.ValidArgs = []string{
		"BBCA", "BBRI", "BMRI", "BBNI", "TLKM",
		"ANTM", "ASII", "ICBP", "INDF", "GOTO",
		"ADRO", "PTBA", "UNTR", "BRIS", "AMMN",
		"KLBF", "MDKA", "TPIA", "CPIN", "PGAS",
	}

	RootCmd.AddCommand(investigateCmd)
}
