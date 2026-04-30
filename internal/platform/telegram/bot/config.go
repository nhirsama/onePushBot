package bot

// Config 描述 Telegram Bot 客户端配置。
type Config struct {
	Token          string
	UpdateTimeout  int
	AllowedUpdates []string
	EventBuffer    int
}

func (c Config) withDefaults() Config {
	if c.UpdateTimeout <= 0 {
		c.UpdateTimeout = 30
	}
	if c.EventBuffer <= 0 {
		c.EventBuffer = 128
	}
	return c
}
