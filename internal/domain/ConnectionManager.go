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
	// ReadChan 返回只读通道，上层通过 <-ReadChan() 获取从 WebSocket 收到的字节流
	ReadChan() <-chan []byte

	// WriteChan 返回写通道，上层通过 WriteChan() <- data 发送字节流到 WebSocket
	WriteChan() chan<- []byte
}
