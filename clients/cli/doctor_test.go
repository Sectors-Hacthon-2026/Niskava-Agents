package cli

import (
	"testing"
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
