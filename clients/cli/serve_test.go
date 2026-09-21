package cli

import (
	"testing"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/config"
)

func TestDefaultServerPortConstant(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Server.Port != 20128 {
		t.Errorf("expected default port to be 20128 per integration guide, got %d", cfg.Server.Port)
	}
}
