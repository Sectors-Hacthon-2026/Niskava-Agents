// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/spf13/cobra"
)

var (
	graphOutputFlag  string
	graphSessionFlag string
	graphOpenFlag    bool
)

// graphCmd represents the command to render or view the Market Intelligence knowledge graph.
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export and view interactive Market Intelligence knowledge graph in your browser",
	Long: `Renders an interactive Market Intelligence knowledge graph visualization of market entities,
anomalies, catalysts, and conversational memories into a standalone HTML file.

Example:
  niskava graph --open
  niskava graph --session INV-2026-0042 -o ./audit_graph.html --open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pythonBin := ""
		if cfg != nil {
			pythonBin = cfg.Engine.PythonBin
		}
		pythonBin = ipc.ResolvePythonBin(pythonBin)

		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, ".niskava", "niskava.db")
		if cfg != nil && cfg.Storage.DBPath != "" {
			dbPath = config.ExpandHome(cfg.Storage.DBPath)
		}
		if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
			dbPath = config.ExpandHome(customDB)
		}

		outputPath := graphOutputFlag
		if outputPath == "" {
			outputPath = filepath.Join(homeDir, ".niskava", "graph.html")
		}
		expandedOutput := filepath.Clean(config.ExpandHome(outputPath))

		wd, _ := os.Getwd()
		execArgs := []string{
			"-m", "engine.runner",
			"--db-path", dbPath,
			"--export-graph-html", expandedOutput,
		}
		if graphSessionFlag != "" {
			execArgs = append(execArgs, "--session", graphSessionFlag)
		}

		fmt.Printf(tui.T("graph_exporting"), expandedOutput)

		proc := exec.Command(pythonBin, execArgs...)
		proc.Dir = wd
		pythonPath := filepath.Join(wd, "backend") + string(filepath.ListSeparator) + wd
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		proc.Env = append(os.Environ(),
			"PYTHONPATH="+pythonPath,
			"PYTHONIOENCODING=utf-8",
			"PYTHONUTF8=1",
		)
		out, err := proc.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gagal mengekspor visualisasi graf: %w\nOutput: %s", err, string(out))
		}

		fmt.Printf(tui.T("graph_exported_success"), expandedOutput)

		if graphOpenFlag {
			absOutput, errAbs := filepath.Abs(expandedOutput)
			if errAbs == nil {
				expandedOutput = absOutput
			}
			fileURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(expandedOutput), "/")
			fmt.Printf(tui.T("graph_opening_browser"), fileURL)
			_ = server.OpenBrowser(fileURL)
		} else {
			fmt.Printf(tui.T("graph_tip"), expandedOutput)
		}

		return nil
	},
}

func init() {
	graphCmd.Flags().StringVarP(&graphOutputFlag, "output", "o", "", "path file output HTML (default: ~/.niskava/graph.html)")
	graphCmd.Flags().StringVarP(&graphSessionFlag, "session", "s", "", "filter simpul & relasi berdasarkan ID sesi")
	graphCmd.Flags().BoolVarP(&graphOpenFlag, "open", "", true, "buka visualisasi HTML di browser secara otomatis")
	RootCmd.AddCommand(graphCmd)
}
