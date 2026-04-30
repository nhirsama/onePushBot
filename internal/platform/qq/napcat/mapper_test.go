package napcat

import (
	"encoding/json"
	"testing"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

func TestMapEnvelopeToEvent_GroupMessage(t *testing.T) {
	message, err := json.Marshal([]map[string]any{
		{
			"type": "text",
			"data": map[string]any{"text": "hello"},
		},
	})
	if err != nil {
		t.Fatalf("marshal message failed: %v", err)
	}

	raw := rawEnvelope{
		Time:        1714470000,
		PostType:    "message",
		MessageType: "group",
		MessageID:   float64(1001),
		GroupID:     float64(2001),
		UserID:      float64(3001),
		GroupName:   "test-group",
		RawMessage:  "hello",
		Message:     message,
	}

	event, ok, err := mapEnvelopeToEvent(raw, "")
	if err != nil {
		t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected event to be mapped")
	}
	if event.Kind != base.EventKindMessage {
		t.Fatalf("expected message event, got %s", event.Kind)
	}
	if event.SubType != "group" {
		t.Fatalf("expected group subtype, got %s", event.SubType)
	}
	if event.Message == nil {
		t.Fatalf("expected message payload")
	}
	if event.Message.Chat.Type != base.ChatTypeGroup {
		t.Fatalf("expected group chat, got %s", event.Message.Chat.Type)
	}
	if event.Message.Text != "hello" {
		t.Fatalf("expected text hello, got %q", event.Message.Text)
	}
}
