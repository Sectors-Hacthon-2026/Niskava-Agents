package tui

import (
	"strings"
	"testing"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
)

func TestRenderASCIIAnomalyChart_Empty(t *testing.T) {
	out := RenderASCIIAnomalyChart("ANTM", nil, 30)
	if !strings.Contains(out, "ANTM") {
		t.Errorf("Expected output to contain ticker ANTM, got: %s", out)
	}
}

func TestRenderASCIIAnomalyChart_WithAnomalies(t *testing.T) {
	anomalies := []ipc.Event{
		{
			Ticker:         "ANTM",
			AnomalyDate:    "2026-09-12",
			MetricType:     "Volume Z-Score Spike",
			ZScore:         3.82,
			PriceChangePct: 8.5,
			MetricValue:    14500000,
			BaselineValue:  3800000,
			Description:    "Unusual volume spike 3.82x above 20-day moving average",
		},
	}

	out := RenderASCIIAnomalyChart("ANTM", anomalies, 30)
	if !strings.Contains(out, "3.82σ") {
		t.Errorf("Expected output to contain Z-Score 3.82σ, got: %s", out)
	}
	if !strings.Contains(out, "+8.50%") {
		t.Errorf("Expected output to contain price change +8.50%%, got: %s", out)
	}
}
