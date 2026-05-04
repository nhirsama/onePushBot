package modules

import (
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/auto_set_msg_emoji_like"
)

func init() {
	register(Module{
		Name: "auto_set_msg_emoji_like",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			return autoSetMsgEmojiLike.Register(rt, deps.Clients)
		},
	})
}
