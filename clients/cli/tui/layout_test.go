package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetBreakpointTier(t *testing.T) {
	tests := []struct {
		width    int
		expected BreakpointTier
	}{
		{40, TierCompact},
		{49, TierCompact},
		{50, TierNarrow},
		{79, TierNarrow},
		{80, TierNormal},
		{120, TierNormal},
		{121, TierWide},
		{160, TierWide},
	}

	for _, tt := range tests {
		got := GetBreakpointTier(tt.width)
		if got != tt.expected {
			t.Errorf("GetBreakpointTier(%d) = %v; want %v", tt.width, got, tt.expected)
		}
	}
}

func TestSep(t *testing.T) {
	sep := Sep(2, 80)
	visualWidth := lipgloss.Width(sep)
	if visualWidth != 78 {
		t.Errorf("Sep(2, 80) visual width = %d; want 78", visualWidth)
	}

	sepCompact := Sep(2, 20)
	if lipgloss.Width(sepCompact) < 10 {
		t.Errorf("Sep(2, 20) visual width = %d; want >= 10", lipgloss.Width(sepCompact))
	}
}

func TestTruncate(t *testing.T) {
	longStr := "Niskava Agent Financial Intelligence Terminal"
	truncated := Truncate(longStr, 20)
	if lipgloss.Width(truncated) > 20 {
		t.Errorf("Truncate visual width = %d; want <= 20", lipgloss.Width(truncated))
	}

	shortStr := "ANTM"
	if Truncate(shortStr, 20) != "ANTM" {
		t.Errorf("Truncate short string modified unexpectedly")
	}
}

func TestTruncateMiddle(t *testing.T) {
	path := "C:\\Users\\Admin\\AppData\\Local\\Microsoft\\WindowsApps\\python3.exe"
	truncated := TruncateMiddle(path, 30)
	if lipgloss.Width(truncated) > 30 {
		t.Errorf("TruncateMiddle visual width = %d; want <= 30", lipgloss.Width(truncated))
	}
}

