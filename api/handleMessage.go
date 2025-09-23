package api

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/nhirsama/onePushBot/global"
)

// Deprecated: 请使用global的统一解析结构体
// BaseMsg 基础消息，只检测 post_type
type BaseMsg struct {
	PostType string `json:"post_type"`
}

// Deprecated: 请使用global的统一解析结构体
// MessageEvent 消息事件
type MessageEvent struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	UserID      int64  `json:"user_id"`
	Message     string `json:"message"`
}

func (w *WebSocketMessage) handleMessage(msg []byte) {
	var messStruct global.Message
	if err := json.Unmarshal(msg, &messStruct); err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			var apiResponse global.ApiResponse
			if err := json.Unmarshal(msg, &apiResponse); err != nil {
				log.Printf("不是我草这接收的json怎么连个共同字段都没有 %s", msg)
			}
			if ch, ok := ResponseMap.Load(apiResponse.Echo); ok {
				ch.(chan []byte) <- msg
			} else {
				log.Printf("接收到未追踪的返回值：%s", msg)
			}
		}
		return
	}

	switch messStruct.PostType {
	case "message":
		messageParse(messStruct, w)
		log.Println(string(msg))
	case "meta_event":
		w.metaEvent(msg)
	default:
		log.Printf("接收到未定义的信息:%s\n", string(msg))
	}
}

// Deprecated: 建议放置到pkg中，作为功能包调用。
func messageParse(message global.Message, w *WebSocketMessage) {
	if message.MessageType == "group" {
		w.AutoSetMsgEmojiLike(message)
	}
}
