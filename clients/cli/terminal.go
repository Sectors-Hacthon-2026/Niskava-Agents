// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"context"
	"fmt"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/spf13/cobra"
)

var terminalCmd = &cobra.Command{
	Use:     "terminal",
	Aliases: []string{"repl", "chat"},
	Short:   "Launch interactive conversational intelligence REPL directly",
	Long: `Launch interactive conversational research REPL terminal directly
to dialog with the investigation agent, evaluate quantitative anomalies, and inspect evidence.`,
	Example: `  # Launch interactive REPL directly:
  niskava terminal
  niskava repl
  niskava chat

  # Resume a specific session directly:
  niskava terminal --session CHAT-20260926-0001

  # Launch in English:
  niskava terminal --lang en`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}
		srv.ConfigPath = cfgFile

		for {
			var res string
			if sessionFlag != "" {
				res = tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)
			} else {
				res = tui.RunLiveREPL(cfg, appDB, srv.URL)
			}
			if res == tui.ReplSetupSentinel {
				_ = RunInteractiveSetup()
				if newCfg, err := config.Load(cfgFile); err == nil {
					cfg = newCfg
					srv.Config = newCfg
				}
				continue
			}
			break
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(terminalCmd)
}
