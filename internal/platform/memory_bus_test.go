package platform

import (
	"context"
	"testing"
	"time"
)

func TestMemoryBusPublishSubscribe(t *testing.T) {
	bus := NewMemoryBus()
	sub, err := bus.Subscribe(EventFilter{Kind: EventKindMessage})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	event := Event{
		ID:       "1",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		Message: &Message{
			Chat:   Chat{ID: "chat-1"},
			Sender: User{ID: "user-1"},
		},
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("发布失败: %v", err)
	}

	select {
	case got := <-sub.Events():
		if got.ID != event.ID {
			t.Fatalf("事件不匹配: got=%s want=%s", got.ID, event.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("等待事件超时")
	}
}

func TestMemoryBusFilter(t *testing.T) {
	bus := NewMemoryBus()
	sub, err := bus.Subscribe(EventFilter{
		Platform: PlatformQQ,
		Kind:     EventKindNotice,
		SubType:  "poke",
		ChatID:   "group-1",
		UserID:   "user-1",
	})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	_ = bus.Publish(context.Background(), Event{
		ID:       "mismatch",
		Platform: PlatformQQ,
		Kind:     EventKindNotice,
		SubType:  "other",
		Notice: &Notice{
			Chat: Chat{ID: "group-1"},
			User: User{ID: "user-1"},
		},
	})

	select {
	case <-sub.Events():
		t.Fatal("不应该收到不匹配的事件")
	case <-time.After(100 * time.Millisecond):
	}

	match := Event{
		ID:       "match",
		Platform: PlatformQQ,
		Kind:     EventKindNotice,
		SubType:  "poke",
		Notice: &Notice{
			Chat: Chat{ID: "group-1"},
			User: User{ID: "user-1"},
		},
	}
	if err := bus.Publish(context.Background(), match); err != nil {
		t.Fatalf("发布失败: %v", err)
	}

	select {
	case got := <-sub.Events():
		if got.ID != match.ID {
			t.Fatalf("事件不匹配: got=%s want=%s", got.ID, match.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("等待事件超时")
	}
}

func TestMemoryBusUnsubscribeAndClose(t *testing.T) {
	bus := NewMemoryBus()
	sub, err := bus.Subscribe(EventFilter{})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	sub.Unsubscribe()
	sub.Unsubscribe()

	select {
	case _, ok := <-sub.Events():
		if ok {
			t.Fatal("取消订阅后通道应关闭")
		}
	case <-time.After(time.Second):
		t.Fatal("等待取消订阅关闭超时")
	}

	if err := bus.Close(); err != nil {
		t.Fatalf("关闭总线失败: %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatalf("重复关闭总线失败: %v", err)
	}
}
