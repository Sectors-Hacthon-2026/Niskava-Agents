package server

import (
	"strings"
	"testing"
)

func TestRenderWorkspaceHTML(t *testing.T) {
	html := RenderWorkspaceHTML(20128)
	if len(html) < 100000 {
		t.Fatalf("expected workspace HTML to be > 100KB, got %d bytes", len(html))
	}
	expectedStrings := []string{
		"NISKAVA — AI for Brighter Investments",
		"lightweight-charts",
		"btnToggleSidebar",
		"btnSidebarClose",
		"openSettingsModal",
		"loadChatSessions",
		"data-theme",
		"Plus Jakarta Sans",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(html, s) {
			t.Errorf("missing expected string in rendered HTML: %s", s)
		}
	}
}
