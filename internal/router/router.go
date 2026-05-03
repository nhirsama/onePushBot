package router

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	pubsub "github.com/cskr/pubsub/v2"
	base "github.com/nhirsama/onePushBot/internal/platform"
)

type topic string

const (
	topicAll     topic = "all"
	topicMessage topic = "kind:message"
	topicNotice  topic = "kind:notice"
	topicRequest topic = "kind:request"
	topicSystem  topic = "kind:system"
	topicRaw     topic = "kind:raw"
)

type broker interface {
	Sub(topics ...topic) chan base.Event
	Pub(msg base.Event, topics ...topic)
	Unsub(ch chan base.Event, topics ...topic)
	Shutdown()
}

type pubsubBroker struct {
	ps *pubsub.PubSub[topic, base.Event]
}

func newBroker(capacity int) broker {
	if capacity <= 0 {
		capacity = 128
	}
	return &pubsubBroker{ps: pubsub.New[topic, base.Event](capacity)}
}

func (b *pubsubBroker) Sub(topics ...topic) chan base.Event {
	return b.ps.Sub(topics...)
}

func (b *pubsubBroker) Pub(msg base.Event, topics ...topic) {
	b.ps.Pub(msg, topics...)
}

func (b *pubsubBroker) Unsub(ch chan base.Event, topics ...topic) {
	b.ps.Unsub(ch, topics...)
}

func (b *pubsubBroker) Shutdown() {
	b.ps.Shutdown()
}

// Stats 是 Router 的运行时统计快照。
type Stats struct {
	// Received 是从平台总线读到的事件数量。
	Received uint64
	// Published 是发布到内部 topic broker 的事件数量。
	Published uint64
	// Routes 是当前注册的路由数量。
	Routes int
	// RouteSummaries 按路由名称记录单条路由的统计快照。
	RouteSummaries map[string]RouteStats
}

// Router 从平台总线订阅事件，并把事件分发给已注册的路由。
type Router interface {
	// Register 注册一条路由。路由器启动后不能再注册。
	Register(route Route) error

	// Start 启动路由器并开始订阅平台总线。
	Start(ctx context.Context) error

	// Close 停止接收新事件并等待路由 worker 退出。
	Close(ctx context.Context) error

	// Stats 返回当前统计快照。
	Stats() Stats
}

type router struct {
	bus    base.Bus
	broker broker

	mu      sync.RWMutex
	routes  map[string]*routeRunner
	started bool
	closed  bool
	cancel  context.CancelFunc
	sub     base.Subscription
	wg      sync.WaitGroup

	received  atomic.Uint64
	published atomic.Uint64
}

// Options 描述 Router 的构造选项。
type Options struct {
	// BrokerBuffer 是内部 topic 订阅通道的缓冲大小。
	BrokerBuffer int
}

// New 创建一个 Router。
//
// Router 只依赖平台总线的稳定事件契约；平台专属能力应通过 route handler
// 的依赖注入获取。
func New(bus base.Bus, opts Options) (Router, error) {
	if bus == nil {
		return nil, fmt.Errorf("router bus is required")
	}
	return &router{
		bus:    bus,
		broker: newBroker(opts.BrokerBuffer),
		routes: make(map[string]*routeRunner),
	}, nil
}

func (r *router) Register(route Route) error {
	cfg := route.withDefaults()
	if err := cfg.validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRouterClosed
	}
	if r.started {
		return ErrRouterStarted
	}
	if _, ok := r.routes[cfg.Name]; ok {
		return ErrRouteExists
	}

	r.routes[cfg.Name] = newRouteRunner(cfg)
	return nil
}

func (r *router) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrRouterClosed
	}
	if r.started {
		r.mu.Unlock()
		return ErrRouterStarted
	}

	runCtx, cancel := context.WithCancel(ctx)
	sub, err := r.bus.Subscribe(base.EventFilter{})
	if err != nil {
		cancel()
		r.mu.Unlock()
		return err
	}

	r.cancel = cancel
	r.sub = sub
	r.started = true

	routes := make([]*routeRunner, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, route)
	}
	r.mu.Unlock()

	for _, route := range routes {
		route.start(runCtx, r.broker)
	}

	r.wg.Add(1)
	go r.forward(runCtx)
	return nil
}

func (r *router) forward(ctx context.Context) {
	defer r.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-r.sub.Events():
			if !ok {
				return
			}
			r.received.Add(1)
			// The internal broker keeps route subscription fan-out local to the
			// router, so platform.Bus only needs one upstream subscription.
			topics := topicsForEvent(event)
			r.broker.Pub(event, topics...)
			r.published.Add(1)
		}
	}
}

func (r *router) Close(ctx context.Context) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	cancel := r.cancel
	sub := r.sub
	r.cancel = nil
	r.sub = nil
	r.started = false

	routes := make([]*routeRunner, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, route)
	}
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if sub != nil {
		sub.Unsubscribe()
	}

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		for _, route := range routes {
			if route.subCh != nil {
				// cskr/pubsub requires Unsub to run outside the subscriber
				// goroutine while the subscriber drains until channel close.
				go r.broker.Unsub(route.subCh, route.topics...)
			}
		}
		for _, route := range routes {
			route.wait()
		}
		r.broker.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *router) Stats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routeSummaries := make(map[string]RouteStats, len(r.routes))
	for name, route := range r.routes {
		routeSummaries[name] = route.stats()
	}

	return Stats{
		Received:       r.received.Load(),
		Published:      r.published.Load(),
		Routes:         len(r.routes),
		RouteSummaries: routeSummaries,
	}
}

func topicsForFilter(filter base.EventFilter) []topic {
	if filter.Kind == "" {
		return []topic{topicAll}
	}

	switch filter.Kind {
	case base.EventKindMessage:
		return []topic{topicMessage}
	case base.EventKindNotice:
		return []topic{topicNotice}
	case base.EventKindRequest:
		return []topic{topicRequest}
	case base.EventKindSystem:
		return []topic{topicSystem}
	case base.EventKindRaw:
		return []topic{topicRaw}
	default:
		return []topic{topicAll}
	}
}

func topicsForEvent(event base.Event) []topic {
	topics := []topic{topicAll}
	switch event.Kind {
	case base.EventKindMessage:
		topics = append(topics, topicMessage)
	case base.EventKindNotice:
		topics = append(topics, topicNotice)
	case base.EventKindRequest:
		topics = append(topics, topicRequest)
	case base.EventKindSystem:
		topics = append(topics, topicSystem)
	case base.EventKindRaw:
		topics = append(topics, topicRaw)
	}
	return topics
}
