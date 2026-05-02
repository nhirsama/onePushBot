package platform

import "context"

// Bus 负责在平台层内部做统一事件发布与订阅。
type Bus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(filter EventFilter) (Subscription, error)
	Close() error
	Stats() BusStats
}

// Subscription 表示一个事件订阅。
type Subscription interface {
	ID() string
	Events() <-chan Event
	Unsubscribe()
}

// BusStats 描述总线运行时快照。
type BusStats struct {
	ActiveSubscriptions uint64
	Published           uint64
	Delivered           uint64
	Dropped             uint64
	Invalid             uint64
}
