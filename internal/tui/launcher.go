// Package tui provides interactive terminal interfaces including the 9router-style
// launcher menu and the persistent live REPL.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LauncherItem represents a selectable menu choice.
type LauncherItem struct {
	Title       string
	Description string
	ActionID    string
}

// LauncherModel is the Bubbletea model for the interface selector menu.
type LauncherModel struct {
	ServerURL string
	Version   string
	Items     []LauncherItem
	Cursor    int
	Selected  string
	Quitting  bool
}

var (
	headerBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF7A00")).
			Padding(0, 1)

	serverURLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0F172A")).
				Background(lipgloss.Color("#E2E8F0")).
				Padding(0, 1)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CBD5E1")).
				Padding(0, 1)
)

// NewLauncherModel initializes the 9router-style launcher menu.
func NewLauncherModel(serverURL string, version string) LauncherModel {
	items := []LauncherItem{
		{Title: "Web UI (Open in Browser)", ActionID: "web"},
		{Title: "Terminal UI (Interactive Live CLI)", ActionID: "terminal"},
		{Title: "Riwayat Sesi & Audit Trail (SQLite)", ActionID: "sessions"},
		{Title: "System & API Key Health Check", ActionID: "health"},
		{Title: "Exit", ActionID: "exit"},
	}

	return LauncherModel{
		ServerURL: serverURL,
		Version:   version,
		Items:     items,
		Cursor:    1, // Default cursor on Terminal UI
	}
}

func (m LauncherModel) Init() tea.Cmd {
	return nil
}

func (m LauncherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			m.Selected = "exit"
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				m.Cursor = len(m.Items) - 1
			}

		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "enter":
			m.Selected = m.Items[m.Cursor].ActionID
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m LauncherModel) View() string {
	if m.Quitting {
		return "\nKeluar dari Niskava Agent. Sampai jumpa!\n"
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("=============================================================================\n")
	b.WriteString(fmt.Sprintf(" %s (%s)\n", headerBoxStyle.Render("Choose Interface"), m.Version))
	b.WriteString(fmt.Sprintf(" 🚀 Server: %s\n", serverURLStyle.Render(m.ServerURL)))
	b.WriteString("=============================================================================\n\n")

	for i, item := range m.Items {
		if i == m.Cursor {
			line := fmt.Sprintf("★ %s", item.Title)
			b.WriteString(selectedItemStyle.Render(line) + "\n")
		} else {
			line := fmt.Sprintf("  ☆ %s", item.Title)
			b.WriteString(unselectedItemStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("[Gunakan panah ↑/↓ atau j/k untuk memilih, Enter untuk mengeksekusi]") + "\n")

	return b.String()
}
