package llm

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func Call(inquiry string) string {
	apiToken := strings.TrimSpace(viper.GetString("llm.api_token"))
	if apiToken == "" {
		log.Println("LLM API token 未配置")
		return ""
	}
	client := NewClient(apiToken, viper.GetString("llm.base_url"))

	req := &ChatRequest{
		Model: viper.GetString("llm.model"),
		Messages: []ChatMessage{
			{Role: "system", Content: "你是一名资深的计算机科学讲师，你叫千代，擅长讲解算法、系统原理、数据结构与底层机制。你是一个聊天机器人，请尽量使用日常聊天的风格，在符合对话风格的情况下尽量简短。"},
			{Role: "user", Content: inquiry},
		},
		Temperature: 0.6,
		Stream:      false,
	}

	resp, err := client.Chat(req)
	if err != nil {
		log.Println("API 调用失败:", err)
		return ""
	}

	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content)
	} else {
		return ""
	}
}
