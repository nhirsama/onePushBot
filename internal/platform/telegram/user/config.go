package user

import "time"

// Config 描述 Telegram 用户态客户端配置。
type Config struct {
	APIID          int
	APIHash        string
	SessionPath    string
	SourceChats    []string
	EventBuffer    int
	RequestTimeout time.Duration
}

func (c Config) withDefaults() Config {
	if c.EventBuffer <= 0 {
		c.EventBuffer = 128
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}
	return c
}
