package tui

import (
	"strings"
	"testing"
)

func TestRenderConstellationLine(t *testing.T) {
	line := RenderConstellationLine(80)
	if line == "" {
		t.Fatalf("expected non-empty constellation line")
	}
	if !strings.Contains(line, "─") {
		t.Errorf("expected constellation line to contain horizontal bar character ─")
	}
}

func TestRenderHUDHeader(t *testing.T) {
	header := RenderHUDHeader("gemini-2.5-flash", "http://localhost:8080", "", "TEST-SESSION-001")
	if !strings.Contains(header, "Investigate them") {
		t.Errorf("expected header to contain motto tagline")
	}
	if !strings.Contains(header, "DESIGNATION") {
		t.Errorf("expected header to contain DESIGNATION metadata label")
	}
	if !strings.Contains(header, "BRAIN SIZE") {
		t.Errorf("expected header to contain BRAIN SIZE metadata label")
	}
	if !strings.Contains(header, "PURPOSE") {
		t.Errorf("expected header to contain PURPOSE metadata label")
	}

	defaultHeader := RenderHUDHeader("", "http://localhost:8080", "", "TEST-SESSION-002")
	if !strings.Contains(defaultHeader, "niskava") {
		t.Errorf("expected default header to contain niskava as default model substrate")
	}
}

func TestRenderResponsiveASCIIHeader(t *testing.T) {
	// Wide terminal (>= 48 cols) -> 45-col double-line block ASCII banner
	wideHeader := RenderResponsiveASCIIHeader(80, hudTitleStyle)
	if !strings.Contains(wideHeader, "███╗ ██╗") {
		t.Errorf("expected wide terminal to render double-line block ASCII banner")
	}

	// Compact terminal (36 to 47 cols) -> 28-col mini block ASCII banner
	medHeader := RenderResponsiveASCIIHeader(40, hudTitleStyle)
	if !strings.Contains(medHeader, "█▄ █") {
		t.Errorf("expected medium terminal to render 28-col mini block ASCII banner")
	}

	// Narrow terminal (30 to 35 cols) -> Compact styled pill title
	narrowHeader := RenderResponsiveASCIIHeader(32, hudTitleStyle)
	if !strings.Contains(narrowHeader, "NISKAVA AGENT") {
		t.Errorf("expected narrow terminal to render compact pill title")
	}

	// Ultra-narrow terminal (< 30 cols) -> Minimal pill title
	ultraNarrowHeader := RenderResponsiveASCIIHeader(25, hudTitleStyle)
	if !strings.Contains(ultraNarrowHeader, "NISKAVA") {
		t.Errorf("expected ultra-narrow terminal to render minimal pill title")
	}
}

func TestGetFullHelpGuideString_ResponsiveWrapping(t *testing.T) {
	// Test help guide rendering at narrow terminal width (45 cols)
	narrowHelp := GetFullHelpGuideString(45)
	if narrowHelp == "" {
		t.Fatalf("expected help guide string not to be empty")
	}
	if !strings.Contains(narrowHelp, "NISKAVA AGENT") {
		t.Errorf("expected help guide to contain title header")
	}
	if !strings.Contains(narrowHelp, "[W] / [1]") {
		t.Errorf("expected help guide to contain direct hotkey items")
	}
}
