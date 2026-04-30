package telegram

import (
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

type (
	ChatID    string
	MessageID string
	UserID    string
)

// Dialog 描述一个 Telegram 会话。
type Dialog struct {
	ID    ChatID
	Type  base.ChatType
	Title string
	Raw   any
}

// Message 描述 Telegram 平台层抽出的消息。
type Message struct {
	ID     MessageID
	ChatID ChatID
	Chat   base.Chat
	Sender base.User
	Text   string
	Time   time.Time
	Raw    any
}

// Command 表示 Telegram Bot 命令定义。
type Command struct {
	Command     string
	Description string
}
