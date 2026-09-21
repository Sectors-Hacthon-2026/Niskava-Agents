package tui

import (
	"strings"
	"testing"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/ipc"
)

func TestParseSSEEventLine(t *testing.T) {
	rawSSE := "event: thought\ndata: {\"thought\": \"Menganalisis...\"}\n\nevent: message_chunk\ndata: {\"chunk\": \"Halo!\"}\n\n"
	reader := strings.NewReader(rawSSE)

	events := parseSSEReader(reader)
	if len(events) != 2 {
		t.Fatalf("expected 2 parsed events, got %d", len(events))
	}
	if events[0].Event != ipc.EventAgentThought || events[0].Thought != "Menganalisis..." {
		t.Errorf("unexpected event 0: %+v", events[0])
	}
	if events[1].Event != ipc.EventAgentMessageChunk || events[1].Chunk != "Halo!" {
		t.Errorf("unexpected event 1: %+v", events[1])
	}
}
