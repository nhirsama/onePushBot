package autoSetMsgEmojiLike

import (
	"context"
	"log"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/plat"
)

func Register(rt router.Router, clients plat.Clients) error {
	return rt.Register(router.Route{
		Name: "auto_set_msg_emoji_like",
		Filter: base.EventFilter{
			Platform: base.PlatformQQ,
			Kind:     base.EventKindMessage,
			ChatType: base.ChatTypeGroup,
		},
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			if event.Message == nil {
				return nil
			}
			client, ok := clientFromSource(clients)
			if !ok {
				return nil
			}
			for _, emojiID := range emojiLikesForUser(event.Message.Sender.ID) {
				if err := client.SetMsgEmojiLike(ctx, event.Message.ID, emojiID, true); err != nil {
					log.Printf("设置消息表情回应失败: %v", err)
				}
			}
			return nil
		}),
	})
}
