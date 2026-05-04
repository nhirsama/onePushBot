package modules

import (
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/sudo"
)

func init() {
	register(Module{
		Name: "sudo",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			return sudo.Register(rt, deps.Clients)
		},
	})
}
