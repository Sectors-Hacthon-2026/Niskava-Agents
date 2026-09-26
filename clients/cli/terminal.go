// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"context"
	"fmt"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/spf13/cobra"
)

var terminalCmd = &cobra.Command{
	Use:     "terminal",
	Aliases: []string{"repl", "chat"},
	Short:   "Launch interactive conversational intelligence REPL directly",
	Long: `Launch interactive Hermes-style conversational research REPL terminal directly
to dialog with the investigation agent, evaluate quantitative anomalies, and inspect evidence.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}
		srv.ConfigPath = cfgFile

		if sessionFlag != "" {
			_ = tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)
		} else {
			_ = tui.RunLiveREPL(cfg, appDB, srv.URL)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(terminalCmd)
}
