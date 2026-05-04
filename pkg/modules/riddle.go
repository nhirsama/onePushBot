package modules

import (
	"net/http"

	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/riddle"
)

func init() {
	register(Module{
		Name: "riddle",
		RegisterRoutes: func(rt router.Router, deps Dependencies) error {
			return riddle.Register(rt, deps.Clients)
		},
		HTTPServers: func(deps Dependencies) []*http.Server {
			_ = deps
			if server := riddle.NewHTTPServer(); server != nil {
				return []*http.Server{server}
			}
			return nil
		},
	})
}
