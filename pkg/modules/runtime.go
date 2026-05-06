package modules

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/auto_set_msg_emoji_like"
	groupmessagetokendb "github.com/nhirsama/onePushBot/pkg/group_message_token_db"
	nowcoderTracker "github.com/nhirsama/onePushBot/pkg/nowcoder_tracker"
	"github.com/nhirsama/onePushBot/pkg/plat"
	"github.com/nhirsama/onePushBot/pkg/reply"
	"github.com/nhirsama/onePushBot/pkg/riddle"
	"github.com/nhirsama/onePushBot/pkg/sudo"
)

type Dependencies struct {
	Plat plat.Clients
}

type HTTPRegistrar interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type Module struct {
	Name           string
	RegisterRoutes func(router.Router, Dependencies) error
	Start          func(context.Context, Dependencies) error
	RegisterHTTP   func(HTTPRegistrar, Dependencies) error
}

type Runtime struct {
	deps    Dependencies
	modules []Module
}

func NewRuntime(deps Dependencies, enabled []string) (Runtime, error) {
	items, err := selectModules(enabled)
	if err != nil {
		return Runtime{}, err
	}
	return Runtime{
		deps:    deps,
		modules: items,
	}, nil
}

func (r Runtime) RegisterRoutes(rt router.Router) error {
	for _, module := range r.modules {
		if module.RegisterRoutes == nil {
			continue
		}
		if err := module.RegisterRoutes(rt, r.deps); err != nil {
			return fmt.Errorf("register module %s routes: %w", module.Name, err)
		}
	}
	return nil
}

func (r Runtime) Start(ctx context.Context) error {
	for _, module := range r.modules {
		if module.Start == nil {
			continue
		}
		if err := module.Start(ctx, r.deps); err != nil {
			return fmt.Errorf("start module %s: %w", module.Name, err)
		}
	}
	return nil
}

func (r Runtime) RegisterHTTP(mux HTTPRegistrar) error {
	for _, module := range r.modules {
		if module.RegisterHTTP == nil {
			continue
		}
		if err := module.RegisterHTTP(mux, r.deps); err != nil {
			return fmt.Errorf("register module %s http: %w", module.Name, err)
		}
	}
	return nil
}

func builtinModules() map[string]Module {
	return map[string]Module{
		"auto_set_msg_emoji_like": {
			Name: "auto_set_msg_emoji_like",
			RegisterRoutes: func(rt router.Router, deps Dependencies) error {
				return autoSetMsgEmojiLike.Register(rt, deps.Plat)
			},
		},
		"group_message_token_db": {
			Name: "group_message_token_db",
			RegisterRoutes: func(rt router.Router, deps Dependencies) error {
				return groupmessagetokendb.Register(rt)
			},
		},
		"nowcoder_daily": {
			Name: "nowcoder_daily",
			Start: func(ctx context.Context, deps Dependencies) error {
				nowcoderTracker.StartDaily(ctx, deps.Plat)
				return nil
			},
		},
		"reply": {
			Name: "reply",
			RegisterRoutes: func(rt router.Router, deps Dependencies) error {
				return reply.Register(rt, deps.Plat)
			},
		},
		"riddle": {
			Name: "riddle",
			RegisterRoutes: func(rt router.Router, deps Dependencies) error {
				return riddle.Register(rt, deps.Plat)
			},
			RegisterHTTP: func(mux HTTPRegistrar, deps Dependencies) error {
				_ = deps
				return riddle.RegisterHTTP(mux)
			},
		},
		"sudo": {
			Name: "sudo",
			RegisterRoutes: func(rt router.Router, deps Dependencies) error {
				return sudo.Register(rt, deps.Plat)
			},
		},
	}
}

func selectModules(enabled []string) ([]Module, error) {
	catalog := builtinModules()
	seen := make(map[string]struct{}, len(enabled))
	result := make([]Module, 0, len(enabled))
	for _, item := range enabled {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		module, ok := catalog[item]
		if !ok {
			return nil, fmt.Errorf("unknown module: %s", item)
		}
		seen[item] = struct{}{}
		result = append(result, module)
	}
	return result, nil
}
