package router

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

var (
	// ErrRouterStarted 表示路由器已经启动，不能再注册路由或重复启动。
	ErrRouterStarted = errors.New("router already started")
	// ErrRouterClosed 表示路由器已经关闭。
	ErrRouterClosed = errors.New("router already closed")
	// ErrRouteExists 表示注册了重复名称的路由。
	ErrRouteExists = errors.New("router route already registered")
	// ErrRouteInvalid 表示路由配置不合法。
	ErrRouteInvalid = errors.New("router route is invalid")
	// ErrRouteHandlerNil 表示路由没有配置处理器。
	ErrRouteHandlerNil = errors.New("router route handler is nil")
	// ErrRouteNameRequired 表示路由名称为空。
	ErrRouteNameRequired = errors.New("router route name is required")
)

// MatchFunc 在 EventFilter 之后执行，用于表达平台或业务相关的细粒度匹配。
type MatchFunc func(base.Event) bool

// Handler 处理命中路由的事件。
type Handler interface {
	Handle(ctx context.Context, event base.Event) error
}

// HandlerFunc 允许普通函数作为 Handler 使用。
type HandlerFunc func(ctx context.Context, event base.Event) error

// Handle 调用 f(ctx, event)。
func (f HandlerFunc) Handle(ctx context.Context, event base.Event) error {
	return f(ctx, event)
}

// OverflowPolicy 描述路由队列满时的处理策略。
type OverflowPolicy string

const (
	// OverflowDropNewest 在路由队列满时丢弃当前事件。
	OverflowDropNewest OverflowPolicy = "drop_newest"
	// OverflowBlock 在路由队列满时阻塞等待空间。
	OverflowBlock OverflowPolicy = "block"
)

// Route 描述一条事件路由。
//
// Filter 负责平台公共字段匹配，Match 负责补充业务或平台专属匹配。
// Handler 在事件进入该路由的本地队列后由 worker 调用。
type Route struct {
	// Name 是路由的唯一名称，用于注册去重和统计索引。
	Name string
	// Filter 使用平台事件的公共字段做第一层匹配。
	Filter base.EventFilter
	// Match 在 Filter 命中后执行，用于补充平台或业务专属条件。
	Match MatchFunc
	// Handler 处理进入该路由队列的事件。
	Handler Handler
	// Workers 是该路由并发消费事件的 worker 数量。
	Workers int
	// QueueSize 是该路由本地队列的缓冲大小。
	QueueSize int
	// OverflowPolicy 决定路由本地队列满时如何处理新事件。
	OverflowPolicy OverflowPolicy
}

func (r Route) withDefaults() Route {
	if r.Workers <= 0 {
		r.Workers = 1
	}
	if r.QueueSize <= 0 {
		r.QueueSize = 64
	}
	if r.OverflowPolicy == "" {
		r.OverflowPolicy = OverflowDropNewest
	}
	return r
}

func (r Route) validate() error {
	if r.Name == "" {
		return ErrRouteNameRequired
	}
	if r.Handler == nil {
		return ErrRouteHandlerNil
	}
	switch r.OverflowPolicy {
	case "", OverflowDropNewest, OverflowBlock:
	default:
		return fmt.Errorf("%w: unsupported overflow policy %q", ErrRouteInvalid, r.OverflowPolicy)
	}
	return nil
}

// RouteStats 是单条路由的运行时统计快照。
type RouteStats struct {
	// Received 是内部 broker 投递到该路由订阅通道的事件数量。
	Received uint64
	// Matched 是通过 Filter 和 Match 的事件数量。
	Matched uint64
	// Enqueued 是成功进入该路由本地队列的事件数量。
	Enqueued uint64
	// Dropped 是因队列溢出或关闭而没有进入本地队列的事件数量。
	Dropped uint64
	// Handled 是 Handler 已处理完成的事件数量。
	Handled uint64
	// Failed 是 Handler 返回错误的事件数量。
	Failed uint64
}

type routeRunner struct {
	cfg Route

	subCh  chan base.Event
	topics []topic
	inbox  chan base.Event

	wg sync.WaitGroup

	received atomic.Uint64
	matched  atomic.Uint64
	enqueued atomic.Uint64
	dropped  atomic.Uint64
	handled  atomic.Uint64
	failed   atomic.Uint64
}

func newRouteRunner(route Route) *routeRunner {
	cfg := route.withDefaults()
	return &routeRunner{
		cfg:    cfg,
		topics: topicsForFilter(cfg.Filter),
		inbox:  make(chan base.Event, cfg.QueueSize),
	}
}

func (r *routeRunner) start(ctx context.Context, ps broker) {
	r.subCh = ps.Sub(r.topics...)

	r.wg.Add(1)
	go r.dispatch(ctx)

	for i := 0; i < r.cfg.Workers; i++ {
		r.wg.Add(1)
		go r.worker(ctx)
	}
}

func (r *routeRunner) dispatch(ctx context.Context) {
	defer r.wg.Done()
	defer close(r.inbox)

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-r.subCh:
			if !ok {
				return
			}
			r.received.Add(1)
			if !r.cfg.Filter.Match(event) {
				continue
			}
			if r.cfg.Match != nil && !r.cfg.Match(event) {
				continue
			}
			r.matched.Add(1)
			if r.enqueue(ctx, event) {
				r.enqueued.Add(1)
			} else {
				r.dropped.Add(1)
			}
		}
	}
}

func (r *routeRunner) enqueue(ctx context.Context, event base.Event) bool {
	switch r.cfg.OverflowPolicy {
	case OverflowBlock:
		select {
		case r.inbox <- event:
			return true
		case <-ctx.Done():
			return false
		}
	default:
		select {
		case r.inbox <- event:
			return true
		default:
			return false
		}
	}
}

func (r *routeRunner) worker(ctx context.Context) {
	defer r.wg.Done()

	for event := range r.inbox {
		if err := r.cfg.Handler.Handle(ctx, event); err != nil {
			r.failed.Add(1)
		}
		r.handled.Add(1)
	}
}

func (r *routeRunner) wait() {
	r.wg.Wait()
}

func (r *routeRunner) stats() RouteStats {
	return RouteStats{
		Received: r.received.Load(),
		Matched:  r.matched.Load(),
		Enqueued: r.enqueued.Load(),
		Dropped:  r.dropped.Load(),
		Handled:  r.handled.Load(),
		Failed:   r.failed.Load(),
	}
}
