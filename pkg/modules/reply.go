package modules

import (
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/reply"
)

func init() {
	register(Module{
		Name: "reply",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			return reply.Register(rt, deps.Clients)
		},
	})
}
