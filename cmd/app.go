package cmd

import (
	"context"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
)

type app struct {
	hub    base.Hub
	router router.Router
}

func newApp() (*app, error) {
	hub, err := newPlatformHub()
	if err != nil {
		return nil, err
	}

	rt, err := router.New(hub.Bus(), router.Options{
		BrokerBuffer: 128,
	})
	if err != nil {
		return nil, err
	}

	if err := registerRoutes(rt, hub); err != nil {
		return nil, err
	}

	return &app{
		hub:    hub,
		router: rt,
	}, nil
}

func (a *app) Start(ctx context.Context) error {
	if err := a.router.Start(ctx); err != nil {
		return err
	}
	if err := a.hub.Start(ctx); err != nil {
		closeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = a.router.Close(closeCtx)
		return err
	}
	return nil
}

func (a *app) Close(ctx context.Context) error {
	if err := a.hub.Close(ctx); err != nil {
		return err
	}
	return a.router.Close(ctx)
}

func registerRoutes(rt router.Router, hub base.Hub) error {
	_ = hub
	_ = rt
	return nil
}
