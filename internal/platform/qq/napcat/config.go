package napcat

import "time"

// Config 描述 NapCat 客户端运行时配置。
type Config struct {
	APIURL               string
	Token                string
	SelfID               string
	HeartbeatTimeout     time.Duration
	ReconnectMaxAttempts int
	DialTimeout          time.Duration
	RequestTimeout       time.Duration
	EventBuffer          int
	WriteTimeout         time.Duration
}

func (c Config) withDefaults() Config {
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = 90 * time.Second
	}
	if c.DialTimeout <= 0 {
		c.DialTimeout = 15 * time.Second
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 60 * time.Second
	}
	if c.EventBuffer <= 0 {
		c.EventBuffer = 128
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 10 * time.Second
	}
	return c
}
