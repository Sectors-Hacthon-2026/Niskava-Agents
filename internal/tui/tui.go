// Package tui provides an interactive Terminal User Interface built with
// charmbracelet/bubbletea and lipgloss, adhering to the Cyber-OSINT / Bloomberg aesthetic.
package tui

import (
	"fmt"
	"strings"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/ipc"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles definitions using lipgloss
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E5FF")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)

	tickerBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#0052CC")).
				Padding(0, 1)

	stagePendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C7293"))

	stageRunningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00E5FF"))

	stageOkStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E676"))

	stageAlertStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD600"))

	supportedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E676")).
			SetString("[SUPPORTED]")

	uncertainStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD600")).
			SetString("[UNCERTAIN]")

	contradictedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF1744")).
				SetString("[CONTRADICTED]")

	disclaimerBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#44475A")).
				Padding(0, 1).
				Foreground(lipgloss.Color("#8BE9FD"))
)

// StageState represents the progress status of each pipeline stage.
type StageState struct {
	Name    string
	Status  string // "PENDING", "RUNNING", "OK", "ALERT"
	Details string
}

// Model is the Bubbletea state container for investigation TUI.
type Model struct {
	Ticker     string
	Days       int
	SessionID  string
	DBPath     string
	Spinner    spinner.Model
	Stages     []StageState
	Anomalies  []ipc.Event
	Findings   []ipc.Event
	Summary    string
	Completed  bool
	Err        error
	EventsChan <-chan ipc.Event
	ErrChan    <-chan error
}

// NewModel creates a configured TUI model.
func NewModel(ticker string, days int, dbPath string, eventsChan <-chan ipc.Event, errChan <-chan error) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E5FF"))

	stages := []StageState{
		{Name: "Baseline Data Sectors v2", Status: "PENDING", Details: "Menyiapkan deret waktu candle & profil emiten"},
		{Name: "Deteksi Anomali Kuantitatif", Status: "PENDING", Details: "Menghitung Z-Score volume (MA20) & abnormal return"},
		{Name: "Penelusuran OSINT Bertarget", Status: "PENDING", Details: "Memanen keterbukaan informasi bursa & berita pasar"},
		{Name: "Validasi Bukti & Kausalitas", Status: "PENDING", Details: "Menilai urutan temporal & klasifikasi 3-tier"},
	}

	return Model{
		Ticker:     ticker,
		Days:       days,
		DBPath:     dbPath,
		Spinner:    s,
		Stages:     stages,
		EventsChan: eventsChan,
		ErrChan:    errChan,
	}
}

// Msg types
type eventMsg ipc.Event
type errMsg error
type finishMsg struct{}

// waitForEvent waits for the next incoming IPC event.
func waitForEvent(eventsChan <-chan ipc.Event, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case ev, ok := <-eventsChan:
			if !ok {
				return finishMsg{}
			}
			return eventMsg(ev)
		case err, ok := <-errChan:
			if ok && err != nil {
				return errMsg(err)
			}
			return finishMsg{}
		}
	}
}

// Init initializes the bubbletea loop.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		waitForEvent(m.EventsChan, m.ErrChan),
	)
}

// Update handles incoming messages and state transitions.
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
			m.Stages[0].Status = "RUNNING"

		case ipc.EventProgressStep:
			switch ev.Stage {
			case "SECTORS_BASELINE":
				m.Stages[0].Status = "RUNNING"
				m.Stages[0].Details = ev.Message
			case "QUANT_ANOMALY":
				m.Stages[0].Status = "OK"
				m.Stages[1].Status = "RUNNING"
				m.Stages[1].Details = ev.Message
			case "OSINT_HARVEST":
				if m.Stages[1].Status != "ALERT" {
					m.Stages[1].Status = "OK"
				}
				m.Stages[2].Status = "RUNNING"
				m.Stages[2].Details = ev.Message
			case "EVIDENCE_CORRELATION":
				m.Stages[2].Status = "OK"
				m.Stages[3].Status = "RUNNING"
				m.Stages[3].Details = ev.Message
			}

		case ipc.EventAnomalyDetected:
			m.Anomalies = append(m.Anomalies, ev)
			m.Stages[1].Status = "ALERT"
			m.Stages[1].Details = fmt.Sprintf("Lonjakan volume (%.2fσ) pada %s", ev.ZScore, ev.AnomalyDate)

		case ipc.EventFindingEmitted:
			m.Findings = append(m.Findings, ev)

		case ipc.EventSessionComplete:
			m.Stages[3].Status = "OK"
			m.Stages[3].Details = fmt.Sprintf("%d temuan tervalidasi", len(m.Findings))
			m.Summary = ev.Summary
			m.Completed = true
			return m, tea.Quit
		}

		return m, waitForEvent(m.EventsChan, m.ErrChan)
	}

	return m, nil
}

// View renders the TUI layout to ANSI string.
func (m Model) View() string {
	var b strings.Builder

	// 1. Header Banner
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" [●] NISKAVA INVESTIGATOR v1.0.0 "))
	b.WriteString(" Target: ")
	b.WriteString(tickerBadgeStyle.Render(m.Ticker))
	b.WriteString(fmt.Sprintf(" (%d Hari Pengamatan)\n", m.Days))
	b.WriteString("─────────────────────────────────────────────────────────────────────────────\n")

	// 2. Stepper Progress
	for i, stage := range m.Stages {
		prefix := " ├── "
		if i == len(m.Stages)-1 {
			prefix = " └── "
		}

		var statusBadge string
		switch stage.Status {
		case "PENDING":
			statusBadge = stagePendingStyle.Render("[WAIT]")
		case "RUNNING":
			statusBadge = stageRunningStyle.Render(fmt.Sprintf("[%s]", m.Spinner.View()))
		case "ALERT":
			statusBadge = stageAlertStyle.Render("[ALERT]")
		case "OK":
			statusBadge = stageOkStyle.Render("[OK]")
		}

		b.WriteString(fmt.Sprintf("%s[%d/4] %-32s %s %s\n", prefix, i+1, stage.Name, statusBadge, stage.Details))
	}

	// 3. Findings Section (Audit Trail)
	if len(m.Findings) > 0 {
		b.WriteString("\n─────────────────────────────────────────────────────────────────────────────\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("RINGKASAN TEMUAN (AUDIT TRAIL):") + "\n")

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
			b.WriteString(fmt.Sprintf("            Status: %s | Skor Keyakinan: %.2f\n", f.CausalityStatus, f.ConfidenceScore))
		}
	}

	// 4. Final Summary
	if m.Summary != "" {
		b.WriteString("\n─────────────────────────────────────────────────────────────────────────────\n")
		b.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#00E5FF")).Render(m.Summary) + "\n")
	}

	// 5. Non-Advisory Disclaimer Footer (Law 2 / Hackathon Rule 12)
	b.WriteString("\n")
	disclaimerText := "DISCLAIMER FINANSIAL (NON-ADVISORY - LAW 2 & ATURAN 12):\n" +
		"Niskava Agent adalah platform intelijen pasar dan OSINT otonom, BUKAN penasihat investasi.\n" +
		"Sistem TIDAK PERNAH memberikan rekomendasi BELI/JUAL atau target harga sekuritas."
	b.WriteString(disclaimerBoxStyle.Render(disclaimerText) + "\n\n")

	// 6. Navigation hint
	if m.SessionID != "" {
		b.WriteString(fmt.Sprintf("Sesi tersimpan: %s (%s)\n", m.SessionID, m.DBPath))
		b.WriteString("Ketik 'niskava serve --open' untuk membuka visual workspace interaktif di browser.\n")
	}

	if m.Err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1744")).Render(fmt.Sprintf("\n[ERROR] %v\n", m.Err)))
	}

	return b.String()
}
