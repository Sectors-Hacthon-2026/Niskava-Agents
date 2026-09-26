package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// DiagnosticStatus represents check status level.
type DiagnosticStatus string

const (
	StatusOk   DiagnosticStatus = "OK"
	StatusWarn DiagnosticStatus = "WARN"
	StatusFail DiagnosticStatus = "FAIL"
)

// DiagnosticCheck represents a single evaluated health item.
type DiagnosticCheck struct {
	Name           string
	Status         DiagnosticStatus
	Details        string
	Recommendation string
}

// DiagnosticReport contains the complete environment evaluation.
type DiagnosticReport struct {
	OS        string
	Arch      string
	GoVersion string
	Checks    []DiagnosticCheck
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run comprehensive system health and environment diagnostics",
	Long:  "Evaluates host platform compatibility, local SQLite WAL database, Python quantitative runtime, AI provider endpoints, and Sectors v2 connectivity.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// doctor handles its own diagnostics even if database cannot be opened
		if cfg == nil {
			cfg, _ = config.Load(cfgFile)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		report := EvaluateSystemDiagnostics()
		RenderDoctorReport(report)
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}

// EvaluateSystemDiagnostics inspects the local runtime, configurations, and external endpoints.
func EvaluateSystemDiagnostics() DiagnosticReport {
	if cfg == nil {
		cfg, _ = config.Load("")
	}
	report := DiagnosticReport{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}

	// 1. Host Platform Check
	report.Checks = append(report.Checks, DiagnosticCheck{
		Name:    "Host Platform",
		Status:  StatusOk,
		Details: fmt.Sprintf("%s/%s (%s)", runtime.GOOS, runtime.GOARCH, runtime.Version()),
	})

	// 2. Local-First SQLite Database (Law 4)
	dbPath := "~/.niskava/niskava.db"
	if cfg != nil && cfg.Storage.DBPath != "" {
		dbPath = cfg.Storage.DBPath
	}
	expandedDB := config.ExpandHome(dbPath)
	dbDir := filepath.Dir(expandedDB)

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:           "Local SQLite Database",
			Status:         StatusFail,
			Details:        fmt.Sprintf("Directory %s unwritable: %v", dbDir, err),
			Recommendation: fmt.Sprintf("Ensure read/write permissions for %s", dbDir),
		})
	} else {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:    "Local SQLite Database (WAL)",
			Status:  StatusOk,
			Details: fmt.Sprintf("%s (Law 4 Sovereign Storage)", expandedDB),
		})
	}

	// 3. Python Runtime & Quantitative Mathematics (Law 1)
	pyBin, pyDesc, pyReady := DetectPythonEnvironment()
	if pyReady {
		// Double check numpy and networkx
		cmd := exec.Command(pyBin, "-c", "import numpy, networkx; print('ok')")
		if out, err := cmd.CombinedOutput(); err == nil && strings.TrimSpace(string(out)) == "ok" {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:    "Python Quant Engine (NumPy/NetworkX)",
				Status:  StatusOk,
				Details: fmt.Sprintf("%s (%s)", pyBin, pyDesc),
			})
		} else {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:           "Python Quant Engine",
				Status:         StatusWarn,
				Details:        fmt.Sprintf("%s found but quantitative libraries missing or incomplete", pyBin),
				Recommendation: "Run 'niskava setup' to auto-install requirements or 'pip install -r backend/engine/requirements.txt'",
			})
		}
	} else {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:           "Python Quant Engine",
			Status:         StatusFail,
			Details:        pyDesc,
			Recommendation: "Install Python 3.11+ and run 'niskava setup' to initialize virtual environment",
		})
	}

	// 4. Configuration & AI Model Inference Provider
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	activeProvider := "openai"
	activeModel := "deepseek/deepseek-chat"
	activeBaseURL := "https://openrouter.ai/api/v1"
	activeKey := ""

	if cfg != nil {
		if cfg.Auth.AIProvider != "" {
			activeProvider = cfg.Auth.AIProvider
		}
		if cfg.Auth.OpenAIModel != "" {
			activeModel = cfg.Auth.OpenAIModel
		}
		if cfg.Auth.OpenAIBaseURL != "" {
			activeBaseURL = cfg.Auth.OpenAIBaseURL
		}
		if activeProvider == "gemini" {
			activeKey = cfg.Auth.GeminiAPIKey
			if cfg.Auth.GeminiModel != "" {
				activeModel = cfg.Auth.GeminiModel
			}
		} else {
			activeKey = cfg.Auth.OpenAIAPIKey
		}
	}

	if activeKey == "" && activeProvider != "ollama" {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:           "AI Inference Provider",
			Status:         StatusWarn,
			Details:        fmt.Sprintf("%s (%s) - API key not configured", activeProvider, activeModel),
			Recommendation: "Run 'niskava setup' or set OPENAI_API_KEY / GEMINI_API_KEY in ~/.niskava/config.yaml",
		})
	} else {
		ok, msg, _ := TestLiveConnection(ctx, activeProvider, activeBaseURL, activeKey)
		if ok {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:    "AI Inference Provider",
				Status:  StatusOk,
				Details: fmt.Sprintf("%s (%s) - %s", activeProvider, activeModel, msg),
			})
		} else {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:           "AI Inference Provider",
				Status:         StatusWarn,
				Details:        fmt.Sprintf("%s (%s) - %s", activeProvider, activeModel, msg),
				Recommendation: "Verify network connectivity or check if API key has sufficient quota",
			})
		}
	}

	// 5. Sectors Financial API v2 (Law 5)
	sectorsKey := ""
	if cfg != nil {
		sectorsKey = cfg.Auth.SectorsAPIKey
	}
	if sectorsKey == "" || (cfg != nil && cfg.Preferences.OfflineMode) {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:    "Sectors Financial API (IDX)",
			Status:  StatusOk,
			Details: "Offline Mock Mode Active (Law 5: 100% Credit Conservation / Static Fixtures)",
		})
	} else {
		secOK, secMsg, _ := TestLiveConnection(ctx, "sectors", "", sectorsKey)
		if secOK {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:    "Sectors Financial API (IDX)",
				Status:  StatusOk,
				Details: fmt.Sprintf("Live Connected (%s) - %s", config.MaskSecret(sectorsKey), secMsg),
			})
		} else {
			report.Checks = append(report.Checks, DiagnosticCheck{
				Name:           "Sectors Financial API (IDX)",
				Status:         StatusWarn,
				Details:        fmt.Sprintf("Connection issue (%s)", secMsg),
				Recommendation: "Check internet connection or run in offline mode via 'niskava investigate --offline'",
			})
		}
	}

	// 6. Repository Root & Engine Directory
	wd, _ := os.Getwd()
	repoRoot := ipc.ResolveRepoRoot(wd)
	enginePath := ipc.ResolveEnginePath(repoRoot, "")
	if _, err := os.Stat(filepath.Join(enginePath, "runner.py")); err == nil {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:    "Engine Directory Mobility",
			Status:  StatusOk,
			Details: fmt.Sprintf("Resolved: %s", enginePath),
		})
	} else {
		report.Checks = append(report.Checks, DiagnosticCheck{
			Name:           "Engine Directory Mobility",
			Status:         StatusWarn,
			Details:        fmt.Sprintf("runner.py not found at %s", enginePath),
			Recommendation: "Specify path using --engine-path flag or set NISKAVA_ENGINE_PATH",
		})
	}

	return report
}

// RenderDoctorReport outputs a beautifully formatted diagnostic HUD card to terminal.
func RenderDoctorReport(report DiagnosticReport) {
	badgeOk := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorSuccess).Render("[✓ READY]")
	badgeWarn := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorWarning).Render("[! WARN ]")
	badgeFail := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorDanger).Render("[✗ FAIL ]")

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorAccent).
		Width(80).
		Padding(0, 1).
		Foreground(tui.ColorFg)

	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorBg).Background(tui.ColorAccent).Padding(0, 1).
		Render("NISKAVA AGENT — SYSTEM & ENVIRONMENT DOCTOR")
	sb.WriteString(title + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(tui.ColorMuted).
		Render(fmt.Sprintf("Target Platform: %s/%s  •  Go: %s", report.OS, report.Arch, report.GoVersion)) + "\n\n")

	hasFail := false
	hasWarn := false

	for _, c := range report.Checks {
		var badge string
		switch c.Status {
		case StatusOk:
			badge = badgeOk
		case StatusWarn:
			badge = badgeWarn
			hasWarn = true
		case StatusFail:
			badge = badgeFail
			hasFail = true
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", badge, lipgloss.NewStyle().Bold(true).Render(c.Name)))
		sb.WriteString(fmt.Sprintf("    %s\n", lipgloss.NewStyle().Foreground(tui.ColorMuted).Render(c.Details)))
		if c.Recommendation != "" {
			sb.WriteString(fmt.Sprintf("    %s %s\n",
				lipgloss.NewStyle().Foreground(tui.ColorWarning).Render("→ Recommendation:"),
				c.Recommendation))
		}
		sb.WriteString("\n")
	}

	summaryStyle := lipgloss.NewStyle().Bold(true)
	if hasFail {
		sb.WriteString(summaryStyle.Foreground(tui.ColorDanger).Render("Summary: System has critical issues preventing execution.") + "\n")
	} else if hasWarn {
		sb.WriteString(summaryStyle.Foreground(tui.ColorWarning).Render("Summary: System is functional with non-critical recommendations.") + "\n")
	} else {
		sb.WriteString(summaryStyle.Foreground(tui.ColorSuccess).Render("Summary: All systems operational. Niskava Agent is 100% ready!") + "\n")
	}

	fmt.Println()
	fmt.Println(cardStyle.Render(sb.String()))
	fmt.Println()
}
