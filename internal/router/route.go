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
	ErrRouterStarted     = errors.New("router already started")
	ErrRouterClosed      = errors.New("router already closed")
	ErrRouteExists       = errors.New("router route already registered")
	ErrRouteInvalid      = errors.New("router route is invalid")
	ErrRouteHandlerNil   = errors.New("router route handler is nil")
	ErrRouteNameRequired = errors.New("router route name is required")
)

type MatchFunc func(base.Event) bool

type Handler interface {
	Handle(ctx context.Context, event base.Event) error
}

type HandlerFunc func(ctx context.Context, event base.Event) error

func (f HandlerFunc) Handle(ctx context.Context, event base.Event) error {
	return f(ctx, event)
}

type OverflowPolicy string

const (
	OverflowDropNewest OverflowPolicy = "drop_newest"
	OverflowBlock      OverflowPolicy = "block"
)

type Route struct {
	Name           string
	Filter         base.EventFilter
	Match          MatchFunc
	Handler        Handler
	Workers        int
	QueueSize      int
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

type RouteStats struct {
	Received uint64
	Matched  uint64
	Enqueued uint64
	Dropped  uint64
	Handled  uint64
	Failed   uint64
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
