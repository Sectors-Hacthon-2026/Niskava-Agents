// Package ipc handles subprocess execution and streaming JSON Lines IPC communication
// between Go Core and the Python Agent Engine.
package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// EventType defines valid IPC event names.
type EventType string

const (
	EventSessionStart     EventType = "session_start"
	EventProgressStep     EventType = "progress_step"
	EventAnomalyDetected  EventType = "anomaly_detected"
	EventFindingEmitted   EventType = "finding_emitted"
	EventSessionComplete  EventType = "session_complete"
	EventSessionError     EventType = "session_error"
	EventAgentThought     EventType = "agent_thought"
	EventAgentToolCall    EventType = "agent_tool_call"
	EventAgentObservation EventType = "agent_observation"
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
			pythonBin = "python3"
		}

		args := []string{
			"-m", "engine.runner",
			"--ticker", params.Ticker,
			"--days", fmt.Sprintf("%d", params.Days),
			"--db-path", params.DBPath,
		}
		if params.SessionID != "" {
			args = append(args, "--session", params.SessionID)
		}
		if params.Offline {
			args = append(args, "--offline")
		}

		cmd := exec.CommandContext(ctx, pythonBin, args...)
		if params.WorkDir != "" {
			cmd.Dir = params.WorkDir
		}

		// Ensure PYTHONPATH includes WorkDir or current directory
		workDir := params.WorkDir
		if workDir == "" {
			workDir = "."
		}
		cmd.Env = append(cmd.Environ(), "PYTHONPATH="+workDir)

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
			_, _ = io.Copy(&stderrBuf, stderr)
		}()

		// Read stdout JSONL line-by-line
		scanner := bufio.NewScanner(stdout)
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
