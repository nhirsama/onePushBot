package platform

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// BusOption 用于调整内存总线行为。
type BusOption func(*memoryBus)

// WithSubscriptionBuffer 调整每个订阅的缓冲区大小。
func WithSubscriptionBuffer(size int) BusOption {
	return func(bus *memoryBus) {
		if size > 0 {
			bus.bufferSize = size
		}
	}
}

type memoryBus struct {
	mu         sync.RWMutex
	closed     bool
	bufferSize int
	counter    atomic.Uint64
	subs       map[string]*memorySubscription

	published atomic.Uint64
	delivered atomic.Uint64
	dropped   atomic.Uint64
	invalid   atomic.Uint64
}

type memorySubscription struct {
	id     string
	filter EventFilter
	ch     chan Event
	bus    *memoryBus

	mu     sync.RWMutex
	closed bool
}

// NewMemoryBus 创建一个进程内事件总线。
func NewMemoryBus(opts ...BusOption) Bus {
	bus := &memoryBus{
		bufferSize: 64,
		subs:       make(map[string]*memorySubscription),
	}
	for _, opt := range opts {
		opt(bus)
	}
	return bus
}

func (b *memoryBus) Publish(ctx context.Context, event Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := event.Validate(); err != nil {
		b.invalid.Add(1)
		return err
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}

	subs := make([]*memorySubscription, 0, len(b.subs))
	for _, sub := range b.subs {
		if sub.filter.Match(event) {
			subs = append(subs, sub)
		}
	}
	b.mu.RUnlock()

	for _, sub := range subs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if sub.deliver(event) {
			b.delivered.Add(1)
		} else {
			b.dropped.Add(1)
		}
	}

	b.published.Add(1)
	return nil
}

func (b *memoryBus) Subscribe(filter EventFilter) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrBusClosed
	}

	id := strconv.FormatUint(b.counter.Add(1), 10) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	sub := &memorySubscription{
		id:     id,
		filter: filter,
		ch:     make(chan Event, b.bufferSize),
		bus:    b,
	}
	b.subs[id] = sub
	return sub, nil
}

func (b *memoryBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := make([]*memorySubscription, 0, len(b.subs))
	for _, sub := range b.subs {
		subs = append(subs, sub)
	}
	b.subs = make(map[string]*memorySubscription)
	b.mu.Unlock()

	for _, sub := range subs {
		sub.close()
	}
	return nil
}

func (b *memoryBus) Stats() BusStats {
	b.mu.RLock()
	activeSubscriptions := len(b.subs)
	b.mu.RUnlock()

	return BusStats{
		ActiveSubscriptions: uint64(activeSubscriptions),
		Published:           b.published.Load(),
		Delivered:           b.delivered.Load(),
		Dropped:             b.dropped.Load(),
		Invalid:             b.invalid.Load(),
	}
}

func (s *memorySubscription) ID() string {
	return s.id
}

func (s *memorySubscription) Events() <-chan Event {
	return s.ch
}

func (s *memorySubscription) Unsubscribe() {
	s.bus.removeSubscription(s.id)
}

func (s *memorySubscription) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true
	close(s.ch)
}

func (s *memorySubscription) deliver(event Event) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return false
	}

	select {
	case s.ch <- event:
		return true
	default:
		return false
	}
}

func (b *memoryBus) removeSubscription(id string) {
	b.mu.Lock()
	sub, ok := b.subs[id]
	if ok {
		delete(b.subs, id)
	}
	b.mu.Unlock()

	if ok {
		sub.close()
	}
}
