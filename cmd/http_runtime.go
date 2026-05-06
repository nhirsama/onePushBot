package cmd

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	"github.com/nhirsama/onePushBot/pkg/modules"
	"github.com/spf13/viper"
)

type httpRuntime struct {
	servers []*http.Server
}

func newHTTPRuntime(hub base.Hub, moduleRuntime modules.Runtime, controller admin.Controller) (*httpRuntime, error) {
	servers, err := newHTTPServers(hub, moduleRuntime, controller)
	if err != nil {
		return nil, err
	}
	return &httpRuntime{servers: servers}, nil
}

func (r *httpRuntime) Start() error {
	if r == nil {
		return nil
	}
	listeners, err := prepareListeners(r.servers)
	if err != nil {
		return err
	}
	for i, server := range r.servers {
		listener := listeners[i]
		go func(server *http.Server, listener net.Listener) {
			if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("HTTP 服务退出: %v", err)
			}
		}(server, listener)
	}
	return nil
}

func (r *httpRuntime) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	for _, server := range r.servers {
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func newHTTPServers(hub base.Hub, moduleRuntime modules.Runtime, controller admin.Controller) ([]*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/", admin.NewHandler(viper.GetString("admin.token"), controller))

	client, ok := hub.Get(base.PlatformFeishu)
	if ok {
		if feishuClient, ok := client.(feishu.Client); ok {
			webhookPath := strings.TrimSpace(viper.GetString("feishu.webhook_path"))
			if webhookPath == "" {
				return nil, errors.New("feishu.webhook_path 不能为空")
			}
			if !strings.HasPrefix(webhookPath, "/") {
				return nil, errors.New("feishu.webhook_path 必须以 / 开头")
			}
			mux.Handle(webhookPath, feishuClient.Handler())
			warnFeishuWebhookExposure(viper.GetString("admin.addr"), webhookPath)
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

func prepareListeners(servers []*http.Server) ([]net.Listener, error) {
	listeners := make([]net.Listener, 0, len(servers))
	for _, server := range servers {
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			closeListeners(listeners)
			return nil, err
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

func closeListeners(listeners []net.Listener) {
	for _, listener := range listeners {
		_ = listener.Close()
	}
}

func warnFeishuWebhookExposure(addr string, webhookPath string) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.TrimSpace(host)
	if host == "localhost" {
		log.Printf("警告: 飞书 webhook 已复用主 HTTP 服务，但 admin.addr=%s 仅本机可访问；如无反向代理暴露，回调路径 %s 将不可达", addr, webhookPath)
		return
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		log.Printf("警告: 飞书 webhook 已复用主 HTTP 服务，但 admin.addr=%s 仅本机可访问；如无反向代理暴露，回调路径 %s 将不可达", addr, webhookPath)
	}
}
