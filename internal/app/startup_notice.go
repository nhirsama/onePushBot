package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
	tgbot "github.com/nhirsama/onePushBot/internal/platform/telegram/bot"
	"github.com/spf13/viper"
)

func (a *App) notifyStarted(ctx context.Context) {
	ownerChatID := strings.TrimSpace(viper.GetString("telegram_bot.owner_chat_id"))
	if ownerChatID == "" || a == nil {
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
