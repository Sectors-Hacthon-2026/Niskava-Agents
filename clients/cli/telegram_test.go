package cli

import (
	"testing"
)

func TestTelegramCLICommandRegistered(t *testing.T) {
	foundTelegramCmd := false
	for _, cmd := range RootCmd.Commands() {
		if cmd.Name() == "telegram" {
			foundTelegramCmd = true
			break
		}
	}
	if !foundTelegramCmd {
		t.Errorf("expected 'telegram' subcommand to be registered in RootCmd")
	}

	telegramFlag := serveCmd.Flags().Lookup("telegram")
	if telegramFlag == nil {
		t.Errorf("expected '--telegram' flag to be present on serveCmd")
	}
}
