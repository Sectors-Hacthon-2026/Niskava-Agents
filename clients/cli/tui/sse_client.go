package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
)

// IsDaemonAlive checks if the local server daemon is responsive.
func IsDaemonAlive(serverURL string) bool {
	if serverURL == "" {
		return false
	}
	client := http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("%s/api/health", serverURL))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// StreamChatViaSSE sends a prompt to the background daemon via HTTP SSE.
func StreamChatViaSSE(ctx context.Context, serverURL, sessionID, prompt string) (<-chan ipc.Event, <-chan error) {
	eventsChan := make(chan ipc.Event, 64)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventsChan)
		defer close(errChan)

		body, _ := json.Marshal(map[string]string{
			"session_id": sessionID,
			"prompt":     prompt,
		})

		req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", serverURL), bytes.NewReader(body))
		if err != nil {
			errChan <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 0}
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("server returned HTTP %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		var currentEvent string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "event:") {
				currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			} else if strings.HasPrefix(line, "data:") {
				dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				var ev ipc.Event
				if err := json.Unmarshal([]byte(dataStr), &ev); err == nil {
					if ev.Event == "" && currentEvent != "" {
						ev.Event = ipc.EventType(currentEvent)
					}
					// Normalize event names
					switch currentEvent {
					case "thought":
						ev.Event = ipc.EventAgentThought
					case "tool_call":
						ev.Event = ipc.EventAgentToolCall
					case "observation":
						ev.Event = ipc.EventAgentObservation
					case "message_chunk":
						ev.Event = ipc.EventAgentMessageChunk
					case "message_complete":
						ev.Event = ipc.EventAgentMessageComplete
					case "finding_emitted":
						ev.Event = ipc.EventFindingEmitted
					case "anomaly_detected":
						ev.Event = ipc.EventAnomalyDetected
					case "session_complete", "done":
						ev.Event = ipc.EventSessionComplete
					case "session_error", "error":
						ev.Event = ipc.EventSessionError
					}
					eventsChan <- ev
				}
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			errChan <- formatSSEStreamError(err)
		}
	}()

	return eventsChan, errChan
}

// formatSSEStreamError translates lower-level transport errors (such as unexpected EOF) into clear English messages.
func formatSSEStreamError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "unexpected EOF") {
		return fmt.Errorf("stream disconnected unexpectedly (server closed connection prematurely)")
	}
	return fmt.Errorf("error reading SSE stream: %w", err)
}

func parseSSEReader(r io.Reader) []ipc.Event {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	var events []ipc.Event
	var currentEvent string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var ev ipc.Event
			if err := json.Unmarshal([]byte(dataStr), &ev); err == nil {
				if ev.Event == "" && currentEvent != "" {
					ev.Event = ipc.EventType(currentEvent)
				}
				switch currentEvent {
				case "thought":
					ev.Event = ipc.EventAgentThought
				case "tool_call":
					ev.Event = ipc.EventAgentToolCall
				case "observation":
					ev.Event = ipc.EventAgentObservation
				case "message_chunk":
					ev.Event = ipc.EventAgentMessageChunk
				case "message_complete":
					ev.Event = ipc.EventAgentMessageComplete
				case "finding_emitted":
					ev.Event = ipc.EventFindingEmitted
				case "anomaly_detected":
					ev.Event = ipc.EventAnomalyDetected
				}
				events = append(events, ev)
			}
		}
	}
	return events
}
