package autoSetMsgEmojiLike

import (
	"context"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

type Client interface {
	SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error
}

func clientFromSource(source platformclient.Source) (Client, bool) {
	return platformclient.Get[Client](source, base.PlatformQQ)
}
