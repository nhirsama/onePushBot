package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
	tgbot "github.com/nhirsama/onePushBot/internal/platform/telegram/bot"
	"github.com/spf13/viper"
)

// notifyStarted 在平台和外部入口启动后发送状态通知；未配置目标时直接跳过。
func (a *app) notifyStarted(ctx context.Context) {
	ownerChatID := firstNonEmpty(viper.GetString("telegram_bot.owner_chat_id"))
	if ownerChatID == "" {
		return
	}

	client, ok := a.hub.Get(base.PlatformTelegramBot)
	if !ok {
		return
	}

	botClient, ok := client.(tgbot.Client)
	if !ok {
		return
	}

	go func() {
		notifyCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := botClient.SendText(notifyCtx, ownerChatID, startupMessage()); err != nil {
			log.Printf("发送 Telegram Bot 启动通知失败: %v", err)
		}
	}()
}

func startupMessage() string {
	return fmt.Sprintf("onePushBot 已启动\n时间: %s", time.Now().Format("2006-01-02 15:04:05"))
}
