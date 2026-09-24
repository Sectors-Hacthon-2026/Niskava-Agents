package cli

import (
	"testing"
)

func TestGraphCmdFlags(t *testing.T) {
	flags := []string{"output", "session", "ticker", "depth", "text", "prune"}
	for _, f := range flags {
		if graphCmd.Flags().Lookup(f) == nil {
			t.Errorf("Expected flag --%s on graphCmd", f)
		}
	}
}
