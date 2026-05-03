package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/spf13/viper"
)

type app struct {
	hub     base.Hub
	router  router.Router
	servers []*http.Server
}

// newApp 在 cmd 层完成依赖注入，平台 Hub 和 Router 都不依赖旧内核。
func newApp() (*app, error) {
	hub, err := newPlatformHub()
	if err != nil {
		return nil, err
	}

	rt, err := router.New(hub.Bus(), router.Options{
		BrokerBuffer: viper.GetInt("router.broker_buffer"),
	})
	if err != nil {
		return nil, err
	}

	if err := registerRoutes(rt, hub); err != nil {
		return nil, err
	}

	return &app{
		hub:     hub,
		router:  rt,
		servers: newHTTPServers(hub),
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
	for _, server := range a.servers {
		go func(server *http.Server) {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("HTTP 服务退出: %v", err)
			}
		}(server)
	}
	a.notifyStarted(ctx)
	return nil
}

// Close 先关闭外部入口，再关闭平台 Hub，最后关闭 Router。
func (a *app) Close(ctx context.Context) error {
	for _, server := range a.servers {
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
	}
	if err := a.hub.Close(ctx); err != nil {
		return err
	}
	return a.router.Close(ctx)
}

// newHTTPServers 为 webhook 型平台补充 HTTP 入口。
func newHTTPServers(hub base.Hub) []*http.Server {
	client, ok := hub.Get(base.PlatformFeishu)
	if !ok {
		return nil
	}

	feishuClient, ok := client.(feishu.Client)
	if !ok {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(viper.GetString("feishu.webhook_path"), feishuClient.Handler())
	return []*http.Server{
		{
			Addr:    viper.GetString("feishu.http_addr"),
			Handler: mux,
		},
	}
}

// registerRoutes 是后续业务路由注入点；无平台配置时保持空路由启动。
func registerRoutes(rt router.Router, hub base.Hub) error {
	_ = hub
	_ = rt
	return nil
}
