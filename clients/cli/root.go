// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

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
	Short: "Niskava Agent — Autonomous IDX Market Intelligence & OSINT",
	Long: `Niskava Agent is an autonomous financial OSINT and market intelligence
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

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Bare command: start background daemon & launch 9router-style interface selector
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.Start(ctx, cfg.Server.Port, appDB)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}

		// If explicit --session flag provided, bypass launcher and jump directly into REPL
		if sessionFlag != "" {
			tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)
			return nil
		}

		hasAPIKey := cfg.Auth.SectorsAPIKey != "" || cfg.Auth.GeminiAPIKey != "" || cfg.Auth.OpenAIAPIKey != ""
		for {
			fmt.Print("\033[H\033[2J")
			launcher := tui.NewLauncherModelWithHealth(srv.URL, "v1.0.0", hasAPIKey)
			p := tea.NewProgram(launcher, tea.WithAltScreen())
			m, err := p.Run()
			if err != nil {
				return fmt.Errorf("launcher error: %w", err)
			}

			selected := m.(tui.LauncherModel).Selected
			switch selected {
			case "web":
				fmt.Printf("\n[●] Membuka Web Workspace di browser: %s\n", srv.URL)
				_ = server.OpenBrowser(srv.URL)
				fmt.Println("Tekan Enter untuk kembali ke Menu...")
				_, _ = fmt.Scanln()

			case "terminal":
				// Launch persistent live interactive CLI REPL
				tui.RunLiveREPL(cfg, appDB, srv.URL)

			case "sessions":
				// Show saved sessions with interactive resume option
				selectedSessionID := runSessionsInteractive(cmd, appDB)
				if selectedSessionID != "" {
					tui.RunLiveREPL(cfg, appDB, srv.URL, selectedSessionID)
				}

			case "help":
				tui.PrintFullHelpGuide()
				fmt.Println("\nTekan Enter untuk kembali ke Menu...")
				_, _ = fmt.Scanln()

			case "health":
				fmt.Println("\nSTATUS KESEHATAN SISTEM:")
				fmt.Println("─────────────────────────────────────────────────────────────────────────────")
				fmt.Printf("• Local Daemon URL: %s [ALIVE]\n", srv.URL)
				fmt.Printf("• Database Path   : %s\n", cfg.Storage.DBPath)
				fmt.Printf("• Python Engine   : %s\n", cfg.Engine.PythonBin)
				fmt.Printf("• Sectors API Key : %t (Terpasang)\n", cfg.Auth.SectorsAPIKey != "")
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
				hasModelKey := cfg.Auth.OpenAIAPIKey != "" || cfg.Auth.GeminiAPIKey != ""
				fmt.Printf("• Inference Engine: Universal ReAct (%s) [ALIVE]\n", baseURL)
				fmt.Printf("• Active Model    : %s\n", activeModel)
				fmt.Printf("• Model API Key   : %t (Terpasang)\n", hasModelKey)
				fmt.Println("─────────────────────────────────────────────────────────────────────────────")
				fmt.Println("Tekan Enter untuk kembali ke Menu...")
				_, _ = fmt.Scanln()

			case "lang":
				langModel := tui.NewLangSelectorModel()
				pLang := tea.NewProgram(langModel, tea.WithAltScreen())
				mLang, errLang := pLang.Run()
				if errLang == nil {
					selLang := mLang.(tui.LangSelectorModel).Selected
					if selLang != "" {
						tui.SetLanguage(selLang)
						cfg.Preferences.Language = tui.ActiveLanguage
					}
				}

			case "setup":
				_ = RunInteractiveSetup()
				fmt.Println("Tekan Enter untuk kembali ke Menu...")
				_, _ = fmt.Scanln()

			case "exit", "":
				fmt.Println("Menghentikan server daemon dan keluar dari Niskava Agent.")
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

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ~/.niskava/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&langFlag, "lang", "l", "", "language preference: 'en' for English (default) or 'id' for Indonesian")
	RootCmd.PersistentFlags().StringVarP(&sessionFlag, "session", "s", "", "chat session ID to resume directly (e.g. CHAT-20260921-0001)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
