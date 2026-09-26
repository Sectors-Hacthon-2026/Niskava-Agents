package cli

import (
	"bytes"
	"testing"
)

func TestCompletionCommand_Bash(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"completion", "bash"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("bash completion")) {
		t.Errorf("expected bash completion header in output, got: %s", buf.String())
	}
}

func TestCompletionCommand_Zsh(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"completion", "zsh"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}
	if len(buf.Bytes()) == 0 {
		t.Errorf("expected non-empty zsh completion output")
	}
}
