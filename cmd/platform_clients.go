package cmd

import (
	"github.com/nhirsama/onePushBot/config"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/feishu"
	qqnapcat "github.com/nhirsama/onePushBot/internal/platform/qq/napcat"
	tgbot "github.com/nhirsama/onePushBot/internal/platform/telegram/bot"
	tguser "github.com/nhirsama/onePushBot/internal/platform/telegram/user"
	"github.com/spf13/viper"
)

func newQQNapcatClient() (base.Client, error) {
	cfg := qqnapcat.Config{
		APIURL:               viper.GetString("qq.api_url"),
		Token:                viper.GetString("qq.token"),
		SelfID:               viper.GetString("qq.self_id"),
		HeartbeatTimeout:     secondsToDuration(viper.GetInt64("qq.heartbeat_timeout")),
		ReconnectMaxAttempts: viper.GetInt("qq.reconnect_max_attempts"),
		DialTimeout:          durationFromConfig("qq.dial_timeout"),
		RequestTimeout:       durationFromConfig("qq.request_timeout"),
		EventBuffer:          viper.GetInt("qq.event_buffer"),
		WriteTimeout:         durationFromConfig("qq.write_timeout"),
	}

	return qqnapcat.New(cfg, qqnapcat.Dependencies{
		Logger: newLogger(),
	})
}

func newTelegramBotClient() (base.Client, error) {
	cfg := tgbot.Config{
		Token:          viper.GetString("telegram_bot.token"),
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
		APIID:          viper.GetInt("telegram_user.api_id"),
		APIHash:        viper.GetString("telegram_user.api_hash"),
		SessionPath:    viper.GetString("telegram_user.session_path"),
		SourceChats:    viper.GetStringSlice("telegram_user.source_chats"),
		AuthMode:       tguser.AuthMode(viper.GetString("telegram_user.auth_mode")),
		EventBuffer:    viper.GetInt("telegram_user.event_buffer"),
		RequestTimeout: durationFromConfig("telegram_user.request_timeout"),
	}
	return tguser.New(cfg, tguser.Dependencies{
		Logger:       newLogger(),
		AuthProvider: config.NewTelegramAuthProvider(),
	})
}

func newFeishuClient() (base.Client, error) {
	cfg := feishu.Config{
		AppID:             viper.GetString("feishu.app_id"),
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
