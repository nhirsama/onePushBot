package api

import (
	"encoding/json"
	"log"
)

type messageStruct struct {
	Time          int64           `json:"time"`
	SelfId        int64           `json:"self_id"`
	PostType      string          `json:"post_type"`
	MetaEventType string          `json:"meta_event_type"`
	Interval      int64           `json:"interval"`
	MessageId     int64           `json:"message_id"`
	MessageSeq    json.RawMessage `json:"message_seq"`
	RealId        json.RawMessage `json:"real_id"`
	RealSeq       json.RawMessage `json:"real_seq"`
	MessageType   string          `json:"message_type"`
	Sender        json.RawMessage `json:"sender"`
	RawMessage    json.RawMessage `json:"raw_message"`
	Font          json.RawMessage `json:"font"`
	SubType       json.RawMessage `json:"sub_type"`
	Message       json.RawMessage `json:"message"`
	MessageFormat string          `json:"message_format"`
	GroupId       int64           `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserId        int64           `json:"user_id"`
	Status        json.RawMessage `json:"status"`
	RetCode       int64           `json:"retcode"`
	Data          struct {
		Result int64  `json:"result"`
		ErrMsg string `json:"errMsg"`
	}
	Wording string `json:"wording"`
	Echo    string `json:"echo"`
}

func (w *WebSocketMessage) handleMessage(msg []byte) {
	var messStruct messageStruct
	if err := json.Unmarshal(msg, &messStruct); err != nil {
		log.Println("json解析失败：", string(msg))
	}

	switch messStruct.PostType {
	case "message":
		w.handleMessageMessage(messStruct)
		log.Println(string(msg))
	case "meta_event":
		w.metaEvent(msg)
	default:
		switch messStruct.Echo {
		case "":
			log.Printf("接收到未定义的信息:%s\n", string(msg))
		default:
			if ch, ok := w.responseMap.Load(messStruct.Echo); ok {
				ch.(chan []byte) <- msg
			} else {
				log.Printf("接收到未追踪的返回值：%s", msg)
			}
		}

	}
}

func (w *WebSocketMessage) handleMessageMessage(ms messageStruct) {
	switch ms.MessageType {
	case "group":
		//TODO: 应该通过 eventBus 广播到所有功能中
		w.AutoSetMsgEmojiLike(ms)
	}
}
