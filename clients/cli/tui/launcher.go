// Package tui provides interactive terminal interfaces including the PRD-revamped
// launcher menu and the persistent live REPL.
package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LauncherItem represents a selectable menu choice with shortcut hotkey and description.
type LauncherItem struct {
	ShortcutKey string
	Title       string
	Description string
	ActionID    string
}

// LauncherModel is the Bubbletea model for the revamped interface selector menu.
type LauncherModel struct {
	ServerURL string
	Version   string
	APIKeyOK  bool
	Items     []LauncherItem
	Cursor    int
	Selected  string
	Quitting  bool
}

// Styles adhering to Binance Dark Financial OSINT Palette Specification (#FCD535 Gold / #1E2329 Dark Slate)
var (
	colorPrdBrightGreen = lipgloss.Color("#FCD535") // Financial Gold Accent
	colorPrdMutedGreen  = lipgloss.Color("#848E9C") // Slate Gray
	colorPrdDarkGreen   = lipgloss.Color("#1E2329") // Dark Slate
	colorPrdWhite       = lipgloss.Color("#FFFFFF") // Pure White
	colorPrdLightGray   = lipgloss.Color("#848E9C") // Muted Slate Gray
	colorPrdDimGray     = lipgloss.Color("#848E9C") // Dimmed Slate Gray
	colorPrdStatusOK    = lipgloss.Color("#0ECB81") // Financial Green OK
	colorPrdStatusErr   = lipgloss.Color("#F6465D") // Financial Red Error

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrdBrightGreen)

	taglineStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(colorPrdLightGray)

	accentBarStyle = lipgloss.NewStyle().
			Foreground(colorPrdBrightGreen)

	separatorLineStyle = lipgloss.NewStyle().
				Foreground(colorPrdMutedGreen)

	shortcutKeyActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrdDarkGreen).
				Background(colorPrdBrightGreen)

	shortcutKeyInactiveStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(colorPrdBrightGreen)

	itemTitleActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrdDarkGreen).
				Background(colorPrdBrightGreen)

	itemTitleInactiveStyle = lipgloss.NewStyle().
				Foreground(colorPrdWhite)

	itemDescStyle = lipgloss.NewStyle().
			Foreground(colorPrdLightGray)

	cursorIndicatorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrdBrightGreen)

	statusBarBgStyle = lipgloss.NewStyle().
				Foreground(colorPrdWhite).
				Background(colorPrdDarkGreen).
				Padding(0, 1)

	statusOKDotStyle = lipgloss.NewStyle().
				Foreground(colorPrdStatusOK)

	statusErrDotStyle = lipgloss.NewStyle().
				Foreground(colorPrdStatusErr)
)

// NewLauncherModel initializes the revamped launcher menu.
func NewLauncherModel(serverURL string, version string) LauncherModel {
	return NewLauncherModelWithHealth(serverURL, version, true)
}

// NewLauncherModelWithHealth initializes the revamped launcher menu with live health state.
func NewLauncherModelWithHealth(serverURL string, version string, apiKeyOK bool) LauncherModel {
	items := GetLocalizedLauncherItems()

	return LauncherModel{
		ServerURL: serverURL,
		Version:   version,
		APIKeyOK:  apiKeyOK,
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
		k := strings.ToLower(msg.String())
		switch k {
		case "ctrl+c", "e", "x", "8":
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

		// Direct Hotkey Shortcuts (PRD Spec 5.3)
		case "w", "1":
			m.Selected = "web"
			return m, tea.Quit

		case "t", "2":
			m.Selected = "terminal"
			return m, tea.Quit

		case "s", "3":
			m.Selected = "sessions"
			return m, tea.Quit

		case "h", "4":
			m.Selected = "help"
			return m, tea.Quit

		case "c", "5":
			m.Selected = "health"
			return m, tea.Quit

		case "l", "6":
			m.Selected = "lang"
			return m, tea.Quit

		case "q", "7":
			m.Selected = "setup"
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m LauncherModel) View() string {
	if m.Quitting {
		return T("launcher_quitting_msg")
	}

	noColor := os.Getenv("NO_COLOR") != ""

	var b strings.Builder

	// 1. ASCII Art Banner: NISKAVA
	asciiLines := []string{
		"███╗   ██╗██╗███████╗██╗  ██╗██████╗  ██╗   ██╗██████╗ ",
		"████╗  ██║██║██╔════╝██║ ██╔╝██╔══██╗ ██║   ██║██╔══██╗",
		"██╔██╗ ██║██║███████╗█████═╝ ███████║ ██║   ██║██████╔╝",
		"██║╚██╗██║██║╚════██║██╔═██╗ ██╔══██║ ╚██╗ ██╔╝██╔══██║",
		"██║ ╚████║██║███████║██║  ██╗██║  ██║  ╚████╔╝ ██║  ██║",
		"╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═══╝  ╚═╝  ╚═╝",
	}

	for _, line := range asciiLines {
		if noColor {
			b.WriteString(line + "\n")
		} else {
			b.WriteString(bannerStyle.Render(line) + "\n")
		}
	}

	tagline := T("launcher_tagline")
	if noColor {
		b.WriteString("  " + tagline + "\n")
	} else {
		b.WriteString("  " + taglineStyle.Render(tagline) + "\n")
	}

	// 2. Solid Muted Green Separator
	sepWidth := 78
	solidLine := strings.Repeat("─", sepWidth)
	if noColor {
		b.WriteString("  " + solidLine + "\n")
	} else {
		b.WriteString("  " + accentBarStyle.Render("▍") + separatorLineStyle.Render(solidLine) + "\n")
	}

	// 3. Compact Menu Items List (Only active item displays description to fit within 24-line terminal)
	for i, item := range m.Items {
		isActive := i == m.Cursor
		shortcutStr := fmt.Sprintf("[%s]", item.ShortcutKey)

		if noColor {
			if isActive {
				b.WriteString(fmt.Sprintf("▶ %s  %-35s\n", shortcutStr, item.Title))
				b.WriteString(fmt.Sprintf("     %s\n", item.Description))
			} else {
				b.WriteString(fmt.Sprintf("  %s  %-35s\n", shortcutStr, item.Title))
			}
		} else {
			if isActive {
				cursorR := cursorIndicatorStyle.Render("▶ ")
				scR := shortcutKeyActiveStyle.Render(shortcutStr)
				titleR := itemTitleActiveStyle.Render(fmt.Sprintf(" %-40s", item.Title))
				descR := itemDescStyle.Render(fmt.Sprintf("     %s", item.Description))

				b.WriteString(fmt.Sprintf("%s%s %s\n%s\n", cursorR, scR, titleR, descR))
			} else {
				scR := shortcutKeyInactiveStyle.Render(shortcutStr)
				titleR := itemTitleInactiveStyle.Render(item.Title)

				b.WriteString(fmt.Sprintf("  %s  %s\n", scR, titleR))
			}
		}
	}

	// 4. Solid Separator Line before Status Bar
	if noColor {
		b.WriteString("  " + solidLine + "\n")
	} else {
		b.WriteString("  " + accentBarStyle.Render("▍") + separatorLineStyle.Render(solidLine) + "\n")
	}

	// 5. Persistent Status Bar
	serverHost := strings.TrimPrefix(m.ServerURL, "http://")
	serverHost = strings.TrimPrefix(serverHost, "https://")
	if serverHost == "" {
		serverHost = "localhost:20128"
	}

	var (
		serverDot string
		apiKeyDot string
	)

	if noColor {
		serverDot = "●"
		apiKeyDot = "●"
	} else {
		serverDot = statusOKDotStyle.Render("●")
		if m.APIKeyOK {
			apiKeyDot = statusOKDotStyle.Render("●")
		} else {
			apiKeyDot = statusErrDotStyle.Render("●")
		}
	}

	apiKeyStatusStr := T("launcher_status_api_ok")
	if !m.APIKeyOK {
		apiKeyStatusStr = T("launcher_status_api_missing")
	}

	statusContent := TF(
		"launcher_status_bar",
		m.Version,
		serverDot,
		serverHost,
		apiKeyDot,
		apiKeyStatusStr,
	)

	if noColor {
		b.WriteString(statusContent + "\n")
	} else {
		b.WriteString(statusBarBgStyle.Render(statusContent) + "\n")
	}

	return b.String()
}
