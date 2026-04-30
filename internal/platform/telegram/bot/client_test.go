package bot

import (
	"context"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	base "github.com/nhirsama/onePushBot/internal/platform"
)

func TestStartRejectsClosedClient(t *testing.T) {
	client := &client{closed: true}

	if err := client.Start(context.Background()); err != errBotClosed {
		t.Fatalf("关闭后的客户端应拒绝启动，got=%v", err)
	}
}

func TestPublishUpdateHandlesInlineCallbackQuery(t *testing.T) {
	client := &client{
		eventCh: make(chan base.Event, 1),
	}
	client.accepting.Store(true)

	client.publishUpdate(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback-1",
			From: &tgbotapi.User{ID: 42, FirstName: "Alice"},
		},
	})

	select {
	case event := <-client.eventCh:
		if event.ID != "callback-1" {
			t.Fatalf("回调事件 ID 错误: %s", event.ID)
		}
		if event.Time.IsZero() {
			t.Fatal("回调事件时间不应为空")
		}
		if event.Notice == nil {
			t.Fatal("回调事件缺少 notice")
		}
		if event.Notice.Chat.Type != base.ChatTypeUnknown {
			t.Fatalf("无关联消息的回调应降级为 unknown chat，got=%s", event.Notice.Chat.Type)
		}
		if event.Notice.User.ID != "42" {
			t.Fatalf("回调用户 ID 错误: %s", event.Notice.User.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("等待回调事件超时")
	}
}

func TestPublishUpdateHandlesMessageWithoutSender(t *testing.T) {
	client := &client{
		eventCh: make(chan base.Event, 1),
	}
	client.accepting.Store(true)

	client.publishUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			MessageID: 7,
			Date:      int(time.Now().Unix()),
			Chat: &tgbotapi.Chat{
				ID:    1001,
				Type:  "channel",
				Title: "news",
			},
			Text: "hello",
		},
	})

	select {
	case event := <-client.eventCh:
		if event.Message == nil {
			t.Fatal("消息事件缺少 message")
		}
		if event.Message.Sender.ID != "" {
			t.Fatalf("无发送者消息不应伪造 sender，got=%s", event.Message.Sender.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("等待消息事件超时")
	}
}
