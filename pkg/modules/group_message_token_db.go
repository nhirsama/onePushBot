package modules

import (
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/group_message_token_db"
)

func init() {
	register(Module{
		Name: "group_message_token_db",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			_ = deps
			return groupmessagetokendb.Register(rt)
		},
	})
}
