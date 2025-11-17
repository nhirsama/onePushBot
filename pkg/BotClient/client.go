package BotClient

import (
	"context"
	"fmt"

	"github.com/asaskevich/EventBus"
	internalDomain "github.com/nhirsama/onePushBot/internal/domain"
	"github.com/nhirsama/onePushBot/internal/infrastructure/bot/napcat"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
	"github.com/nhirsama/onePushBot/pkg/logger"
)

// client 是 Napcat Client 的具体实现，隐藏在 internal 之外，仅通过 Client 接口对外暴露。
type client struct {
	cfg        *internalDomain.BotConfig
	log        pkgDomain.Log
	bus        EventBus.Bus
	connMgr    internalDomain.ConnectionManager
	dispatcher internalDomain.MessageDispatcher
	api        internalDomain.ApiDistribute
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewClient 创建一个 Napcat 客户端，并在内部把各组件全部初始化好。
// 依赖注入关系：
//   - ConnectionManager 需要 BotConfig + Log
//   - EventBus 被 MessageDispatcher 用来发布消息事件
//   - MessageDispatcher 需要 Log + ConnectionManager + EventBus
//
// 对外只暴露 Client 接口，调用方不再直接依赖 internal/napcat。
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
	// 1. 创建事件总线（供 dispatcher 发布事件）
	bus := EventBus.New()
	log := logger.NewLog(pkgDomain.DebugLevel)
	// 2. 创建连接管理器（负责连接 / 重连 / 心跳）—— internal/napcat 实现
	connMgr := napcat.NewConnectionManager(ctx, log, cfg)

	// 3. 创建消息分发器（负责根据 post_type 分发消息）—— internal/napcat 实现
	dispatcher := napcat.NewMessageDispatcher(log, connMgr, &bus)

	return &client{
		cfg:        cfg,
		log:        log,
		bus:        bus,
		connMgr:    connMgr,
		dispatcher: dispatcher,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start 启动 Napcat 客户端：
//   - 建立 WebSocket 连接
//   - 启动心跳监控
func (c *client) Start() error {

	// 1. 建立连接
	if err := c.connMgr.Connect(c.ctx); err != nil {
		c.log.Error("Napcat 连接失败", "err", err)
		return err
	}

	// 2. 启动心跳监控
	c.connMgr.StartHeartbeat(c.ctx)
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case msg := <-c.connMgr.ReadChan():
				go c.dispatcher.Dispatch(c.ctx, msg)
			}
		}
	}()
	c.log.Info("Napcat 客户端启动完成")
	return nil
}

// Close 关闭 Napcat 客户端，释放连接和内部协程。
func (c *client) Close() error {
	c.log.Info("正在关闭 Napcat 客户端")
	return c.connMgr.Close()
}

// Bus 返回事件总线，外部可以订阅消息事件。
func (c *client) Bus() EventBus.Bus {
	return c.bus
}

// Stats 获取当前消息分发统计数据。
func (c *client) Stats() internalDomain.DispatcherStats {
	return c.dispatcher.GetStats()
}

// ConnectionManager 暴露连接管理器（只读使用为主，用于状态查询）。
func (c *client) ConnectionManager() internalDomain.ConnectionManager {
	return c.connMgr
}
