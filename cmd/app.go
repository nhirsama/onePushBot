package cmd

import (
	"context"
	"time"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/modules"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
	"github.com/spf13/viper"
)

type app struct {
	hub     base.Hub
	http    *httpRuntime
	modules modules.Runtime
	router  router.Router
}

// newApp 在 cmd 层完成依赖注入，平台 Hub 和 Router 都不依赖旧内核。
func newApp(controller admin.Controller) (*app, error) {
	logger := newLogger()

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
	clients := platformclient.NewSource(hub)
	moduleRuntime := modules.NewRuntime(clients)

	if err := moduleRuntime.RegisterRoutes(rt); err != nil {
		return nil, err
	}

	httpRuntime, err := newHTTPRuntime(hub, moduleRuntime, controller)
	if err != nil {
		return nil, err
	}

	return &app{
		hub:     hub,
		http:    httpRuntime,
		modules: moduleRuntime,
		router:  rt,
	}, nil
}

// Start 先启动路由再启动平台，避免平台事件早于路由订阅。
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
	if err := a.http.Start(); err != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.hub.Close(closeCtx)
		_ = a.router.Close(closeCtx)
		return err
	}
	if err := a.modules.Start(ctx); err != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.Close(closeCtx)
		return err
	}
	a.notifyStarted(ctx)
	return nil
}

// Close 先关闭外部入口，再关闭平台 Hub，最后关闭 Router。
func (a *app) Close(ctx context.Context) error {
	if err := a.http.Close(ctx); err != nil {
		return err
	}
	if err := a.hub.Close(ctx); err != nil {
		return err
	}
	return a.router.Close(ctx)
}
