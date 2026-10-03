// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionGatewayModel is an interactive sub-menu model to choose between Chat Sessions & Investigation Audit Trails.
type SessionGatewayModel struct {
	Cursor   int
	Selected string // "chats", "investigations", or "" (canceled)
	Canceled bool
	Width    int
	Height   int
}

// NewSessionGatewayModel creates a new session category choice model.
func NewSessionGatewayModel() SessionGatewayModel {
	return SessionGatewayModel{
		Cursor: 0,
		Width:  GetTermWidth(),
		Height: GetTermHeight(),
	}
}

func (m SessionGatewayModel) Init() tea.Cmd {
	return nil
}

func (m SessionGatewayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		k := strings.ToLower(msg.String())
		switch k {
		case "ctrl+c", "esc", "q":
			m.Canceled = true
			m.Selected = ""
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				m.Cursor = 1
			}

		case "down", "j":
			if m.Cursor < 1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "1", "c":
			m.Selected = "chats"
			return m, tea.Quit

		case "2", "i":
			m.Selected = "investigations"
			return m, tea.Quit

		case "enter":
			if m.Cursor == 0 {
				m.Selected = "chats"
			} else {
				m.Selected = "investigations"
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SessionGatewayModel) View() string {
	var b strings.Builder

	termW := m.Width
	if termW <= 0 {
		termW = GetTermWidth()
	}
	boxW := termW - 4
	if boxW > termW-2 {
		boxW = termW - 2
	}
	if boxW < 16 {
		boxW = max(10, termW-2)
	}

	sessionBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Width(boxW).
		Padding(0, 1).
		Foreground(ColorFg)

	title := T("sessions_gateway_title")
	b.WriteString(sessionTitleStyle.Render(title))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(T("sessions_gateway_prompt")))
	b.WriteString("\n\n")

	options := []struct {
		id   string
		title string
		desc  string
		key   string
	}{
		{
			id:    "chats",
			title: T("sessions_gateway_opt_chats"),
			desc:  T("sessions_gateway_opt_chats_desc"),
			key:   "[1]",
		},
		{
			id:    "investigations",
			title: T("sessions_gateway_opt_inv"),
			desc:  T("sessions_gateway_opt_inv_desc"),
			key:   "[2]",
		},
	}

	for i, opt := range options {
		var lineTitle string
		lineTitle = fmt.Sprintf("  %s %s", opt.key, opt.title)
		descLine := fmt.Sprintf("      ↳ %s", opt.desc)

		if i == m.Cursor {
			b.WriteString(sessionCursorStyle.Render("> "))
			b.WriteString(sessionActiveStyle.Render(lineTitle))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorFg).Render(descLine))
		} else {
			b.WriteString(sessionMetaStyle.Render("  "))
			b.WriteString(sessionMetaStyle.Render(lineTitle))
			b.WriteString("\n")
			b.WriteString(sessionMetaStyle.Render(descLine))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render(
		T("sessions_gateway_footer_hint"),
	))

	return "\n" + sessionBoxStyle.Render(b.String()) + "\n\033[J"
}
