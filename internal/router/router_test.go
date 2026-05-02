package router

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

type captureHandler struct {
	mu     sync.Mutex
	events []base.Event
	wait   chan struct{}
	err    error
}

func (h *captureHandler) Handle(ctx context.Context, event base.Event) error {
	if h.wait != nil {
		select {
		case <-h.wait:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	h.mu.Lock()
	h.events = append(h.events, event)
	h.mu.Unlock()
	return h.err
}

func (h *captureHandler) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

func TestRouterRegisterAndDispatch(t *testing.T) {
	bus := base.NewMemoryBus()
	rt, err := New(bus, Options{BrokerBuffer: 16})
	if err != nil {
		t.Fatalf("创建 router 失败: %v", err)
	}

	handler := &captureHandler{}
	if err := rt.Register(Route{
		Name:    "qq-group-message",
		Filter:  base.EventFilter{Platform: base.PlatformQQ, Kind: base.EventKindMessage},
		Handler: handler,
	}); err != nil {
		t.Fatalf("注册路由失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := rt.Start(ctx); err != nil {
		t.Fatalf("启动 router 失败: %v", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
		defer closeCancel()
		if err := rt.Close(closeCtx); err != nil {
			t.Fatalf("关闭 router 失败: %v", err)
		}
	}()

	event := base.Event{
		ID:       "msg-1",
		Platform: base.PlatformQQ,
		Kind:     base.EventKindMessage,
		Message: &base.Message{
			ID:     "msg-1",
			Chat:   base.Chat{ID: "chat-1", Type: base.ChatTypeGroup},
			Sender: base.User{ID: "user-1"},
		},
	}
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("发布事件失败: %v", err)
	}

	deadline := time.After(time.Second)
	for handler.count() != 1 {
		select {
		case <-deadline:
			t.Fatal("等待 handler 收到事件超时")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	stats := rt.Stats()
	if stats.Received != 1 || stats.Published != 1 {
		t.Fatalf("router 统计错误: %+v", stats)
	}
	routeStats := stats.RouteSummaries["qq-group-message"]
	if routeStats.Handled != 1 || routeStats.Matched != 1 {
		t.Fatalf("route 统计错误: %+v", routeStats)
	}
}

func TestRouterDropNewestOverflow(t *testing.T) {
	bus := base.NewMemoryBus()
	rt, err := New(bus, Options{BrokerBuffer: 16})
	if err != nil {
		t.Fatalf("创建 router 失败: %v", err)
	}

	wait := make(chan struct{})
	handler := &captureHandler{wait: wait}
	if err := rt.Register(Route{
		Name:           "slow-route",
		Filter:         base.EventFilter{Kind: base.EventKindMessage},
		Handler:        handler,
		QueueSize:      1,
		Workers:        1,
		OverflowPolicy: OverflowDropNewest,
	}); err != nil {
		t.Fatalf("注册路由失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := rt.Start(ctx); err != nil {
		t.Fatalf("启动 router 失败: %v", err)
	}

	makeEvent := func(id string) base.Event {
		return base.Event{
			ID:       id,
			Platform: base.PlatformQQ,
			Kind:     base.EventKindMessage,
			Message: &base.Message{
				ID:     id,
				Chat:   base.Chat{ID: "chat-1", Type: base.ChatTypeGroup},
				Sender: base.User{ID: "user-1"},
			},
		}
	}

	if err := bus.Publish(context.Background(), makeEvent("1")); err != nil {
		t.Fatalf("发布事件 1 失败: %v", err)
	}
	if err := bus.Publish(context.Background(), makeEvent("2")); err != nil {
		t.Fatalf("发布事件 2 失败: %v", err)
	}
	if err := bus.Publish(context.Background(), makeEvent("3")); err != nil {
		t.Fatalf("发布事件 3 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	close(wait)

	closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
	defer closeCancel()
	if err := rt.Close(closeCtx); err != nil {
		t.Fatalf("关闭 router 失败: %v", err)
	}

	stats := rt.Stats().RouteSummaries["slow-route"]
	if stats.Dropped == 0 {
		t.Fatalf("预期至少丢弃 1 条消息，got=%+v", stats)
	}
}

func TestRouterRejectsInvalidRoute(t *testing.T) {
	bus := base.NewMemoryBus()
	rt, err := New(bus, Options{})
	if err != nil {
		t.Fatalf("创建 router 失败: %v", err)
	}

	err = rt.Register(Route{})
	if !errors.Is(err, ErrRouteNameRequired) {
		t.Fatalf("预期 ErrRouteNameRequired，got=%v", err)
	}
}
