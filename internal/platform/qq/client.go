package qq

import (
	"context"
	"encoding/json"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

// Client 定义 QQ 平台的原生能力。
type Client interface {
	base.Client

	// Call 允许上层直接访问 NapCat/OneBot action。
	Call(ctx context.Context, action string, params any) (json.RawMessage, error)
	SendGroupText(ctx context.Context, groupID string, text string) error
	SendPrivateText(ctx context.Context, userID string, text string) error
	SendPoke(ctx context.Context, groupID string, userID string) error
	SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error
}
