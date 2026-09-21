package cli

import (
	"testing"
)

func TestRootCmd_SessionFlagRegistered(t *testing.T) {
	flag := RootCmd.PersistentFlags().Lookup("session")
	if flag == nil {
		t.Fatal("expected persistent flag --session to be registered on RootCmd")
	}
	if flag.Shorthand != "s" {
		t.Fatalf("expected shorthand 's' for --session flag, got '%s'", flag.Shorthand)
	}
}
