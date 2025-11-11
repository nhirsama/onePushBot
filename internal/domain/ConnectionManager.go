package domain

import (
	"context"
	"time"
)

// ConnectionManager 连接管理器接口
type ConnectionManager interface {
	// Connect 连接管理
	Connect(ctx context.Context) error
	Reconnect(ctx context.Context) error
	Close() error

	// IsConnected 状态查询
	IsConnected() bool
	GetConnectTime() time.Time
	GetReconnectCount() int

	// StartHeartbeat 心跳管理
	StartHeartbeat(ctx context.Context)
	OnHeartbeat(callback func())
	NotifyHeartbeat()
}
