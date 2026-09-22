package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunSubprocessMock(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	curr := wd
	rootDir := ""
	for {
		if _, err := os.Stat(filepath.Join(curr, "go.mod")); err == nil {
			rootDir = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	pythonBin := filepath.Join(rootDir, ".venv", "Scripts", "python.exe")
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		pythonBin = filepath.Join(rootDir, ".venv", "bin", "python3")
		if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
			if path, err := exec.LookPath("python"); err == nil {
				pythonBin = path
			} else if path, err := exec.LookPath("python3"); err == nil {
				pythonBin = path
			} else {
				pythonBin = "python"
			}
		}
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

func TestScannerHandlesLargeJSONLine(t *testing.T) {
	// Create a large JSON event line (>64KB, e.g., 70KB)
	largeContent := strings.Repeat("x", 70*1024)
	ev := Event{
		Event:   EventAgentMessageChunk,
		Content: largeContent,
	}
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("failed to marshal large event: %v", err)
	}
	if len(data) <= 64*1024 {
		t.Fatalf("expected marshaled json to be > 64KB, got %d bytes", len(data))
	}

	reader := strings.NewReader(string(data) + "\n")
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, ipcScannerBufferBytes), ipcScannerBufferBytes)

	if !scanner.Scan() {
		t.Fatalf("scanner failed to scan large line: %v", scanner.Err())
	}

	var parsed Event
	if err := json.Unmarshal(scanner.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal scanned event: %v", err)
	}

	if parsed.Content != largeContent {
		t.Errorf("content mismatch: got length %d, want %d", len(parsed.Content), len(largeContent))
	}
}
