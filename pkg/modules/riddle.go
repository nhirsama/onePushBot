package modules

import (
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/riddle"
)

func init() {
	register(Module{
		Name: "riddle",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			return riddle.Register(rt, deps.Clients)
		},
		RegisterHTTP: func(mux HTTPRegistrar, deps Dependencies) error {
			_ = deps
			return riddle.RegisterHTTP(mux)
		},
	})
}
