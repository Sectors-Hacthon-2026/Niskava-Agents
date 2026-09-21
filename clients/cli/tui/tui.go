// Package tui provides an interactive Terminal User Interface for Niskava Agent,
// featuring live thought streaming, dynamic tool execution badges, and Bloomberg/OSINT audit trail cards.
package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Aesthetic styles using lipgloss (Light Green / Matrix OSINT Theme)
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22C55E")).
			Background(lipgloss.Color("#052E16")).
			Padding(0, 1)

	tickerBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#052E16")).
				Background(lipgloss.Color("#4ADE80")).
				Padding(0, 1)

	thoughtBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22C55E")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#86EFAC")).
			Italic(true)

	toolCallingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FACC15"))

	toolDoneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22C55E"))

	anomalyBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FCA5A5"))

	supportedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22C55E")).
			SetString("[SUPPORTED]")

	uncertainStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FACC15")).
			SetString("[UNCERTAIN]")

	contradictedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#EF4444")).
				SetString("[CONTRADICTED]")

	disclaimerBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#475569")).
				Padding(0, 1).
				Foreground(lipgloss.Color("#94A3B8"))
)

// ToolActivity tracks a single tool invocation step.
type ToolActivity struct {
	ToolName    string
	Arguments   string
	Observation string
	Done        bool
}

// Model is the Bubbletea state container for Niskava Agent.
type Model struct {
	Ticker         string
	Days           int
	SessionID      string
	DBPath         string
	Spinner        spinner.Model
	CurrentThought string
	ToolActivities []ToolActivity
	Anomalies      []ipc.Event
	Findings       []ipc.Event
	Summary        string
	Completed      bool
	Err            error
	EventsChan     <-chan ipc.Event
	ErrChan        <-chan error
}

// NewModel creates an interactive TUI model.
func NewModel(ticker string, days int, dbPath string, eventsChan <-chan ipc.Event, errChan <-chan error) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80"))

	return Model{
		Ticker:     ticker,
		Days:       days,
		DBPath:     dbPath,
		Spinner:    s,
		EventsChan: eventsChan,
		ErrChan:    errChan,
	}
}

// Msg types
type eventMsg ipc.Event
type errMsg error
type finishMsg struct{}

func waitForEvent(eventsChan <-chan ipc.Event, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		activeErr := errChan
		for {
			select {
			case ev, ok := <-eventsChan:
				if !ok {
					return finishMsg{}
				}
				return eventMsg(ev)
			case err, ok := <-activeErr:
				if !ok {
					activeErr = nil
					continue
				}
				if err != nil {
					return errMsg(err)
				}
			}
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		waitForEvent(m.EventsChan, m.ErrChan),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case errMsg:
		m.Err = msg
		m.Completed = true
		return m, tea.Quit

	case finishMsg:
		m.Completed = true
		return m, tea.Quit

	case eventMsg:
		ev := ipc.Event(msg)
		switch ev.Event {
		case ipc.EventSessionStart:
			m.SessionID = ev.SessionID

		case ipc.EventAgentThought:
			m.CurrentThought = ev.Thought

		case ipc.EventAgentToolCall:
			argsBytes, _ := json.Marshal(ev.Args)
			m.ToolActivities = append(m.ToolActivities, ToolActivity{
				ToolName:  ev.Tool,
				Arguments: string(argsBytes),
				Done:      false,
			})

		case ipc.EventAgentObservation:
			if len(m.ToolActivities) > 0 {
				m.ToolActivities[len(m.ToolActivities)-1].Observation = ev.Summary
				m.ToolActivities[len(m.ToolActivities)-1].Done = true
			}

		case ipc.EventAnomalyDetected:
			m.Anomalies = append(m.Anomalies, ev)

		case ipc.EventFindingEmitted:
			m.Findings = append(m.Findings, ev)

		case ipc.EventSessionComplete:
			m.Summary = ev.Summary
			m.Completed = true
			return m, tea.Quit
		}

		return m, waitForEvent(m.EventsChan, m.ErrChan)
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// 1. Header Banner
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(T("header_title")))
	b.WriteString(T("target_label"))
	b.WriteString(tickerBadgeStyle.Render(m.Ticker))
	b.WriteString(fmt.Sprintf(T("observation_horizon"), m.Days))
	b.WriteString(RenderConstellationLine(80) + "\n\n")

	// 2. Live Thought Stream (ReAct Inner Monologue)
	if m.CurrentThought != "" {
		thoughtHeader := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38BDF8")).Render(T("agent_reasoning"))
		b.WriteString(fmt.Sprintf("%s\n", thoughtHeader))
		b.WriteString(thoughtBoxStyle.Render(m.CurrentThought) + "\n\n")
	}

	// 3. Dynamic Tool Invocations
	if len(m.ToolActivities) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8FAFC")).Render(T("tool_activity")) + "\n")
		for _, act := range m.ToolActivities {
			if act.Done {
				b.WriteString(fmt.Sprintf("  %s %s\n", toolDoneStyle.Render("✔"), lipgloss.NewStyle().Bold(true).Render(act.ToolName)))
				if act.Observation != "" {
					b.WriteString(fmt.Sprintf("    %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render(act.Observation)))
				}
			} else {
				b.WriteString(fmt.Sprintf("  %s %s %s...\n", toolCallingStyle.Render("⚡"), m.Spinner.View(), act.ToolName))
			}
		}
		b.WriteString("\n")
	}

	// 4. Anomaly Alert Box (if triggered)
	if len(m.Anomalies) > 0 {
		var anomLines []string
		for _, a := range m.Anomalies {
			anomLines = append(anomLines, fmt.Sprintf("• [%s] %s (Z-Score: %.2fσ | Return: %+.2f%%)\n  %s",
				a.AnomalyDate, a.MetricType, a.ZScore, a.PriceChangePct, a.Description))
		}
		b.WriteString(anomalyBoxStyle.Render(fmt.Sprintf("%s\n%s", T("anomaly_detected"), strings.Join(anomLines, "\n"))) + "\n\n")
	}

	// 5. Findings Section (Audit Trail 3-Tier Taxonomy)
	if len(m.Findings) > 0 {
		b.WriteString("─────────────────────────────────────────────────────────────────────────────\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(T("audit_trail_summary")) + "\n")

		for _, f := range m.Findings {
			var badge string
			switch f.VerificationStat {
			case "SUPPORTED":
				badge = supportedStyle.Render()
			case "UNCERTAIN":
				badge = uncertainStyle.Render()
			case "CONTRADICTED":
				badge = contradictedStyle.Render()
			default:
				badge = "[" + f.VerificationStat + "]"
			}

			b.WriteString(fmt.Sprintf("\n%s %s\n", badge, lipgloss.NewStyle().Bold(true).Render(f.Title)))
			b.WriteString(fmt.Sprintf("            %s\n", f.ClaimText))
			b.WriteString(fmt.Sprintf("            %s: %s | %s: %.2f\n", T("causality_label"), f.CausalityStatus, T("confidence_score_label"), f.ConfidenceScore))
		}
	}

	// 6. Final Summary
	if m.Summary != "" {
		b.WriteString("\n─────────────────────────────────────────────────────────────────────────────\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5FF")).Render(m.Summary) + "\n")
	}

	// 7. Non-Advisory Disclaimer Footer (Law 2 / Hackathon Rule 12)
	b.WriteString("\n")
	disclaimerText := T("financial_disclaimer")
	b.WriteString(disclaimerBoxStyle.Render(disclaimerText) + "\n\n")

	// 8. Navigation hint
	if m.SessionID != "" {
		b.WriteString(fmt.Sprintf(T("session_saved_hint"), m.SessionID, m.DBPath))
	}

	if m.Err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(fmt.Sprintf("\n[ERROR] %v\n", m.Err)))
	}

	return b.String()
}
