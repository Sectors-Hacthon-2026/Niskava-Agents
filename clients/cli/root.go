// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Version defines the release version of Niskava Agent.
const Version = "0.2.1"

var (
	cfgFile     string
	langFlag    string
	sessionFlag string
	verbose     bool
	cfg         *config.Config
	appDB       *db.DB
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "niskava",
	Short: "Niskava Agent — Autonomous IDX Market Intelligence Platform",
	Long: `Niskava Agent is an autonomous financial market intelligence and equity research
orchestration platform designed specifically for the Indonesia Stock Exchange (IDX).
Bridges the gap between quantitative market facts (Sectors Financial API v2)
and qualitative market disclosures/news.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if langFlag != "" {
			cfg.Preferences.Language = langFlag
		}
		tui.SetLanguage(cfg.Preferences.Language)

		appDB, err = db.Open(cfg.Storage.DBPath)
		if err != nil {
			return fmt.Errorf("failed to open database at %s: %w", cfg.Storage.DBPath, err)
		}

		// Commands that do not require a Sectors key:
		// - setup: configures the key interactively
		// - version, help, completion: informational only, no data access
		// - sessions: inspects/exports local SQLite history only, no remote calls
		exemptCommands := map[string]bool{
			"setup":      true,
			"version":    true,
			"help":       true,
			"completion": true,
			"sessions":   true,
		}
		topCmd := cmd
		for topCmd.HasParent() && topCmd.Parent().HasParent() {
			topCmd = topCmd.Parent()
		}
		if !exemptCommands[cmd.Name()] && !exemptCommands[topCmd.Name()] {
			if keyErr := config.RequireSectorsKey(cfg); keyErr != nil {
				return keyErr
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Bare command: start background daemon & launch 9router-style interface selector
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}
		srv.ConfigPath = cfgFile

		// If explicit --session flag provided, bypass launcher and jump directly into REPL
		if sessionFlag != "" {
			res := tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)
			if res == tui.ReplSetupSentinel {
				_ = RunInteractiveSetup()
			}
			return nil
		}

		hasAPIKey := cfg.Auth.SectorsAPIKey != "" || cfg.Auth.GeminiAPIKey != "" || cfg.Auth.OpenAIAPIKey != ""
		activeCursor := 0
		for {
			fmt.Print("\033[H\033[2J")
			launcher := tui.NewLauncherModelWithHealthAndCursor(srv.URL, "v"+Version, hasAPIKey, activeCursor)
			p := tea.NewProgram(launcher, tea.WithAltScreen())
			m, err := p.Run()
			if err != nil {
				return fmt.Errorf("launcher error: %w", err)
			}

			resModel := m.(tui.LauncherModel)
			activeCursor = resModel.Cursor
			selected := resModel.Selected
			switch selected {
			case "web":
				_ = server.OpenBrowser(srv.URL)
				tui.ShowWebWorkspaceLaunchScreen(srv.URL)

			case "terminal":
				// RunLiveREPL returns control to launcher menu when user exits or types /back or /exit
				res := tui.RunLiveREPL(cfg, appDB, srv.URL)
				if res == tui.ReplSetupSentinel {
					_ = RunInteractiveSetup()
					if newCfg, err := config.Load(cfgFile); err == nil {
						cfg = newCfg
						srv.Config = newCfg
						hasAPIKey = cfg.Auth.SectorsAPIKey != "" || cfg.Auth.GeminiAPIKey != "" || cfg.Auth.OpenAIAPIKey != ""
					}
				}

			case "sessions":
				// Show saved sessions with interactive resume option
				selectedSessionID := runSessionsInteractive(cmd, appDB)
				if selectedSessionID != "" {
					res := tui.RunLiveREPL(cfg, appDB, srv.URL, selectedSessionID)
					if res == tui.ReplSetupSentinel {
						_ = RunInteractiveSetup()
						if newCfg, err := config.Load(cfgFile); err == nil {
							cfg = newCfg
							srv.Config = newCfg
							hasAPIKey = cfg.Auth.SectorsAPIKey != "" || cfg.Auth.GeminiAPIKey != "" || cfg.Auth.OpenAIAPIKey != ""
						}
					}
				}

			case "help":
				tui.PrintFullHelpGuide()

			case "health":
				tui.ShowHealthDiagnosticsScreen(cfg, srv.URL)

			case "lang":
				langModel := tui.NewLangSelectorModel()
				pLang := tea.NewProgram(langModel, tea.WithAltScreen())
				mLang, errLang := pLang.Run()
				if errLang == nil {
					selLang := mLang.(tui.LangSelectorModel).Selected
					if selLang != "" {
						tui.SetLanguage(selLang)
						cfg.Preferences.Language = tui.ActiveLanguage
						_ = config.SaveConfig(cfg)
					}
				}

			case "setup":
				_ = RunInteractiveSetup()
				tui.PromptPressEscToReturn()

			case "exit", "":
				fmt.Println(tui.T("menu_exit_msg"))
				return nil
			}
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if appDB != nil {
			return appDB.Close()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current Niskava Agent version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("niskava version %s (%s/%s, %s)\n", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	},
}

func init() {
	RootCmd.Version = Version
	RootCmd.SetVersionTemplate("niskava version {{.Version}}\n")
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ~/.niskava/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&langFlag, "lang", "l", "", "language preference: 'en' for English (default) or 'id' for Indonesian")
	RootCmd.PersistentFlags().StringVarP(&sessionFlag, "session", "s", "", "chat session ID to resume directly (e.g. CHAT-20260921-0001)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
