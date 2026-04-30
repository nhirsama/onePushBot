package platform

import (
	"context"
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
