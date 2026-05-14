package run

import (
	"fmt"
	"strings"
	"time"
)

// StreamEvent wraps an SSE event with metadata.
type StreamEvent struct {
	Type      string `json:"type"`      // "log", "status", "heartbeat"
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
	Offset    int    `json:"offset"`
}

// FormatSSE formats a StreamEvent as an SSE message.
func FormatSSE(event StreamEvent) string {
	var b strings.Builder
	b.WriteString("event: ")
	b.WriteString(event.Type)
	b.WriteString("\n")
	b.WriteString("data: ")
	b.WriteString(event.Data)
	b.WriteString("\n")
	b.WriteString("id: ")
	b.WriteString(fmt.Sprintf("%d", event.Offset))
	b.WriteString("\n\n")
	return b.String()
}

// HeartbeatEvent creates a keepalive SSE event.
func HeartbeatEvent() StreamEvent {
	return StreamEvent{
		Type:      "heartbeat",
		Data:      "ping",
		Timestamp: time.Now().UnixMilli(),
	}
}
