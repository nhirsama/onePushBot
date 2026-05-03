package user

import "time"

// AuthMode 描述 Telegram 用户态首次授权方式。
type AuthMode string

const (
	// AuthModeQR 优先使用 Telegram 客户端扫码登录。
	AuthModeQR AuthMode = "qr"
	// AuthModeCode 使用手机号和验证码登录。
	AuthModeCode AuthMode = "code"
)

// Config 描述 Telegram 用户态客户端配置。
type Config struct {
	APIID          int
	APIHash        string
	SessionPath    string
	SourceChats    []string
	AuthMode       AuthMode
	EventBuffer    int
	RequestTimeout time.Duration
}

func (c Config) withDefaults() Config {
	if c.AuthMode == "" {
		c.AuthMode = AuthModeQR
	}
	if c.EventBuffer <= 0 {
		c.EventBuffer = 128
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}
	return c
}
