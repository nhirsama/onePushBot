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
	selfID := firstNonEmpty(viper.GetString("bot_config.self_id"), viper.GetString("selfId"))

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
		AuthProvider: config.NewTelegramAuthProvider(),
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
