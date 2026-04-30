package feishu

import "time"

// Config 描述飞书平台配置。
type Config struct {
	AppID             string
	AppSecret         string
	VerificationToken string
	EncryptKey        string
	ReceiveIDType     string
	EventBuffer       int
	RequestTimeout    time.Duration
}

func (c Config) withDefaults() Config {
	if c.ReceiveIDType == "" {
		c.ReceiveIDType = "chat_id"
	}
	if c.EventBuffer <= 0 {
		c.EventBuffer = 128
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 10 * time.Second
	}
	return c
}
