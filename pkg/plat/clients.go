package plat

import (
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	"github.com/nhirsama/onePushBot/internal/platform/qq"
	tgbot "github.com/nhirsama/onePushBot/internal/platform/telegram/bot"
	tguser "github.com/nhirsama/onePushBot/internal/platform/telegram/user"
)

type Source interface {
	Get(platform base.Platform) (base.Client, bool)
}

type TelegramClients struct {
	Bot  tgbot.Client
	User tguser.Client
}

type Clients struct {
	QQ       qq.Client
	Feishu   feishu.Client
	Telegram *TelegramClients
}

func New(source Source) Clients {
	var clients Clients
	if source == nil {
		return clients
	}

	if client, ok := source.Get(base.PlatformQQ); ok {
		clients.QQ, _ = client.(qq.Client)
	}
	if client, ok := source.Get(base.PlatformFeishu); ok {
		clients.Feishu, _ = client.(feishu.Client)
	}

	var telegram TelegramClients
	if client, ok := source.Get(base.PlatformTelegramBot); ok {
		telegram.Bot, _ = client.(tgbot.Client)
	}
	if client, ok := source.Get(base.PlatformTelegramUser); ok {
		telegram.User, _ = client.(tguser.Client)
	}
	if telegram.Bot != nil || telegram.User != nil {
		clients.Telegram = &telegram
	}

	return clients
}
