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
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	qqnapcat "github.com/nhirsama/onePushBot/internal/platform/qq/napcat"
	tgbot "github.com/nhirsama/onePushBot/internal/platform/telegram/bot"
	tguser "github.com/nhirsama/onePushBot/internal/platform/telegram/user"
	"github.com/spf13/viper"
)

func Cli() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := newApp()
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}

	if err := app.Start(ctx); err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}

	<-ctx.Done()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Close(closeCtx); err != nil {
		log.Printf("关闭应用失败: %v", err)
	}

	config.SaveConfig()
	log.Println("程序正常结束，正在保存与释放资源")
}

// newPlatformHub 是 cmd 的平台组合入口，只负责把配置中的平台客户端注册进 Hub。
func newPlatformHub() (base.Hub, error) {
	hub := base.NewHub(nil)

	clients, err := newPlatformClients()
	if err != nil {
		return nil, err
	}
	for _, client := range clients {
		if err := hub.Register(client); err != nil {
			return nil, err
		}
	}

	return hub, nil
}

// newPlatformClients 根据配置创建多个平台客户端，各平台能力保持独立。
func newPlatformClients() ([]base.Client, error) {
	platforms := selectedPlatforms()
	clients := make([]base.Client, 0, len(platforms))

	for _, platform := range platforms {
		client, err := newPlatformClient(platform)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

// newPlatformClient 只做平台名称到具体实现的选择，不在这里抽象平台能力。
func newPlatformClient(name string) (base.Client, error) {
	switch name {
	case "", "qq", "napcat", "qq_napcat":
		return newQQNapcatClient()
	case "telegram_bot", "tg_bot":
		return newTelegramBotClient()
	case "telegram_user", "tg_user":
		return newTelegramUserClient()
	case "feishu", "lark":
		return newFeishuClient()
	default:
		return nil, fmt.Errorf("暂不支持的平台: %s", name)
	}
}

// selectedPlatforms 同时兼容单平台 platform 和多平台 platforms 配置。
func selectedPlatforms() []string {
	var configured []string
	if err := viper.UnmarshalKey("platforms", &configured); err != nil {
		configured = nil
	}
	if len(configured) == 0 {
		configured = viper.GetStringSlice("platforms")
	}
	if len(configured) == 0 {
		configured = []string{viper.GetString("platform")}
	}

	seen := make(map[string]struct{}, len(configured))
	result := make([]string, 0, len(configured))
	for _, item := range configured {
		name := canonicalPlatformName(item)
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	if len(result) == 0 {
		return []string{"qq"}
	}
	return result
}

func canonicalPlatformName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "-", "_")))
	switch normalized {
	case "", "qq", "napcat", "qq_napcat":
		return "qq"
	case "tg_bot":
		return "telegram_bot"
	case "tg_user":
		return "telegram_user"
	case "lark":
		return "feishu"
	default:
		return normalized
	}
}

func newQQNapcatClient() (base.Client, error) {
	selfID := firstNonEmpty(viper.GetString("bot_config.self_id"), viper.GetString("selfId"))
	if selfID == "" && config.SelfId != 0 {
		selfID = strconv.FormatInt(config.SelfId, 10)
	}

	cfg := qqnapcat.Config{
		APIURL:               firstNonEmpty(viper.GetString("bot_config.api_url"), viper.GetString("apiUrl")),
		Token:                firstNonEmpty(viper.GetString("bot_config.token"), viper.GetString("token")),
		SelfID:               selfID,
		HeartbeatTimeout:     secondsToDuration(viper.GetInt64("bot_config.heartbeat_timeout")),
		ReconnectMaxAttempts: int(viper.GetInt64("bot_config.reconnect_max_attempts")),
		DialTimeout:          durationFromConfig("bot_config.dial_timeout"),
		RequestTimeout:       durationFromConfig("bot_config.request_timeout"),
		EventBuffer:          viper.GetInt("bot_config.event_buffer"),
		WriteTimeout:         durationFromConfig("bot_config.write_timeout"),
	}

	return qqnapcat.New(cfg, qqnapcat.Dependencies{
		Logger: newLogger(),
	})
}

func newTelegramBotClient() (base.Client, error) {
	cfg := tgbot.Config{
		Token:          firstNonEmpty(viper.GetString("telegram_bot.token"), viper.GetString("telegram.token")),
		UpdateTimeout:  viper.GetInt("telegram_bot.update_timeout"),
		AllowedUpdates: viper.GetStringSlice("telegram_bot.allowed_updates"),
		EventBuffer:    viper.GetInt("telegram_bot.event_buffer"),
	}
	return tgbot.New(cfg, tgbot.Dependencies{
		Logger: newLogger(),
	})
}

func newTelegramUserClient() (base.Client, error) {
	cfg := tguser.Config{
		APIID:          firstNonZeroInt(viper.GetInt("telegram_user.api_id"), viper.GetInt("telegram.api_id")),
		APIHash:        firstNonEmpty(viper.GetString("telegram_user.api_hash"), viper.GetString("telegram.api_hash")),
		SessionPath:    firstNonEmpty(viper.GetString("telegram_user.session_path"), viper.GetString("telegram.session_path")),
		SourceChats:    firstNonEmptyStringSlice(viper.GetStringSlice("telegram_user.source_chats"), viper.GetStringSlice("telegram.source_chats")),
		AuthMode:       tguser.AuthMode(firstNonEmpty(viper.GetString("telegram_user.auth_mode"), viper.GetString("telegram.auth_mode"))),
		EventBuffer:    viper.GetInt("telegram_user.event_buffer"),
		RequestTimeout: durationFromConfig("telegram_user.request_timeout"),
	}
	return tguser.New(cfg, tguser.Dependencies{
		Logger:       newLogger(),
		AuthProvider: consoleTelegramAuth{},
	})
}

func newFeishuClient() (base.Client, error) {
	cfg := feishu.Config{
		AppID:             firstNonEmpty(viper.GetString("feishu.app_id"), viper.GetString("feishu.appID")),
		AppSecret:         viper.GetString("feishu.app_secret"),
		VerificationToken: viper.GetString("feishu.verification_token"),
		EncryptKey:        viper.GetString("feishu.encrypt_key"),
		ReceiveIDType:     viper.GetString("feishu.receive_id_type"),
		EventBuffer:       viper.GetInt("feishu.event_buffer"),
		RequestTimeout:    durationFromConfig("feishu.request_timeout"),
	}
	return feishu.New(cfg, feishu.Dependencies{
		Logger: newLogger(),
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

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonEmptyStringSlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func secondsToDuration(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func durationFromConfig(key string) time.Duration {
	if value := viper.GetDuration(key); value > 0 {
		return value
	}
	return secondsToDuration(viper.GetInt64(key))
}
