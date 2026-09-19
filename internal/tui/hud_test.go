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
	if !strings.Contains(header, "I think, therefore I process") {
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
}
