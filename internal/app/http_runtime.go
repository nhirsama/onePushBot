package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	"github.com/spf13/viper"
)

type httpRuntime struct {
	logger  base.Logger
	servers []*http.Server
}

func newHTTPRuntime(hub base.Hub, controller admin.Controller, logger base.Logger) (*httpRuntime, error) {
	if logger == nil {
		logger = base.NewDiscardLogger()
	}
	servers, err := newHTTPServers(hub, controller)
	if err != nil {
		return nil, err
	}
	return &httpRuntime{logger: logger, servers: servers}, nil
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
	if len(r.servers) > 0 {
		r.logger.Info("管理面板已启动", "url", adminPanelURL(r.servers[0].Addr, viper.GetString("admin.token")))
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

func newHTTPServers(hub base.Hub, controller admin.Controller) ([]*http.Server, error) {
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

func adminPanelURL(addr, token string) string {
	if strings.TrimSpace(addr) == "" {
		return ""
	}
	u := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/",
	}
	if strings.TrimSpace(token) != "" {
		u.RawQuery = url.Values{"token": []string{token}}.Encode()
	}
	return u.String()
}
