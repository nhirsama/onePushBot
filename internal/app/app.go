package app

import (
	"context"
	"time"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/spf13/viper"
)

type App struct {
	logger base.Logger
	hub    base.Hub
	http   *httpRuntime
	rt     router.Router
}

func New(controller admin.Controller, logger base.Logger) (*App, error) {
	if logger == nil {
		logger = base.NewDiscardLogger()
	}

	hub, err := newPlatformHub(logger)
	if err != nil {
		return nil, err
	}

	rt, err := router.New(hub.Bus(), router.Options{
		BrokerBuffer: viper.GetInt("router.broker_buffer"),
		Logger:       logger,
	})
	if err != nil {
		return nil, err
	}

	httpRuntime, err := newHTTPRuntime(hub, controller, logger)
	if err != nil {
		return nil, err
	}

	return &App{
		logger: logger,
		hub:    hub,
		http:   httpRuntime,
		rt:     rt,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	if err := a.rt.Start(ctx); err != nil {
		return err
	}
	if err := a.hub.Start(ctx); err != nil {
		closeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = a.rt.Close(closeCtx)
		return err
	}
	if err := a.http.Start(); err != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.hub.Close(closeCtx)
		_ = a.rt.Close(closeCtx)
		return err
	}
	a.logger.Info("程序已启动",
		"platform_count", len(a.hub.All()),
	)
	a.notifyStarted(ctx)
	return nil
}

func (a *App) Close(ctx context.Context) error {
	if err := a.http.Close(ctx); err != nil {
		return err
	}
	if err := a.hub.Close(ctx); err != nil {
		return err
	}
	return a.rt.Close(ctx)
}

func (a *App) PlatformStatuses() map[string]string {
	statuses := make(map[string]string)
	if a == nil || a.hub == nil {
		return statuses
	}
	for _, client := range a.hub.All() {
		statuses[string(client.Platform())] = string(client.Status())
	}
	return statuses
}
