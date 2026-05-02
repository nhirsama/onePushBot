package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nhirsama/onePushBot/config"
	base "github.com/nhirsama/onePushBot/internal/platform"
	qqnapcat "github.com/nhirsama/onePushBot/internal/platform/qq/napcat"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
	"github.com/nhirsama/onePushBot/pkg/logger"
	"github.com/spf13/viper"
)

func Cli() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub, err := newPlatformHub()
	if err != nil {
		log.Fatalf("初始化 platform hub 失败: %v", err)
	}

	if err := hub.Start(ctx); err != nil {
		log.Fatalf("启动 platform hub 失败: %v", err)
	}

	<-ctx.Done()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := hub.Close(closeCtx); err != nil {
		log.Printf("关闭 platform hub 失败: %v", err)
	}

	config.SaveConfig()
	log.Println("程序正常结束，正在保存与释放资源")
}

func newPlatformHub() (base.Hub, error) {
	hub := base.NewHub(nil)

	client, err := newPlatformClient()
	if err != nil {
		return nil, err
	}
	if err := hub.Register(client); err != nil {
		return nil, err
	}

	return hub, nil
}

func newPlatformClient() (base.Client, error) {
	switch selectedPlatform() {
	case "", "qq", "napcat", "qq_napcat":
		return newQQNapcatClient()
	default:
		return nil, fmt.Errorf("暂不支持的平台: %s", selectedPlatform())
	}
}

func selectedPlatform() string {
	return strings.TrimSpace(strings.ToLower(viper.GetString("platform")))
}

func newQQNapcatClient() (base.Client, error) {
	selfID := strings.TrimSpace(viper.GetString("bot_config.self_id"))
	if selfID == "" && config.SelfId != 0 {
		selfID = strconv.FormatInt(config.SelfId, 10)
	}

	cfg := qqnapcat.Config{
		APIURL:               firstNonEmpty(viper.GetString("bot_config.api_url"), viper.GetString("apiUrl")),
		Token:                firstNonEmpty(viper.GetString("bot_config.token"), viper.GetString("token")),
		SelfID:               selfID,
		HeartbeatTimeout:     secondsToDuration(viper.GetInt64("bot_config.heartbeat_timeout")),
		ReconnectMaxAttempts: int(viper.GetInt64("bot_config.reconnect_max_attempts")),
	}

	return qqnapcat.New(cfg, qqnapcat.Dependencies{
		Logger: logger.NewLog(pkgDomain.InfoLevel),
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func secondsToDuration(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
