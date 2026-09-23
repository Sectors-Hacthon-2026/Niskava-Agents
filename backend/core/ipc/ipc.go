// Package ipc handles subprocess execution and streaming JSON Lines IPC communication
// between Go Core and the Python Agent Engine.
package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ipcScannerBufferBytes = 1024 * 1024 // 1MB
	ipcStderrMaxBytes     = 32 * 1024   // 32KB
)

// EventType defines valid IPC event names.
type EventType string

const (
	EventSessionStart         EventType = "session_start"
	EventProgressStep         EventType = "progress_step"
	EventAnomalyDetected      EventType = "anomaly_detected"
	EventFindingEmitted       EventType = "finding_emitted"
	EventSessionComplete      EventType = "session_complete"
	EventSessionError         EventType = "session_error"
	EventAgentThought         EventType = "agent_thought"
	EventAgentToolCall        EventType = "agent_tool_call"
	EventAgentObservation     EventType = "agent_observation"
	EventAgentMessageChunk    EventType = "agent_message_chunk"
	EventAgentMessageComplete EventType = "agent_message_complete"
)

// Event represents a generic JSON Lines IPC payload.
type Event struct {
	Event            EventType              `json:"event"`
	SessionID        string                 `json:"session_id"`
	Ticker           string                 `json:"ticker,omitempty"`
	TimeframeDays    int                    `json:"timeframe_days,omitempty"`
	Timestamp        string                 `json:"timestamp,omitempty"`
	Thought          string                 `json:"thought,omitempty"`
	Tool             string                 `json:"tool,omitempty"`
	Args             map[string]interface{} `json:"args,omitempty"`
	Content          string                 `json:"content,omitempty"`
	Chunk            string                 `json:"chunk,omitempty"`
	Stage            string                 `json:"stage,omitempty"`
	StepIndex        int                    `json:"step_index,omitempty"`
	TotalSteps       int                    `json:"total_steps,omitempty"`
	Message          string                 `json:"message,omitempty"`
	AnomalyDate      string                 `json:"anomaly_date,omitempty"`
	MetricType       string                 `json:"metric_type,omitempty"`
	ZScore           float64                `json:"z_score,omitempty"`
	MetricValue      float64                `json:"metric_value,omitempty"`
	BaselineValue    float64                `json:"baseline_value,omitempty"`
	PriceChangePct   float64                `json:"price_change_pct,omitempty"`
	SectorChangePct  float64                `json:"sector_change_pct,omitempty"`
	Description      string                 `json:"description,omitempty"`
	ID               string                 `json:"id,omitempty"`
	Title            string                 `json:"title,omitempty"`
	ClaimText        string                 `json:"claim_text,omitempty"`
	VerificationStat string                 `json:"verification_status,omitempty"`
	ConfidenceScore  float64                `json:"confidence_score,omitempty"`
	CausalityStatus  string                 `json:"causality_status,omitempty"`
	Evidence         json.RawMessage        `json:"evidence,omitempty"`
	TotalAnomalies   int                    `json:"total_anomalies,omitempty"`
	TotalFindings    int                    `json:"total_findings,omitempty"`
	DurationMs       int                    `json:"duration_ms,omitempty"`
	Summary          string                 `json:"summary,omitempty"`
	Error            string                 `json:"error,omitempty"`
}

// RunnerParams defines parameters to invoke the Python engine.
type RunnerParams struct {
	PythonBin  string
	EnginePath string
	WorkDir    string
	DBPath     string
	Ticker     string
	Days       int
	SessionID  string
	Offline    bool
	Prompt     string
	Language   string
}

// RunConversationStream spawns the Python runner for interactive or batch conversation turns.
func RunConversationStream(ctx context.Context, params RunnerParams) (<-chan Event, <-chan error) {
	return RunSubprocess(ctx, params)
}

// RunSubprocess spawns the Python runner and returns a channel of streaming events.
func RunSubprocess(ctx context.Context, params RunnerParams) (<-chan Event, <-chan error) {
	eventsChan := make(chan Event, 64)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventsChan)
		defer close(errChan)

		pythonBin := params.PythonBin
		if pythonBin == "" {
			if path, err := exec.LookPath("python3"); err == nil {
				pythonBin = path
			} else if path, err := exec.LookPath("python"); err == nil {
				pythonBin = path
			} else {
				pythonBin = "python"
			}
		}

		args := []string{
			"-m", "engine.runner",
			"--db-path", params.DBPath,
		}
		if params.Prompt != "" {
			args = append(args, "--prompt", params.Prompt)
		}
		if params.Ticker != "" {
			args = append(args, "--ticker", params.Ticker)
		}
		if params.Days > 0 {
			args = append(args, "--days", fmt.Sprintf("%d", params.Days))
		}
		if params.SessionID != "" {
			args = append(args, "--session", params.SessionID)
		}
		if params.Offline {
			args = append(args, "--offline")
		}
		if params.Language != "" {
			args = append(args, "--language", params.Language)
		}

		cmd := exec.CommandContext(ctx, pythonBin, args...)
		if params.WorkDir != "" {
			cmd.Dir = params.WorkDir
		}

		// Ensure PYTHONPATH includes backend directory, WorkDir, and environment PYTHONPATH
		workDir := params.WorkDir
		if workDir == "" {
			workDir = "."
		}
		backendDir := filepath.Join(workDir, "backend")
		pythonPath := backendDir + string(filepath.ListSeparator) + workDir
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = pythonPath + string(filepath.ListSeparator) + existing
		}
		cmd.Env = append(cmd.Environ(), "PYTHONPATH="+pythonPath)
		if params.Language != "" {
			cmd.Env = append(cmd.Env, "NISKAVA_LANG="+params.Language)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errChan <- fmt.Errorf("failed to create stdout pipe: %w", err)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			errChan <- fmt.Errorf("failed to create stderr pipe: %w", err)
			return
		}

		if err := cmd.Start(); err != nil {
			errChan <- fmt.Errorf("failed to start engine subprocess: %w", err)
			return
		}

		// Read stderr in background for error reporting
		var stderrBuf strings.Builder
		go func() {
			limitedStderr := io.LimitReader(stderr, ipcStderrMaxBytes)
			_, _ = io.Copy(&stderrBuf, limitedStderr)
		}()

		// Read stdout JSONL line-by-line
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, ipcScannerBufferBytes), ipcScannerBufferBytes)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var ev Event
			if err := json.Unmarshal(line, &ev); err != nil {
				// Ignore non-JSON log lines
				continue
			}
			eventsChan <- ev
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading stdout scanner: %w", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			errStr := strings.TrimSpace(stderrBuf.String())
			if errStr != "" {
				errChan <- fmt.Errorf("engine subprocess error: %s (%w)", errStr, err)
			} else {
				errChan <- fmt.Errorf("engine subprocess exited with error: %w", err)
			}
			return
		}
	}()

	return eventsChan, errChan
}
