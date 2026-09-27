package cli

import (
	"testing"
)

func TestTerminalCmd_RegistrationAndAliases(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"terminal"})
	if err != nil || cmd == nil {
		t.Fatalf("expected 'terminal' subcommand to be registered on RootCmd, got err: %v", err)
	}

	if cmd.Name() != "terminal" {
		t.Fatalf("expected command name 'terminal', got '%s'", cmd.Name())
	}

	aliases := cmd.Aliases
	hasRepl := false
	hasChat := false
	for _, alias := range aliases {
		if alias == "repl" {
			hasRepl = true
		}
		if alias == "chat" {
			hasChat = true
		}
	}

	if !hasRepl || !hasChat {
		t.Fatalf("expected aliases ['repl', 'chat'], got %v", aliases)
	}
}
