package app

import (
	"fmt"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/spf13/viper"
)

func newPlatformHub(logger base.Logger) (base.Hub, error) {
	if logger == nil {
		logger = base.NewDiscardLogger()
	}
	hub := base.NewHubWithOptions(nil, base.WithHubLogger(logger))

	clients, err := newPlatformClients(logger)
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

func newPlatformClients(logger base.Logger) ([]base.Client, error) {
	platforms := selectedPlatforms()
	clients := make([]base.Client, 0, len(platforms))

	for _, platform := range platforms {
		client, err := newPlatformClient(platform, logger)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func newPlatformClient(name string, logger base.Logger) (base.Client, error) {
	switch name {
	case "qq":
		return newQQNapcatClient(logger)
	case "telegram_bot":
		return newTelegramBotClient(logger)
	case "telegram_user":
		return newTelegramUserClient(logger)
	case "feishu":
		return newFeishuClient(logger)
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
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
