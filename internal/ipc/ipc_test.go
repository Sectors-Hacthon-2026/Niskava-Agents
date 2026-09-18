package ipc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunSubprocessMock(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	rootDir := filepath.Clean(filepath.Join(wd, "..", ".."))

	pythonBin := filepath.Join(rootDir, ".venv", "bin", "python3")
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		pythonBin = "python3"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	params := RunnerParams{
		PythonBin: pythonBin,
		WorkDir:   rootDir,
		DBPath:    dbPath,
		Ticker:    "ANTM",
		Days:      30,
		Offline:   true,
	}

	eventsChan, errChan := RunSubprocess(ctx, params)

	var receivedEvents []Event
	for ev := range eventsChan {
		receivedEvents = append(receivedEvents, ev)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("subprocess failed with error: %v", err)
	}

	if len(receivedEvents) == 0 {
		t.Fatalf("expected events, got 0")
	}

	// Verify that session_start and session_complete are present
	hasStart := false
	hasComplete := false
	for _, ev := range receivedEvents {
		if ev.Event == EventSessionStart {
			hasStart = true
		}
		if ev.Event == EventSessionComplete {
			hasComplete = true
		}
	}

	if !hasStart || !hasComplete {
		t.Errorf("missing start or complete events. Got: %+v", receivedEvents)
	}
}
