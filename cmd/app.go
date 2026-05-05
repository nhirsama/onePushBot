package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/modules"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
	"github.com/spf13/viper"
)

type app struct {
	hub     base.Hub
	modules modules.Runtime
	router  router.Router
	servers []*http.Server
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

	servers, err := newHTTPServers(hub, moduleRuntime, controller)
	if err != nil {
		return nil, err
	}

	return &app{
		hub:     hub,
		modules: moduleRuntime,
		router:  rt,
		servers: servers,
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
	if err := a.modules.Start(ctx); err != nil {
		return err
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

// newHTTPServers 把 admin、平台 webhook 和模块 HTTP 路由统一挂到一个 HTTP 服务上。
func newHTTPServers(hub base.Hub, moduleRuntime modules.Runtime, controller admin.Controller) ([]*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/", admin.NewHandler(viper.GetString("admin.token"), controller))

	client, ok := hub.Get(base.PlatformFeishu)
	if ok {
		if feishuClient, ok := client.(feishu.Client); ok {
			mux.Handle(viper.GetString("feishu.webhook_path"), feishuClient.Handler())
		}
	}

	if err := moduleRuntime.RegisterHTTP(mux); err != nil {
		return nil, err
	}

	return []*http.Server{{
		Addr:    viper.GetString("admin.addr"),
		Handler: mux,
	}}, nil
}
