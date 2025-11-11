package domain

type BotConfig struct {
	Type                 string `mapstructure:"bot_config.type"`
	ApiUrl               string `mapstructure:"bot_config.api_url"`
	Token                string `mapstructure:"bot_config.token"`
	HeartbeatTimeout     int64  `mapstructure:"bot_config.heartbeat_timeout"`
	ReconnectMaxAttempts int64  `mapstructure:"bot_config.reconnect_max_attempts"`
}
