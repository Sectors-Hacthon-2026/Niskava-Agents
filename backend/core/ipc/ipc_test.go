package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunSubprocessMock(t *testing.T) {
	t.Setenv("NISKAVA_TESTING", "1")
	t.Setenv("MOCK_SECTORS", "1")
	t.Setenv("SECTORS_API_KEY", "test-fixture-key-not-real")

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

	pythonBin := resolveTestPythonBin(rootDir)

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

func TestResolvePythonBin(t *testing.T) {
	// 1. When given an explicit existing binary
	resolved := ResolvePythonBin("")
	if resolved == "" {
		t.Errorf("expected non-empty python binary, got empty")
	}

	// 2. When given "python3" or "python"
	resolved3 := ResolvePythonBin("python3")
	if resolved3 == "" {
		t.Errorf("expected resolved python for 'python3', got empty")
	}

	// 3. When given non-existent explicit path, it should fallback safely
	nonExistent := ResolvePythonBin("/non/existent/path/to/python_custom")
	if nonExistent == "" {
		t.Errorf("expected fallback when given non-existent binary, got empty")
	}
}

func TestRunSubprocessEnvOverrides(t *testing.T) {
	t.Setenv("NISKAVA_TESTING", "1")
	t.Setenv("MOCK_SECTORS", "1")
	t.Setenv("SECTORS_API_KEY", "test-fixture-key-not-real")

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

	pythonBin := resolveTestPythonBin(rootDir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_env.db")

	params := RunnerParams{
		PythonBin: pythonBin,
		WorkDir:   rootDir,
		DBPath:    dbPath,
		Ticker:    "BBCA",
		Days:      5,
		Offline:   true,
		EnvOverrides: map[string]string{
			"TEST_CUSTOM_KEY": "custom_value_42",
		},
	}

	eventsChan, errChan := RunSubprocess(ctx, params)
	for range eventsChan {
	}
	if err := <-errChan; err != nil {
		t.Fatalf("subprocess with EnvOverrides failed: %v", err)
	}
}

func resolveTestPythonBin(rootDir string) string {
	pythonBin := filepath.Join(rootDir, ".venv", "Scripts", "python.exe")
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		pythonBin = filepath.Join(rootDir, ".venv", "bin", "python3")
		if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
			if path, err := exec.LookPath("python3"); err == nil {
				pythonBin = path
			} else if path, err := exec.LookPath("python"); err == nil {
				pythonBin = path
			} else {
				pythonBin = "python"
			}
		}
	}
	return pythonBin
}

func TestResolveRepoRootAndEngine(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := ResolveRepoRoot(wd)
	if root == "" {
		t.Fatal("expected non-empty repository root")
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod not found at resolved root %s: %v", root, err)
	}

	engine := ResolveEnginePath(root, "")
	if !strings.HasSuffix(engine, filepath.Join("backend", "engine")) {
		t.Errorf("unexpected resolved engine path: %s", engine)
	}
}

func TestResolveRepoRootWithNiskavaRootEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "niskava_root_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	engineDir := filepath.Join(tempDir, "backend", "engine")
	if err := os.MkdirAll(engineDir, 0755); err != nil {
		t.Fatal(err)
	}
	runnerFile := filepath.Join(engineDir, "runner.py")
	if err := os.WriteFile(runnerFile, []byte("# runner"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NISKAVA_ROOT", tempDir)

	resolved := ResolveRepoRoot("/some/unrelated/empty/dir")
	if resolved != tempDir {
		t.Fatalf("expected resolved root to be %s from NISKAVA_ROOT, got %s", tempDir, resolved)
	}

	engineResolved := ResolveEnginePath(resolved, "")
	if engineResolved != engineDir {
		t.Fatalf("expected resolved engine path to be %s, got %s", engineDir, engineResolved)
	}
}

func TestResolvePythonBinFromUserHomeVenv(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "niskava_home_venv_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	var venvPy string
	if runtime.GOOS == "windows" {
		venvPy = filepath.Join(tempHome, ".niskava", "venv", "Scripts", "python.exe")
	} else {
		venvPy = filepath.Join(tempHome, ".niskava", "venv", "bin", "python3")
	}
	if err := os.MkdirAll(filepath.Dir(venvPy), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(venvPy, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	resolved := ResolvePythonBin("")
	if resolved != venvPy {
		t.Fatalf("expected resolved python to be %s from ~/.niskava/venv, got %s", venvPy, resolved)
	}
}

func TestResolveRepoRootUserSpaceFallback(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("NISKAVA_ROOT", "")

	// Simulate user-space cached engine in ~/.niskava/engine/runner.py
	engineDir := filepath.Join(tempHome, ".niskava", "engine")
	if err := os.MkdirAll(engineDir, 0755); err != nil {
		t.Fatal(err)
	}
	runnerFile := filepath.Join(engineDir, "runner.py")
	if err := os.WriteFile(runnerFile, []byte("# runner"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempHome); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	root := ResolveRepoRoot("")
	expected := filepath.Join(tempHome, ".niskava")
	if root != expected && !strings.Contains(root, ".niskava") {
		t.Errorf("ResolveRepoRoot did not find ~/.niskava fallback, got: %s, want: %s", root, expected)
	}
}
