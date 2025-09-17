package pkg

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/global"
)

// BaseMsg 基础消息，只检测 post_type
type BaseMsg struct {
	PostType string `json:"post_type"`
}

// 消息事件
type MessageEvent struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	UserID      int64  `json:"user_id"`
	Message     string `json:"message"`
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
		type response struct {
			Echo string `json:"echo"`
		}
		var reva response
		if err := json.Unmarshal(msg, &reva); err != nil {
			log.Println("解析返回值失败:", err)
			return
		}
		if ch, ok := global.ResponseMap.Load(reva.Echo); ok {
			ch.(chan []byte) <- msg
		} else {
			log.Printf("接收到未追踪的返回值：%s", reva.Echo)
		}
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
		if message.User_id == 0 || message.User_id == 1 {
			go api.Set_msg_emoji_like(message.Message_id, 147, true)
			//可以给特定人的消息回复爱心表情
		} else if message.User_id == 0 || message.User_id == 1 {
			go api.Set_msg_emoji_like(message.Message_id, 66, true)
		}
	}
}
