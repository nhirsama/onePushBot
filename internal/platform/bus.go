package platform

import "context"

// Bus 负责在平台层内部做统一事件发布与订阅。
type Bus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(filter EventFilter) (Subscription, error)
	Close() error
}

// Subscription 表示一个事件订阅。
type Subscription interface {
	ID() string
	Events() <-chan Event
	Unsubscribe()
}
