package cmd

import (
	"fmt"
	"strings"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/spf13/viper"
)

// newPlatformHub 是 cmd 的平台组合入口，只负责把配置中的平台客户端注册进 Hub。
func newPlatformHub(logger base.Logger) (base.Hub, error) {
	if logger == nil {
		logger = newLogger()
	}
	hub := base.NewHubWithOptions(nil, base.WithHubLogger(logger))

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
	case "qq", "napcat", "qq_napcat":
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

func selectedPlatforms() []string {
	var configured []string
	if err := viper.UnmarshalKey("platforms", &configured); err != nil {
		configured = nil
	}

	seen := make(map[string]struct{}, len(configured))
	result := make([]string, 0, len(configured))
	for _, item := range configured {
		name := canonicalPlatformName(item)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

func canonicalPlatformName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "-", "_")))
	switch normalized {
	case "":
		return ""
	case "qq", "napcat", "qq_napcat":
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
