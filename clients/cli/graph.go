// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/spf13/cobra"
)

var (
	graphOutputFlag  string
	graphSessionFlag string
	graphOpenFlag    bool
)

// graphCmd represents the command to render or view the Cyber-OSINT knowledge graph.
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export and view interactive Cyber-OSINT knowledge graph in your browser",
	Long: `Renders an interactive Cyber-OSINT knowledge graph visualization of market entities,
anomalies, catalysts, and conversational memories into a standalone HTML file.

Example:
  niskava graph --open
  niskava graph --session INV-2026-0042 -o ./audit_graph.html --open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pythonBin := "python3"
		if cfg != nil && cfg.Engine.PythonBin != "" {
			pythonBin = cfg.Engine.PythonBin
		}
		if pythonBin == "" || pythonBin == "python3" {
			localVenv := filepath.Join(".venv", "bin", "python3")
			if _, err := os.Stat(localVenv); err == nil {
				pythonBin = localVenv
			}
		}

		dbPath := filepath.Join(os.Getenv("HOME"), ".niskava", "niskava.db")
		if cfg != nil && cfg.Storage.DBPath != "" {
			dbPath = cfg.Storage.DBPath
		}
		if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
			dbPath = customDB
		}

		outputPath := graphOutputFlag
		if outputPath == "" {
			outputPath = filepath.Join(os.Getenv("HOME"), ".niskava", "graph.html")
		}
		expandedOutput := filepath.Clean(outputPath)
		if len(outputPath) > 0 && outputPath[0] == '~' {
			expandedOutput = filepath.Join(os.Getenv("HOME"), outputPath[1:])
		}

		wd, _ := os.Getwd()
		execArgs := []string{
			"-m", "engine.runner",
			"--db-path", dbPath,
			"--export-graph-html", expandedOutput,
		}
		if graphSessionFlag != "" {
			execArgs = append(execArgs, "--session", graphSessionFlag)
		}

		fmt.Printf("⚡ Mengekspor Cyber-OSINT Knowledge Graph ke %s...\n", expandedOutput)

		proc := exec.Command(pythonBin, execArgs...)
		proc.Dir = wd
		pythonPath := filepath.Join(wd, "backend") + string(filepath.ListSeparator) + wd
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		proc.Env = append(os.Environ(), "PYTHONPATH="+pythonPath)
		out, err := proc.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gagal mengekspor visualisasi graf: %w\nOutput: %s", err, string(out))
		}

		fmt.Printf("[✓] File visualisasi graf berhasil dibuat: %s\n", expandedOutput)

		if graphOpenFlag {
			fileURL := fmt.Sprintf("file://%s", expandedOutput)
			fmt.Printf("Membuka di peramban web: %s\n", fileURL)
			_ = server.OpenBrowser(fileURL)
		} else {
			fmt.Printf("Tip: Jalankan dengan flag --open atau buka langsung di browser Anda:\n  file://%s\n", expandedOutput)
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
