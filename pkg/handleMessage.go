package pkg

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// 基础消息（只关心公共字段）
type BaseMsg struct {
	PostType string          `json:"post_type"`
	Raw      json.RawMessage `json:"-"` // 保存原始 JSON，后面二次解析用
}

// 消息事件
type MessageEvent struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	UserID      int64  `json:"user_id"`
	Message     string `json:"message"`
}

// 请求响应（没有 post_type，而是有 status/echo）
type Response struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Echo    string          `json:"echo"`
	Data    json.RawMessage `json:"data"`
}

func HandleMessage(msg []byte, c *websocket.Conn) {
	// 第一步：解析公共字段
	var base BaseMsg
	if err := json.Unmarshal(msg, &base); err != nil {
		log.Println("解析基础消息失败:", err)
		return
	}

	// 第二步：根据 post_type 分派
	switch base.PostType {
	case "message":
		messageParse(msg, c)
		log.Println(string(msg))
	case "meta_event":
		log.Println("收到元事件:", string(msg))
	default:
		log.Println("未知消息:", string(msg))
	}
}

func messageParse(msg []byte, c *websocket.Conn) {
	type messagePreParse struct {
		Message_type string `json:"message_type"`
		Sender       string `json:"sender"`
		Sub_type     string `json:"sub_type"`
		User_id      int64  `json:"user_id"`
		Message_id   int64  `json:"message_id"`
	}
	var message messagePreParse
	if err := json.Unmarshal(msg, &message); err != nil {
		//可以给特定人的消息回复棒棒糖表情
		if message.User_id == 0 {
			err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\n  \"action\": \"set_msg_emoji_like\",\n  \"params\":{\n    \"emoji_id\": 147,\n    \"set\": true,\n    \"message_id\":%d\n  }\n}", message.Message_id)))
			if err != nil {
				log.Println("发送失败:", err)
			}

			//可以给特定人的消息回复爱心表情
		} else if message.User_id == 0 {
			err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\n  \"action\": \"set_msg_emoji_like\",\n  \"params\":{\n    \"emoji_id\": 66,\n    \"set\": true,\n    \"message_id\":%d\n  }\n}", message.Message_id)))
			if err != nil {
				log.Println("发送失败:", err)
			}
		}
	}
}
