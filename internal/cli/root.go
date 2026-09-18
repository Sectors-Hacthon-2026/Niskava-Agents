// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"fmt"
	"os"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/db"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	cfg     *config.Config
	appDB   *db.DB
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

		appDB, err = db.Open(cfg.Storage.DBPath)
		if err != nil {
			return fmt.Errorf("failed to open database at %s: %w", cfg.Storage.DBPath, err)
		}

		return nil
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
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
