// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/spf13/cobra"
)

var (
	mcpStdioFlag bool
	mcpPyBinFlag string
)

// mcpCmd represents the mcp command to start the Model Context Protocol server.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start Model Context Protocol (MCP) server for external AI agents (Claude Desktop, Cursor, Antigravity)",
	Long: `Starts the Niskava Unified Model Context Protocol (MCP) server over standard input/output (stdio).

Exposes:
- Sectors Financial API v2 Market Data Primitives
- Dual-Engine OSINT Harvester & Trafilatura Content Sanitizer
- Deterministic Quantitative Math Calculations (NumPy Firewall)
- Local Associative Graph Memory Context & Storage
- MCP Resources (System status, cache metrics)
- MCP Prompts (Standardized investigative workflows)

Example Claude Desktop configuration (~/.config/Claude/claude_desktop_config.json):
{
  "mcpServers": {
    "niskava": {
      "command": "niskava",
      "args": ["mcp"]
    }
  }
}`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine Python binary: CLI flag > Config file / Env > Virtual env fallback
		pythonBin := mcpPyBinFlag
		if pythonBin == "" && cfg != nil {
			pythonBin = cfg.Engine.PythonBin
		}
		pythonBin = ipc.ResolvePythonBin(pythonBin)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Trap SIGINT and SIGTERM for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		wd, _ := os.Getwd()

		proc := exec.CommandContext(ctx, pythonBin, "-u", "-m", "engine.mcp.server")
		proc.Dir = wd
		proc.Stdin = os.Stdin
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr

		// Build environment
		proc.Env = os.Environ()
		pythonPath := filepath.Join(wd, "backend") + string(filepath.ListSeparator) + wd
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		proc.Env = append(proc.Env,
			"PYTHONPATH="+pythonPath,
			"PYTHONUNBUFFERED=1",
			"PYTHONIOENCODING=utf-8",
			"PYTHONUTF8=1",
		)
		if cfg != nil {
			proc.Env = append(proc.Env, fmt.Sprintf("NISKAVA_DB_PATH=%s", cfg.Storage.DBPath))
			if cfg.Auth.SectorsAPIKey != "" {
				proc.Env = append(proc.Env, fmt.Sprintf("SECTORS_API_KEY=%s", cfg.Auth.SectorsAPIKey))
			}
			if cfg.Preferences.OfflineMode {
				proc.Env = append(proc.Env, "MOCK_SECTORS=1", "NISKAVA_OFFLINE=1")
			}
		}

		if err := proc.Run(); err != nil {
			if ctx.Err() != nil {
				return nil // Clean cancellation
			}
			return fmt.Errorf("MCP server process terminated with error: %w", err)
		}

		return nil
	},
}

func init() {
	mcpCmd.Flags().BoolVar(&mcpStdioFlag, "stdio", true, "run MCP server in standard I/O (stdio) mode")
	mcpCmd.Flags().StringVar(&mcpPyBinFlag, "python-bin", "", "path to python binary for MCP engine")
	RootCmd.AddCommand(mcpCmd)
}
