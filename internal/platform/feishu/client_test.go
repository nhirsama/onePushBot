package feishu

import "testing"

func TestExtractMessageText(t *testing.T) {
	messageType := "text"
	content := "{\"text\":\"你好\"}"
	if text := extractMessageText(&messageType, &content); text != "你好" {
		t.Fatalf("expected extracted text, got %q", text)
	}
}
