package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	daysFlag        int
	offlineFlag     bool
	interactiveFlag bool
	pyBinFlag       string
	enginePath      string
)

var investigateCmd = &cobra.Command{
	Use:   "investigate [TICKER]",
	Short: "Run autonomous investigation on an IDX ticker (e.g. ANTM)",
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

		eventsChan, errChan := ipc.RunSubprocess(ctx, runnerParams)

		model := tui.NewModel(ticker, daysFlag, cfg.Storage.DBPath, eventsChan, errChan)
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running interactive TUI: %w", err)
		}

		return nil
	},
}

func init() {
	investigateCmd.Flags().IntVarP(&daysFlag, "days", "d", 30, "observation window in days (30, 60, or 90)")
	investigateCmd.Flags().BoolVar(&offlineFlag, "offline", false, "run in offline mock mode without calling remote APIs")
	investigateCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "run in interactive conversational investigation mode")
	investigateCmd.Flags().StringVar(&pyBinFlag, "python-bin", "", "path to python binary")
	investigateCmd.Flags().StringVar(&enginePath, "engine-path", "", "path to python engine directory")

	investigateCmd.ValidArgs = []string{
		"BBCA", "BBRI", "BMRI", "BBNI", "TLKM",
		"ANTM", "ASII", "ICBP", "INDF", "GOTO",
		"ADRO", "PTBA", "UNTR", "BRIS", "AMMN",
		"KLBF", "MDKA", "TPIA", "CPIN", "PGAS",
	}

	RootCmd.AddCommand(investigateCmd)
}
