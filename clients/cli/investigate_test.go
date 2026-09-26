package cli

import (
	"testing"
)

func TestInvestigateCmd_InteractiveFlagRegistered(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"investigate"})
	if err != nil || cmd == nil {
		t.Fatalf("expected 'investigate' subcommand to be registered on RootCmd, got err: %v", err)
	}

	flag := cmd.Flags().Lookup("interactive")
	if flag == nil {
		t.Fatal("expected flag --interactive to be registered on investigateCmd")
	}

	if flag.Shorthand != "i" {
		t.Fatalf("expected shorthand 'i' for --interactive flag, got '%s'", flag.Shorthand)
	}
}
