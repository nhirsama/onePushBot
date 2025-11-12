package domain

import (
	"context"
	"encoding/json"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeGroup       MessageType = "group"
	MessageTypePrivate     MessageType = "private"
	MessageTypeNotice      MessageType = "notice"
	MessageTypePoke        MessageType = "poke"
	MessageTypeHeartbeat   MessageType = "heartbeat"
	MessageTypeLifecycle   MessageType = "lifecycle"
	MessageTypeAPIResponse MessageType = "api_response"
)

type DispatcherStats struct {
	TotalMessages   int64
	HandledMessages int64
	ErrorMessages   int64
	TotalLatency    int64
}
type RawMessage []byte
type MessageHandler interface {
	Handle(ctx context.Context, messageStruct *MessageStruct) error
}
type MessageDispatcher interface {
	// Dispatch 分发消息到不同的处理器
	Dispatch(ctx context.Context, rawMsg RawMessage) error

	// Handler 注册处理器
	HandlerPostTypeMessage(ctx context.Context, msg *MessageStruct) error
	HandlerPostTypeMetaEvent(ctx context.Context, msg *MessageStruct) error
	HandlerPostTypeNotice(ctx context.Context, msg *MessageStruct) error
	HandlerPostTypeMessageSent(ctx context.Context, msg *MessageStruct) error
	HandlerPostTypeRequest(ctx context.Context, msg *MessageStruct) error
	HandlerPostTypeOther(ctx context.Context, msg *MessageStruct) error

	// GetStats 获取统计信息
	GetStats() DispatcherStats
}

type PostType string

// PostType
const (
	PostTypeMessage     PostType = "message"
	PostTypeMetaEvent   PostType = "meta_event"
	PostTypeNotice      PostType = "notice"
	PostTypeMessageSent PostType = "message_sent"
	PostTypeRequest     PostType = "request"
	PostTypeOther       PostType = "other"
)

type MessageStruct struct {
	Time          int64           `json:"time"`
	SelfId        int64           `json:"self_id"`
	PostType      PostType        `json:"post_type"`
	MetaEventType string          `json:"meta_event_type"`
	Interval      int64           `json:"interval"`
	MessageId     int64           `json:"message_id"`
	MessageSeq    json.RawMessage `json:"message_seq"`
	RealId        json.RawMessage `json:"real_id"`
	RealSeq       json.RawMessage `json:"real_seq"`
	MessageType   string          `json:"message_type"`
	Sender        json.RawMessage `json:"sender"`
	RawMessage    string          `json:"raw_message"`
	Font          json.RawMessage `json:"font"`
	SubType       string          `json:"sub_type"`
	Message       json.RawMessage `json:"message"`
	MessageFormat string          `json:"message_format"`
	GroupId       int64           `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserId        int64           `json:"user_id"`
	Status        json.RawMessage `json:"status"`
	RetCode       int64           `json:"retcode"`
	TargetId      int64           `json:"target_id"`
	Wording       string          `json:"wording"`
	Echo          string          `json:"echo"`
}
