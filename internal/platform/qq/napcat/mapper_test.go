package napcat

import (
	"encoding/json"
	"testing"
	"time"

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
	if event.Message.RawText != "hello" {
		t.Fatalf("expected raw text hello, got %q", event.Message.RawText)
	}
	if event.Message.DetailType != "" {
		t.Fatalf("expected empty detail type, got %q", event.Message.DetailType)
	}
}

func TestMapEnvelopeToEvent_MessageSent(t *testing.T) {
	raw := rawEnvelope{
		Time:        1714470000,
		SelfID:      float64(10000),
		PostType:    "message_sent",
		MessageType: "private",
		SubType:     "friend",
		MessageID:   float64(1001),
		UserID:      float64(3001),
		RawMessage:  "hello",
	}

	event, ok, err := mapEnvelopeToEvent(raw, "10000")
	if err != nil {
		t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected event to be mapped")
	}
	if event.SubType != "sent_private" {
		t.Fatalf("expected sent_private subtype, got %s", event.SubType)
	}
	if !event.Message.SentBySelf {
		t.Fatal("expected sent message to be marked as self")
	}
	if event.Message.Sender.ID != "10000" {
		t.Fatalf("expected sender to be self id, got %s", event.Message.Sender.ID)
	}
}

func TestMapEnvelopeToEvent_NoticeBusinessFields(t *testing.T) {
	file, err := json.Marshal(map[string]any{
		"id":   "file-1",
		"name": "report.txt",
		"size": float64(1024),
	})
	if err != nil {
		t.Fatalf("marshal file failed: %v", err)
	}

	raw := rawEnvelope{
		Time:       1714470000,
		SelfID:     float64(10000),
		PostType:   "notice",
		NoticeType: "group_upload",
		GroupID:    float64(2001),
		UserID:     float64(3001),
		OperatorID: float64(4001),
		MessageID:  float64(5001),
		Duration:   60,
		File:       file,
	}

	event, ok, err := mapEnvelopeToEvent(raw, "")
	if err != nil {
		t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected event to be mapped")
	}
	if event.SelfID != "10000" {
		t.Fatalf("expected self id, got %s", event.SelfID)
	}
	if event.Notice.Type != "group_upload" {
		t.Fatalf("expected group_upload notice, got %s", event.Notice.Type)
	}
	if event.Notice.Operator.ID != "4001" {
		t.Fatalf("expected operator id, got %s", event.Notice.Operator.ID)
	}
	if event.Notice.MessageID != "5001" {
		t.Fatalf("expected message id, got %s", event.Notice.MessageID)
	}
	if event.Notice.FileID != "file-1" || event.Notice.FileName != "report.txt" || event.Notice.FileSize != 1024 {
		t.Fatalf("unexpected file payload: %+v", event.Notice)
	}
	if event.Notice.Duration != time.Minute {
		t.Fatalf("expected duration 1m, got %s", event.Notice.Duration)
	}
}

func TestMapEnvelopeToEvent_NotifyPokeUsesLeafSubType(t *testing.T) {
	raw := rawEnvelope{
		PostType:   "notice",
		NoticeType: "notify",
		SubType:    "poke",
		GroupID:    float64(2001),
		UserID:     float64(3001),
		TargetID:   float64(10000),
	}

	event, ok, err := mapEnvelopeToEvent(raw, "")
	if err != nil {
		t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected event to be mapped")
	}
	if event.SubType != "poke" || event.Notice.Type != "poke" {
		t.Fatalf("expected poke subtype, got event=%s notice=%s", event.SubType, event.Notice.Type)
	}
	if event.Notice.DetailType != "notify" {
		t.Fatalf("expected notify detail type, got %s", event.Notice.DetailType)
	}
}

func TestMapEnvelopeToEvent_RequestBusinessFields(t *testing.T) {
	raw := rawEnvelope{
		PostType:    "request",
		RequestType: "group",
		SubType:     "invite",
		GroupID:     float64(2001),
		UserID:      float64(3001),
		Comment:     "join please",
		Flag:        "flag-1",
	}

	event, ok, err := mapEnvelopeToEvent(raw, "")
	if err != nil {
		t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected event to be mapped")
	}
	if event.Request.Type != "group" || event.Request.DetailType != "invite" {
		t.Fatalf("unexpected request type: %+v", event.Request)
	}
	if event.Request.Comment != "join please" || event.Request.Flag != "flag-1" {
		t.Fatalf("request fields not mapped: %+v", event.Request)
	}
}

func TestMapEnvelopeToEvent_DropsProtocolEvents(t *testing.T) {
	cases := []rawEnvelope{
		{PostType: "meta_event", MetaEventType: "heartbeat"},
		{Echo: "call-1"},
		{PostType: "unknown"},
	}

	for _, raw := range cases {
		event, ok, err := mapEnvelopeToEvent(raw, "")
		if err != nil {
			t.Fatalf("mapEnvelopeToEvent returned error: %v", err)
		}
		if ok {
			t.Fatalf("protocol event should be dropped, got %+v", event)
		}
	}
}
