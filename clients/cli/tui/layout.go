package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Breakpoint Tier definitions for Niskava TUI layout adaptation
type BreakpointTier int

const (
	TierCompact BreakpointTier = iota // < 50 cols
	TierNarrow                        // 50 - 79 cols
	TierNormal                        // 80 - 120 cols
	TierWide                          // > 120 cols
)

// GetBreakpointTier returns the BreakpointTier based on terminal width.
func GetBreakpointTier(width int) BreakpointTier {
	if width < 50 {
		return TierCompact
	}
	if width < 80 {
		return TierNarrow
	}
	if width <= 120 {
		return TierNormal
	}
	return TierWide
}

// GetTermWidth returns the current terminal width with safe clamping (min 30).
func GetTermWidth() int {
	if f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0); err == nil {
		defer f.Close()
		if w, _, err := term.GetSize(uintptr(f.Fd())); err == nil && w > 0 {
			if w < 30 {
				return 30
			}
			return w
		}
	}
	if w, _, err := term.GetSize(uintptr(os.Stdout.Fd())); err == nil && w > 0 {
		if w < 30 {
			return 30
		}
		return w
	}
	if w, _, err := term.GetSize(uintptr(os.Stdin.Fd())); err == nil && w > 0 {
		if w < 30 {
			return 30
		}
		return w
	}
	if w, _, err := term.GetSize(uintptr(os.Stderr.Fd())); err == nil && w > 0 {
		if w < 30 {
			return 30
		}
		return w
	}
	return 80
}

// GetTermHeight returns the current terminal height with safe clamping (min 10).
func GetTermHeight() int {
	if f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0); err == nil {
		defer f.Close()
		if _, h, err := term.GetSize(uintptr(f.Fd())); err == nil && h > 0 {
			if h < 10 {
				return 10
			}
			return h
		}
	}
	if _, h, err := term.GetSize(uintptr(os.Stdout.Fd())); err == nil && h > 0 {
		if h < 10 {
			return 10
		}
		return h
	}
	return 24
}

// Sep renders a horizontal line separator adjusted to terminal width minus margin.
func Sep(margin int, overrideWidth ...int) string {
	w := GetTermWidth()
	if len(overrideWidth) > 0 && overrideWidth[0] > 0 {
		w = overrideWidth[0]
	}
	lineLen := w - margin
	if lineLen < 10 {
		lineLen = 10
	}
	lineStr := strings.Repeat("─", lineLen)
	return lipgloss.NewStyle().Foreground(ColorMuted).Render(lineStr)
}

// Truncate safely cuts a string so its visual width (including ANSI) does not exceed maxLen.
func Truncate(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	if lipgloss.Width(s) <= maxLen {
		return s
	}

	// Iterate characters until visual width fits maxLen - 1 (for ellipsis)
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i]) + "…"
		if lipgloss.Width(candidate) <= maxLen {
			return candidate
		}
	}
	return string(runes[:maxLen-1]) + "…"
}

// TruncateMiddle safely truncates a long string/path from the middle (e.g. C:\...\file.txt).
func TruncateMiddle(s string, maxLen int) string {
	if maxLen <= 5 || lipgloss.Width(s) <= maxLen {
		return s
	}
	half := (maxLen - 3) / 2
	runes := []rune(s)
	left := string(runes[:half])
	right := string(runes[len(runes)-half:])
	return left + "..." + right
}

// AtomicBadge renders a styled badge and ensures it is never split across lines.
func AtomicBadge(text string, style lipgloss.Style) string {
	return style.Inline(true).Render(text)
}
