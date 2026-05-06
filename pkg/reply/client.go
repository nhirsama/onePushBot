package reply

import (
	"context"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/plat"
)

type Client interface {
	SendGroupText(ctx context.Context, groupID string, text string) error
	SendPoke(ctx context.Context, groupID string, userID string) error
}

func clientFromEvent(clients plat.Clients, event base.Event) (Client, bool) {
	switch event.Platform {
	case base.PlatformQQ:
		if clients.QQ == nil {
			return nil, false
		}
		client, ok := clients.QQ.(Client)
		return client, ok
	case base.PlatformFeishu:
		if clients.Feishu == nil {
			return nil, false
		}
		client, ok := clients.Feishu.(Client)
		return client, ok
	case base.PlatformTelegramBot:
		if clients.Telegram == nil || clients.Telegram.Bot == nil {
			return nil, false
		}
		client, ok := clients.Telegram.Bot.(Client)
		return client, ok
	case base.PlatformTelegramUser:
		if clients.Telegram == nil || clients.Telegram.User == nil {
			return nil, false
		}
		client, ok := clients.Telegram.User.(Client)
		return client, ok
	default:
		return nil, false
	}
}
