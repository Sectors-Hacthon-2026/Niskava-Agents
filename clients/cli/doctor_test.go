package cli

import (
	"strings"
	"testing"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
)

func TestEvaluateSystemDiagnostics(t *testing.T) {
	report := EvaluateSystemDiagnostics()
	if report.OS == "" || report.Arch == "" {
		t.Errorf("expected OS and Arch to be populated, got: %+v", report)
	}
	if len(report.Checks) == 0 {
		t.Error("expected at least 1 diagnostic check")
	}
}

func TestEvaluateSystemDiagnostics_Telegram(t *testing.T) {
	// 1. With empty token -> should report StatusOk (optional)
	emptyCfg := config.DefaultConfig()
	emptyCfg.Telegram.BotToken = ""
	reportEmpty := evaluateTelegramDiagnostic(emptyCfg, nil)
	if reportEmpty.Status != StatusOk || !strings.Contains(reportEmpty.Details, "Optional") {
		t.Errorf("expected StatusOk with optional note for empty token, got: %+v", reportEmpty)
	}

	// 2. With valid mock bot info -> should report StatusOk with bot username
	validCfg := config.DefaultConfig()
	validCfg.Telegram.BotToken = "dummy_token"
	validCfg.Telegram.AllowedUsers = []string{"user1", "user2"}
	mockInfo := &TelegramBotInfo{Username: "idx_advisor_bot", ID: 12345}
	reportValid := evaluateTelegramDiagnosticWithBotInfo(validCfg, mockInfo, nil)
	if reportValid.Status != StatusOk || !strings.Contains(reportValid.Details, "@idx_advisor_bot") {
		t.Errorf("expected StatusOk with @idx_advisor_bot, got: %+v", reportValid)
	}
	if !strings.Contains(reportValid.Details, "2 whitelisted user(s)") {
		t.Errorf("expected 2 whitelisted users, got: %s", reportValid.Details)
	}
}
