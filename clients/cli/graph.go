// Package cli provides Cobra CLI routing and subcommands for Niskava Agent.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/db"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/spf13/cobra"
)

var (
	graphOutputFlag  string
	graphSessionFlag string
	graphTickerFlag  string
	graphDepthFlag   int
	graphTextFlag    bool
	graphPruneFlag   bool
	graphOpenFlag    bool
)

type graphSummaryOutput struct {
	TotalNodes int `json:"total_nodes"`
	TotalEdges int `json:"total_edges"`
	Nodes      []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
		Group string `json:"group"`
	} `json:"nodes"`
	Edges []struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Rel    string  `json:"rel"`
		Weight float64 `json:"weight"`
	} `json:"edges"`
	Stats struct {
		TopCentralEntities []struct {
			ID          string  `json:"id"`
			Label       string  `json:"label"`
			Type        string  `json:"type"`
			Connections int     `json:"connections"`
			PageRank    float64 `json:"pagerank"`
		} `json:"top_central_entities"`
	} `json:"stats"`
}

// graphCmd represents the command to render, inspect, or manage the Market Intelligence knowledge graph.
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Export, inspect, or prune the interactive Market Intelligence knowledge graph",
	Long: `Renders an interactive Market Intelligence knowledge graph visualization of market entities,
anomalies, catalysts, and conversational memories into a standalone HTML file or terminal summary.

Examples:
  niskava graph --open
  niskava graph --ticker ANTM --depth 2 --open
  niskava graph --text
  niskava graph --ticker BBRI --text
  niskava graph --prune`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, ".niskava", "niskava.db")
		if cfg != nil && cfg.Storage.DBPath != "" {
			dbPath = config.ExpandHome(cfg.Storage.DBPath)
		}
		if customDB := os.Getenv("NISKAVA_DB_PATH"); customDB != "" {
			dbPath = config.ExpandHome(customDB)
		}

		// 1. Prune mock evaluation data if requested
		if graphPruneFlag {
			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("gagal membuka database: %w", err)
			}
			defer database.Close()

			deleted, err := database.PruneMockTestData()
			if err != nil {
				return fmt.Errorf("gagal membersihkan data uji coba: %w", err)
			}
			fmt.Printf(tui.T("graph_pruned_success"), deleted)
			return nil
		}

		pythonBin := ""
		if cfg != nil {
			pythonBin = cfg.Engine.PythonBin
		}
		pythonBin = ipc.ResolvePythonBin(pythonBin)
		wd, _ := os.Getwd()

		// 2. Terminal text summary inspection
		if graphTextFlag {
			execArgs := []string{
				"-m", "engine.runner",
				"--db-path", dbPath,
				"--summary-graph",
				"--depth", strconv.Itoa(graphDepthFlag),
			}
			if graphSessionFlag != "" {
				execArgs = append(execArgs, "--session", graphSessionFlag)
			}
			if graphTickerFlag != "" {
				execArgs = append(execArgs, "--ticker", strings.ToUpper(graphTickerFlag))
			}

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
				return fmt.Errorf("gagal mengambil data graf: %w\nOutput: %s", err, string(out))
			}

			var summary graphSummaryOutput
			if err := json.Unmarshal(out, &summary); err != nil {
				return fmt.Errorf("gagal memproses respon graf: %w\nRaw: %s", err, string(out))
			}

			scopeStr := "Global Market Graph"
			if graphTickerFlag != "" {
				scopeStr = fmt.Sprintf("Ego-Network (%s, radius=%d)", strings.ToUpper(graphTickerFlag), graphDepthFlag)
			} else if graphSessionFlag != "" {
				scopeStr = fmt.Sprintf("Session (%s)", graphSessionFlag)
			}

			fmt.Println("\n" + strings.Repeat("═", 64))
			fmt.Println("  NISKAVA AGENT — MEMORY GRAPH INTELLIGENCE REPORT")
			fmt.Println(strings.Repeat("═", 64))
			fmt.Printf("  Scope       : %s\n", scopeStr)
			fmt.Printf("  Total Nodes : %d\n", summary.TotalNodes)
			fmt.Printf("  Total Edges : %d\n", summary.TotalEdges)
			fmt.Println(strings.Repeat("─", 64))
			fmt.Println("  TOP CENTRAL HUBS & ENTITIES:")
			if len(summary.Stats.TopCentralEntities) == 0 {
				fmt.Println("    (Belum ada entitas sentral terdeteksi)")
			} else {
				for _, h := range summary.Stats.TopCentralEntities {
					prStr := ""
					if h.PageRank > 0 {
						prStr = fmt.Sprintf(" [PR: %.4f]", h.PageRank)
					}
					fmt.Printf("    • %-16s (%-10s) : %d koneksi%s\n", h.Label, h.Type, h.Connections, prStr)
				}
			}
			fmt.Println(strings.Repeat("─", 64))
			fmt.Println("  ACTIVE MARKET RELATIONS:")
			if len(summary.Edges) == 0 {
				fmt.Println("    (Tidak ada relasi aktif untuk cakupan ini)")
			} else {
				limit := len(summary.Edges)
				if limit > 12 {
					limit = 12
				}
				for i := 0; i < limit; i++ {
					e := summary.Edges[i]
					fmt.Printf("    • %-14s ──[%-16s | w:%.2f]──> %s\n", e.From, e.Rel, e.Weight, e.To)
				}
				if len(summary.Edges) > 12 {
					fmt.Printf("    ... dan %d relasi lainnya.\n", len(summary.Edges)-12)
				}
			}
			fmt.Println(strings.Repeat("═", 64) + "\n")
			return nil
		}

		// 3. HTML graph visualization export
		outputPath := graphOutputFlag
		if outputPath == "" {
			outputPath = filepath.Join(homeDir, ".niskava", "graph.html")
		}
		expandedOutput := filepath.Clean(config.ExpandHome(outputPath))

		execArgs := []string{
			"-m", "engine.runner",
			"--db-path", dbPath,
			"--export-graph-html", expandedOutput,
			"--depth", strconv.Itoa(graphDepthFlag),
		}
		if graphSessionFlag != "" {
			execArgs = append(execArgs, "--session", graphSessionFlag)
		}
		if graphTickerFlag != "" {
			execArgs = append(execArgs, "--ticker", strings.ToUpper(graphTickerFlag))
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
	graphCmd.Flags().StringVarP(&graphTickerFlag, "ticker", "t", "", "fokus graf pada emiten tertentu (contoh: ANTM)")
	graphCmd.Flags().IntVarP(&graphDepthFlag, "depth", "d", 1, "radius tetangga (hop depth: 1 atau 2)")
	graphCmd.Flags().BoolVarP(&graphTextFlag, "text", "", false, "tampilkan ringkasan teks graf di terminal tanpa membuka browser")
	graphCmd.Flags().BoolVarP(&graphPruneFlag, "prune", "", false, "bersihkan data uji coba / benchmark (EVAL-*) dari basis data lokal")
	graphCmd.Flags().BoolVarP(&graphOpenFlag, "open", "", true, "buka visualisasi HTML di browser secara otomatis")
	RootCmd.AddCommand(graphCmd)
}
