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
		"NISKAVA - AI for Brighter Investments",
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

// TestWorkspaceDossierHasNoFabricatedData guards the investigation dossier against
// regressing to hardcoded/random placeholder data presented as real market evidence.
func TestWorkspaceDossierHasNoFabricatedData(t *testing.T) {
	html := RenderWorkspaceHTML(20128)

	required := []string{
		`id="invChartSourceBadge"`,            // data provenance badge
		"/timeline`",                          // dynamic timeline fetch
		"function renderDossierEvidenceEmpty", // honest empty state for evidence
	}
	for _, s := range required {
		if !strings.Contains(html, s) {
			t.Errorf("missing required dossier element: %s", s)
		}
	}

	forbidden := []string{
		"RADAR_TICKER_PROFILES",       // hardcoded per-ticker fake Z-scores
		"getTickerProfileFallback",    // seeded fake stats for unknown tickers
		"Hari ini, 09:05",             // static fake timeline timestamps
		"Today, 09:05",                // static fake timeline timestamps (EN)
		"(Math.random() - 0.48) * 30", // random candle generator in chart renderer
	}
	for _, s := range forbidden {
		if strings.Contains(html, s) {
			t.Errorf("fabricated dossier data still present: %s", s)
		}
	}
}

// TestWorkspaceChatIsLeanAndFast asserts that the chat interface remains lean, responsive,
// and blazingly fast by keeping TradingView canvas charts decoupled from the chat message DOM.
func TestWorkspaceChatIsLeanAndFast(t *testing.T) {
	html := RenderWorkspaceHTML(20128)

	// Invariants: Chat message DOM must NOT inject inline candlestick canvases or heavy chart renderers
	// during streaming, as layout recalculations degrade agent responsiveness.
	forbidden := []string{
		"chat-inline-market-card",     // no inline chart containers in chat messages
		"renderChatInlineCandlestick", // no inline chart renderer execution in chat
	}
	for _, s := range forbidden {
		if strings.Contains(html, s) {
			t.Errorf("chat interface contains bloat element that degrades streaming speed: %s", s)
		}
	}

	// Required: Chat retains high-performance markdown, ReAct accordion, and dossier navigation
	required := []string{
		"react-steps-box",     // ReAct accordion for thought progression
		"markdown-rendered",   // lightweight streaming markdown rendering
		"openDossierFromChat", // navigation bridge to full visual dossier tab
	}
	for _, s := range required {
		if !strings.Contains(html, s) {
			t.Errorf("missing required lightweight chat element: %s", s)
		}
	}
}
