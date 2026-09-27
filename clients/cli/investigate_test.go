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

func TestInvestigateCmd_ExportFlagsRegistered(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"investigate"})
	if err != nil || cmd == nil {
		t.Fatalf("expected 'investigate' subcommand to be registered, got err: %v", err)
	}

	fmtFlag := cmd.Flags().Lookup("export-format")
	if fmtFlag == nil || fmtFlag.Shorthand != "f" {
		t.Fatalf("expected --export-format (-f) flag to be registered")
	}

	outFlag := cmd.Flags().Lookup("export-out")
	if outFlag == nil || outFlag.Shorthand != "o" {
		t.Fatalf("expected --export-out (-o) flag to be registered")
	}
}
