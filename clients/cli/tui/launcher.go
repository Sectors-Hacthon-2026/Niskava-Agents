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
	Width     int
	Height    int
}

// Styles adhering to Binance Dark Financial Intelligence Palette Specification (#FCD535 Gold / #1E2329 Dark Slate)
var (
	colorPrdBrightGreen = lipgloss.Color("#FCD535") // Financial Gold Accent
	colorPrdMutedGreen  = lipgloss.Color("#848E9C") // Slate Gray
	colorPrdDarkGreen   = lipgloss.Color("#1E2329") // Dark Slate
	colorPrdWhite       = lipgloss.Color("#FFFFFF") // Pure White
	colorPrdLightGray   = lipgloss.Color("#848E9C") // Muted Slate Gray
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
	return NewLauncherModelWithHealthAndCursor(serverURL, version, apiKeyOK, 0)
}

// NewLauncherModelWithHealthAndCursor initializes launcher menu with custom initial cursor position.
func NewLauncherModelWithHealthAndCursor(serverURL string, version string, apiKeyOK bool, initialCursor int) LauncherModel {
	items := GetLocalizedLauncherItems()
	if initialCursor < 0 || initialCursor >= len(items) {
		initialCursor = 0
	}

	return LauncherModel{
		ServerURL: serverURL,
		Version:   version,
		APIKeyOK:  apiKeyOK,
		Items:     items,
		Cursor:    initialCursor,
	}
}

func (m LauncherModel) Init() tea.Cmd {
	return nil
}

func (m LauncherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

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

	w := m.Width
	if w <= 0 {
		w = GetTermWidth()
	}

	noColor := os.Getenv("NO_COLOR") != ""

	var b strings.Builder

	// 1. Responsive ASCII Art Banner: NISKAVA
	b.WriteString(RenderResponsiveASCIIHeader(w, bannerStyle))

	tagline := T("launcher_tagline")
	if noColor {
		b.WriteString("  ")
		b.WriteString(tagline)
		b.WriteString("\n")
	} else {
		b.WriteString("  ")
		b.WriteString(taglineStyle.Render(tagline))
		b.WriteString("\n")
	}

	// 2. Solid Muted Green Separator
	sepWidth := w - 4
	if sepWidth > w-2 {
		sepWidth = w - 2
	}
	if sepWidth < 10 {
		sepWidth = max(5, w-2)
	}
	solidLine := strings.Repeat("─", sepWidth)
	if noColor {
		b.WriteString("  ")
		b.WriteString(solidLine)
		b.WriteString("\n")
	} else {
		b.WriteString("  ")
		b.WriteString(accentBarStyle.Render("▍"))
		b.WriteString(separatorLineStyle.Render(solidLine))
		b.WriteString("\n")
	}

	// 3. Compact Menu Items List (Only active item displays description to fit within 24-line terminal)
	titleW := w - 10
	if titleW < 15 {
		titleW = 15
	}
	descW := w - 8
	if descW < 15 {
		descW = 15
	}

	for i, item := range m.Items {
		isActive := i == m.Cursor
		shortcutStr := fmt.Sprintf("[%s]", item.ShortcutKey)
		tVal := Truncate(item.Title, titleW)

		if noColor {
			if isActive {
				b.WriteString(fmt.Sprintf("▶ %s  %s\n", shortcutStr, tVal))
				b.WriteString(fmt.Sprintf("     %s\n", Truncate(item.Description, descW)))
			} else {
				b.WriteString(fmt.Sprintf("  %s  %s\n", shortcutStr, tVal))
			}
		} else {
			if isActive {
				cursorR := cursorIndicatorStyle.Render("▶ ")
				scR := shortcutKeyActiveStyle.Render(shortcutStr)
				titleR := itemTitleActiveStyle.Render(fmt.Sprintf(" %s ", tVal))
				descR := itemDescStyle.Render(fmt.Sprintf("     %s", Truncate(item.Description, descW)))

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
		b.WriteString("  ")
		b.WriteString(solidLine)
		b.WriteString("\n")
	} else {
		b.WriteString("  ")
		b.WriteString(accentBarStyle.Render("▍"))
		b.WriteString(separatorLineStyle.Render(solidLine))
		b.WriteString("\n")
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

	var statusContent string
	if w < 55 {
		statusContent = fmt.Sprintf("v%s • %s %s", m.Version, apiKeyDot, apiKeyStatusStr)
	} else {
		statusContent = TF(
			"launcher_status_bar",
			m.Version,
			serverDot,
			serverHost,
			apiKeyDot,
			apiKeyStatusStr,
		)
	}

	if noColor {
		b.WriteString(statusContent)
		b.WriteString("\n")
	} else {
		b.WriteString(statusBarBgStyle.Render(statusContent))
		b.WriteString("\n")
	}

	return b.String() + "\033[J"
}

// PromptEscReturnModel is a Bubbletea sub-model that prompts the user to press ESC or Enter to return to main menu.
type PromptEscReturnModel struct{}

func (m PromptEscReturnModel) Init() tea.Cmd {
	return nil
}

func (m PromptEscReturnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "q", "space", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PromptEscReturnModel) View() string {
	return "\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(T("menu_press_enter")) + "\n\033[J"
}

// PromptPressEscToReturn renders "Press ESC to return to Menu..." and waits for keypress.
func PromptPressEscToReturn() {
	p := tea.NewProgram(PromptEscReturnModel{})
	_, _ = p.Run()
}
