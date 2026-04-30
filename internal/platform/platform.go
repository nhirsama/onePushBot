package platform

import "context"

// Platform 标识具体平台实现。
type Platform string

const (
	PlatformQQ           Platform = "qq"
	PlatformTelegramUser Platform = "telegram_user"
	PlatformTelegramBot  Platform = "telegram_bot"
	PlatformFeishu       Platform = "feishu"
)

// Status 描述平台客户端当前生命周期状态。
type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusClosing  Status = "closing"
	StatusClosed   Status = "closed"
	StatusError    Status = "error"
)

// Client 是平台客户端对外暴露的最小公共生命周期接口。
type Client interface {
	Platform() Platform
	Start(ctx context.Context) error
	Close(ctx context.Context) error
	Events() <-chan Event
	Status() Status
}
