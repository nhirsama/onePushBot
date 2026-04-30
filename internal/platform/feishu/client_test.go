package feishu

import (
	"context"
	"testing"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

func TestExtractMessageText(t *testing.T) {
	messageType := "text"
	content := "{\"text\":\"你好\"}"
	if text := extractMessageText(&messageType, &content); text != "你好" {
		t.Fatalf("expected extracted text, got %q", text)
	}
}

func TestParseMilliTimeUsesUnixMilliseconds(t *testing.T) {
	value := "1714550400123"
	got := parseMilliTime(&value)

	if got.UnixMilli() != 1714550400123 {
		t.Fatalf("毫秒时间戳解析错误: got=%d", got.UnixMilli())
	}
}

func TestPublishDropsEventAfterClose(t *testing.T) {
	client := &client{
		eventCh: make(chan base.Event, 1),
		started: true,
	}
	client.accepting.Store(true)

	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("关闭客户端失败: %v", err)
	}

	client.publish(base.Event{ID: "late"})

	select {
	case event := <-client.eventCh:
		t.Fatalf("关闭后不应继续投递事件，got=%s", event.ID)
	default:
	}
}
