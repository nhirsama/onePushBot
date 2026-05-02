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
			ID:     "1",
			Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
			Sender: User{ID: "user-1"},
			Text:   "hello",
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
			Type: "other",
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
			Type: "poke",
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
	if stats := bus.Stats(); stats.ActiveSubscriptions != 1 {
		t.Fatalf("订阅后活跃订阅数错误: got=%d want=1", stats.ActiveSubscriptions)
	}

	sub.Unsubscribe()
	sub.Unsubscribe()
	if stats := bus.Stats(); stats.ActiveSubscriptions != 0 {
		t.Fatalf("取消订阅后活跃订阅数错误: got=%d want=0", stats.ActiveSubscriptions)
	}

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

func TestMemoryBusRejectInvalidEvent(t *testing.T) {
	bus := NewMemoryBus()

	err := bus.Publish(context.Background(), Event{
		ID:       "broken",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		Message: &Message{
			ID:     "broken",
			Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
			Sender: User{},
		},
	})
	if err == nil {
		t.Fatal("发布非法事件应返回错误")
	}

	stats := bus.Stats()
	if stats.Invalid != 1 {
		t.Fatalf("非法事件计数错误: got=%d want=1", stats.Invalid)
	}
	if stats.Published != 0 {
		t.Fatalf("非法事件不应计入已发布: got=%d", stats.Published)
	}
}

func TestMemoryBusTracksDroppedDeliveries(t *testing.T) {
	bus := NewMemoryBus(WithSubscriptionBuffer(1))
	sub, err := bus.Subscribe(EventFilter{Kind: EventKindMessage})
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}
	defer sub.Unsubscribe()

	first := Event{
		ID:       "first",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		Message: &Message{
			ID:     "first",
			Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
			Sender: User{ID: "user-1"},
		},
	}
	second := Event{
		ID:       "second",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		Message: &Message{
			ID:     "second",
			Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
			Sender: User{ID: "user-1"},
		},
	}

	if err := bus.Publish(context.Background(), first); err != nil {
		t.Fatalf("发布第一条消息失败: %v", err)
	}
	if err := bus.Publish(context.Background(), second); err != nil {
		t.Fatalf("发布第二条消息失败: %v", err)
	}

	stats := bus.Stats()
	if stats.Published != 2 {
		t.Fatalf("发布计数错误: got=%d want=2", stats.Published)
	}
	if stats.Delivered != 1 {
		t.Fatalf("投递计数错误: got=%d want=1", stats.Delivered)
	}
	if stats.Dropped != 1 {
		t.Fatalf("丢弃计数错误: got=%d want=1", stats.Dropped)
	}
}
