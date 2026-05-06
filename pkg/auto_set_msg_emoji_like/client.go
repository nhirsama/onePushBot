package autoSetMsgEmojiLike

import (
	"context"

	"github.com/nhirsama/onePushBot/pkg/plat"
)

type Client interface {
	SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error
}

func clientFromSource(clients plat.Clients) (Client, bool) {
	if clients.QQ == nil {
		return nil, false
	}
	client, ok := clients.QQ.(Client)
	return client, ok
}
