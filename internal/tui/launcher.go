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

// Styles adhering to PRD Color Palette Specification (#00FF87 Bright Green Theme)
var (
	colorPrdBrightGreen = lipgloss.Color("#00FF87")
	colorPrdMutedGreen  = lipgloss.Color("#1F5C3F")
	colorPrdDarkGreen   = lipgloss.Color("#052E16")
	colorPrdWhite       = lipgloss.Color("#FFFFFF")
	colorPrdLightGray   = lipgloss.Color("#B0B0B0")
	colorPrdDimGray     = lipgloss.Color("#6E6E6E")
	colorPrdStatusOK    = lipgloss.Color("#3DDC97")
	colorPrdStatusErr   = lipgloss.Color("#FF5555")

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
				Foreground(colorPrdBrightGreen).
				Background(colorPrdDarkGreen)

	shortcutKeyInactiveStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(colorPrdBrightGreen)

	itemTitleActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrdWhite).
				Background(colorPrdDarkGreen)

	itemTitleInactiveStyle = lipgloss.NewStyle().
				Foreground(colorPrdLightGray)

	itemDescStyle = lipgloss.NewStyle().
			Foreground(colorPrdDimGray)

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
	items := []LauncherItem{
		{
			ShortcutKey: "W",
			Title:       "Web UI (Open in Browser)",
			Description: "Jalankan server web & buka otomatis di browser default",
			ActionID:    "web",
		},
		{
			ShortcutKey: "T",
			Title:       "Terminal UI (Interactive Live CLI)",
			Description: "Sesi REPL interaktif berbasis perintah riset & anomali",
			ActionID:    "terminal",
		},
		{
			ShortcutKey: "S",
			Title:       "Riwayat Sesi & Audit Trail (SQLite)",
			Description: "Inspeksi riwayat investigasi & bukti terverifikasi dari database",
			ActionID:    "sessions",
		},
		{
			ShortcutKey: "H",
			Title:       "Panduan & Instruksi Penggunaan (Help Guide)",
			Description: "Instruksi lengkap navigasi, opsi menu, dan perintah slash",
			ActionID:    "help",
		},
		{
			ShortcutKey: "C",
			Title:       "System & API Key Health Check",
			Description: "Periksa status daemon server, koneksi database, dan provider AI",
			ActionID:    "health",
		},
		{
			ShortcutKey: "Q",
			Title:       "Quick Setup Wizard (.env)",
			Description: "Konfigurasi cepat API key Sectors, Gemini, atau OpenAI",
			ActionID:    "setup",
		},
		{
			ShortcutKey: "E",
			Title:       "Exit",
			Description: "Hentikan daemon server dan keluar dari Niskava Agent",
			ActionID:    "exit",
		},
	}

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
		case "ctrl+c", "e", "x", "7":
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

		case "q", "6":
			m.Selected = "setup"
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m LauncherModel) View() string {
	if m.Quitting {
		return "\nKeluar dari Niskava Agent. Sampai jumpa!\n"
	}

	noColor := os.Getenv("NO_COLOR") != ""

	var b strings.Builder

	// 1. ASCII Art Banner: NISKAVA (PRD Spec 5.1)
	b.WriteString("\n")
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

	tagline := "Multi-Interface AI Agent Runtime"
	if noColor {
		b.WriteString("  " + tagline + "\n\n")
	} else {
		b.WriteString("  " + taglineStyle.Render(tagline) + "\n\n")
	}

	// 2. Solid Muted Green Separator (PRD Spec 5.2)
	sepWidth := 78
	solidLine := strings.Repeat("─", sepWidth)
	if noColor {
		b.WriteString("  " + solidLine + "\n\n")
	} else {
		b.WriteString("  " + accentBarStyle.Render("▍") + separatorLineStyle.Render(solidLine) + "\n\n")
	}

	// 3. Menu Items List with Shortcuts & Descriptions (PRD Spec 5.3)
	for i, item := range m.Items {
		isActive := i == m.Cursor

		cursor := "  "
		if isActive {
			cursor = "▶ "
		}

		shortcutStr := fmt.Sprintf("[%s]", item.ShortcutKey)

		if noColor {
			if isActive {
				b.WriteString(fmt.Sprintf("%s%s  %-35s\n", cursor, shortcutStr, item.Title))
				b.WriteString(fmt.Sprintf("     %s\n\n", item.Description))
			} else {
				b.WriteString(fmt.Sprintf("  %s  %-35s\n", shortcutStr, item.Title))
				b.WriteString(fmt.Sprintf("     %s\n\n", item.Description))
			}
		} else {
			if isActive {
				cursorR := cursorIndicatorStyle.Render(cursor)
				scR := shortcutKeyActiveStyle.Render(shortcutStr)
				titleR := itemTitleActiveStyle.Render(fmt.Sprintf(" %-40s", item.Title))
				descR := itemDescStyle.Render(fmt.Sprintf("     %s", item.Description))

				b.WriteString(fmt.Sprintf("%s%s %s\n%s\n\n", cursorR, scR, titleR, descR))
			} else {
				scR := shortcutKeyInactiveStyle.Render(shortcutStr)
				titleR := itemTitleInactiveStyle.Render(item.Title)
				descR := itemDescStyle.Render(fmt.Sprintf("     %s", item.Description))

				b.WriteString(fmt.Sprintf("  %s  %s\n%s\n\n", scR, titleR, descR))
			}
		}
	}

	// 4. Solid Separator Line before Status Bar
	if noColor {
		b.WriteString("  " + solidLine + "\n")
	} else {
		b.WriteString("  " + accentBarStyle.Render("▍") + separatorLineStyle.Render(solidLine) + "\n")
	}

	// 5. Persistent Status Bar (PRD Spec 5.4)
	serverHost := strings.TrimPrefix(m.ServerURL, "http://")
	serverHost = strings.TrimPrefix(serverHost, "https://")
	if serverHost == "" {
		serverHost = "localhost:8080"
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

	apiKeyStatusStr := "OK"
	if !m.APIKeyOK {
		apiKeyStatusStr = "Missing/Offline"
	}

	statusContent := fmt.Sprintf(
		" niskava %s  ·  %s Server: %s  ·  %s API Key: %s  ·  ↑/↓ nav  ·  [W/T/S/H/C/Q/E] select ",
		m.Version,
		serverDot,
		serverHost,
		apiKeyDot,
		apiKeyStatusStr,
	)

	if noColor {
		b.WriteString("\n" + statusContent + "\n\n")
	} else {
		b.WriteString("\n" + statusBarBgStyle.Render(statusContent) + "\n\n")
	}

	return b.String()
}
