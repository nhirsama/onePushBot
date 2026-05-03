package BotClient

import (
	"context"
	"fmt"

	"github.com/asaskevich/EventBus"
	internalDomain "github.com/nhirsama/onePushBot/internal/domain"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
)

// client 是旧 pkg BotClient 的占位实现。
//
// 新启动链路已经迁移到 internal/platform，旧 NapCat infrastructure 已移除。
type client struct {
	cfg    *internalDomain.BotConfig
	bus    EventBus.Bus
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient 保留旧 pkg 构造入口，但不再接入已删除的旧内核实现。
func NewClient(
	parentCtx context.Context,
	cfg *internalDomain.BotConfig,
) (pkgDomain.BotClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("napcat.BotConfig 不能为空")
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	bus := EventBus.New()

	return &client{
		cfg:    cfg,
		bus:    bus,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start 不再启动旧 infrastructure，调用方应迁移到 cmd/internal/platform。
func (c *client) Start() error {
	return fmt.Errorf("旧 pkg/BotClient 已停用，请使用 internal/platform")
}

// Close 关闭 Napcat 客户端，释放连接和内部协程。
func (c *client) Close() error {
	c.cancel()
	return nil
}

// Bus 返回事件总线，外部可以订阅消息事件。
func (c *client) Bus() EventBus.Bus {
	return c.bus
}

// Stats 获取当前消息分发统计数据。
func (c *client) Stats() internalDomain.DispatcherStats {
	return internalDomain.DispatcherStats{}
}

// ConnectionManager 暴露连接管理器（只读使用为主，用于状态查询）。
func (c *client) ConnectionManager() internalDomain.ConnectionManager {
	return nil
}
