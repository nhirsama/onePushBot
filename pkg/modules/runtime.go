package modules

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

type Dependencies struct {
	Clients platformclient.Source
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

var registeredModules []Module

func register(module Module) {
	registeredModules = append(registeredModules, module)
}

func modules() []Module {
	result := make([]Module, len(registeredModules))
	copy(result, registeredModules)
	return result
}

func NewRuntime(clients platformclient.Source) Runtime {
	return Runtime{
		deps:    Dependencies{Clients: clients},
		modules: modules(),
	}
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
