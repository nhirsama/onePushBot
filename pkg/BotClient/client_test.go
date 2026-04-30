package BotClient

import (
	"context"
	"testing"

	"github.com/nhirsama/onePushBot/internal/domain"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	bot, err := NewClient(ctx, &domain.BotConfig{
		Type:                 "napcat",
		ApiUrl:               "127.0.0.1:3001",
		Token:                "",
		HeartbeatTimeout:     50,
		ReconnectMaxAttempts: 5,
	})
	if err != nil {
		t.Fatalf("NewClient() 返回错误: %v", err)
	}
	if bot == nil {
		t.Fatal("NewClient() 返回了空客户端")
	}
}
