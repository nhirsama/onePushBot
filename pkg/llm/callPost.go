package llm

import (
	"log"

	"github.com/spf13/viper"
)

func Call(inquiry string) string {
	var apiKey string
	err := viper.UnmarshalKey("apiKey", &apiKey)
	if err != nil {
		return ""
	}

	client := NewClient(apiKey) // 你的 Bearer Token

	req := &ChatRequest{
		Model: "glm-4.5-flash",
		Messages: []ChatMessage{
			{Role: "system", Content: "你是一名资深的计算机科学讲师，你叫千代，擅长讲解算法、系统原理、数据结构与底层机制。\n\n我希望你在回答问题时：\n\n重点解释思路、算法、原理，不要直接给出完整可运行的代码。\n\n可以使用 C++ 伪代码 或部分实现片段，展示关键逻辑。\n\n对于复杂问题，请描述性能分析（时间/空间复杂度）和设计考量。\n\n所有回答都应体现出严谨的计算机专业表达风格。\n\n例如，如果我问“如何检测图中是否存在环”，你不应直接给出完整的 DFS 代码，而应：\n\n说明图的表示方式（邻接表/邻接矩阵）；\n\n讲解 DFS 检测环的原理；\n\n用几行 C++ 片段展示关键思想；\n\n最后给出复杂度分析。并且你是一个聊天机器人，请尽量使用日常聊天的风格。"},
			{Role: "user", Content: inquiry},
		},
		Temperature: 0.6,
		Stream:      false,
	}

	resp, err := client.Chat(req)
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content
	} else {
		return ""
	}
}
