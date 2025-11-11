package domain

import (
	"context"
)

type Bot interface {
	// SendGroupMsg 发送群消息
	SendGroupMsg(ctx context.Context, groupID int64, text string, msgType string) error
	// SendPrivateMsg 发送私聊消息
	SendPrivateMsg(ctx context.Context, userID int64, text string, msgType string) error
	// SendPoke 发送戳一戳
	SendPoke(ctx context.Context, groupID, targetID int64) error
	// SetMsgEmojiLike 设置消息表情回应
	SetMsgEmojiLike(ctx context.Context, messageID int64, emojiID int, set bool) error
	// SendLike 发送点赞
	SendLike(ctx context.Context, userID, times int) error
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleGroupMessage(ctx context.Context, msg *Message) error
	HandlePrivateMessage(ctx context.Context, msg *Message) error
	HandlePoke(ctx context.Context, msg *Message) error
	HandleNotice(ctx context.Context, msg *Message) error
}

// LLMClient LLM 客户端接口
type LLMClient interface {
	Chat(ctx context.Context, messages []ChatMessage) (string, error)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	UpdateUserEntropy(ctx context.Context, userID int64, text string) error
	GetUserEntropy(ctx context.Context, userID int64) (float64, error)
	Save(ctx context.Context) error
}

// AuthService 认证服务接口
type AuthService interface {
	Authenticate(message []byte) (string, bool)
	AddPublicKey(pubKey string)
}

// EventBus 事件总线接口
type EventBus interface {
	Subscribe(topic string, fn interface{}) error
	Publish(topic string, args ...interface{})
}

// Message 统一消息结构
type Message struct {
	MessageID   int64
	MessageType string // group, private
	SubType     string
	GroupID     int64
	UserID      int64
	TargetID    int64
	RawMessage  string
	Sender      map[string]interface{}
	Time        int64
}

// ChatMessage LLM 消息
type ChatMessage struct {
	Role    string
	Content string
}
