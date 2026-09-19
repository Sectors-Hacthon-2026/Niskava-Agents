package server

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestServerStartAndHealth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := Start(ctx, 0, nil) // 0 binds to a free port
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("failed to GET health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}
}
