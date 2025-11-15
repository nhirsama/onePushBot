package domain

import (
	"github.com/asaskevich/EventBus"
	internalDomain "github.com/nhirsama/onePushBot/internal/domain"
)

// BotClient 对外暴露的机器人客户端接口
type BotClient interface {
	// Start 启动机器人客户端：
	//  - 建立 WebSocket 连接
	//  - 启动心跳监控
	Start() error

	// Close 关闭客户端（关闭连接并停止内部协程）
	Close() error

	// Bus 返回事件总线，外部可以通过订阅 topic 来获取消息：
	Bus() EventBus.Bus

	// Stats 返回当前分发统计数据
	Stats() internalDomain.DispatcherStats

	// ConnectionManager 暴露机器人连接状态查询能力（只读接口）
	ConnectionManager() internalDomain.ConnectionManager
}
