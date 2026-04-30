package platform

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHubForwardEvents(t *testing.T) {
	client := &stubClient{
		platform: PlatformQQ,
		events:   make(chan Event, 1),
	}

	hub := NewHub(nil)
	if err := hub.Register(client); err != nil {
		t.Fatalf("注册客户端失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := hub.Start(ctx); err != nil {
		t.Fatalf("启动 Hub 失败: %v", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
		defer closeCancel()
		if err := hub.Close(closeCtx); err != nil {
			t.Fatalf("关闭 Hub 失败: %v", err)
		}
	}()

	sub, err := hub.Bus().Subscribe(EventFilter{Platform: PlatformQQ, Kind: EventKindMessage})
	if err != nil {
		t.Fatalf("订阅总线失败: %v", err)
	}
	defer sub.Unsubscribe()

	want := Event{
		ID:       "hub-1",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		Message: &Message{
			Chat: Chat{ID: "chat-1"},
		},
	}
	client.events <- want

	select {
	case got := <-sub.Events():
		if got.ID != want.ID {
			t.Fatalf("事件转发错误: got=%s want=%s", got.ID, want.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("等待 Hub 转发事件超时")
	}
}

type rollbackClient struct {
	platform Platform
	events   chan Event
	startErr error
	started  int
	closed   int
}

func (c *rollbackClient) Platform() Platform {
	return c.platform
}

func (c *rollbackClient) Start(context.Context) error {
	c.started++
	return c.startErr
}

func (c *rollbackClient) Close(context.Context) error {
	c.closed++
	return nil
}

func (c *rollbackClient) Events() <-chan Event {
	return c.events
}

func (c *rollbackClient) Status() Status {
	return StatusStopped
}

type orderedRegistry struct {
	clients []Client
}

func (r *orderedRegistry) Register(Client) error {
	return nil
}

func (r *orderedRegistry) Get(platform Platform) (Client, bool) {
	for _, client := range r.clients {
		if client.Platform() == platform {
			return client, true
		}
	}
	return nil, false
}

func (r *orderedRegistry) All() []Client {
	return append([]Client(nil), r.clients...)
}

func TestHubStartRollbackClosesStartedClients(t *testing.T) {
	first := &rollbackClient{
		platform: PlatformQQ,
		events:   make(chan Event),
	}
	secondErr := errors.New("start failed")
	second := &rollbackClient{
		platform: PlatformTelegramBot,
		events:   make(chan Event),
		startErr: secondErr,
	}

	hub := &hub{
		bus:      NewMemoryBus(),
		registry: &orderedRegistry{clients: []Client{first, second}},
	}

	if err := hub.Start(context.Background()); !errors.Is(err, secondErr) {
		t.Fatalf("启动 Hub 应返回第二个客户端错误，got=%v", err)
	}
	if first.closed != 1 {
		t.Fatalf("启动失败后应回滚已启动客户端，close=%d", first.closed)
	}

	if err := hub.Start(context.Background()); !errors.Is(err, secondErr) {
		t.Fatalf("回滚后 Hub 应允许再次启动，got=%v", err)
	}
	if first.closed != 2 {
		t.Fatalf("二次启动失败后应再次回滚，close=%d", first.closed)
	}
}
