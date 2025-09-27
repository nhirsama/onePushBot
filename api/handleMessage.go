package api

import (
	"encoding/json"
	"errors"
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
	MessageType   json.RawMessage `json:"message_type"`
	Sender        json.RawMessage `json:"sender"`
	RawMessage    json.RawMessage `json:"raw_message"`
	Font          json.RawMessage `json:"font"`
	SubType       json.RawMessage `json:"sub_type"`
	Message       json.RawMessage `json:"message"`
	MessageFormat string          `json:"message_format"`
	GroupId       int64           `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserId        int64           `json:"user_id"`
}

type apiResponse struct {
	Status  string `json:"status"`
	RetCode int64  `json:"retcode"`
	Data    struct {
		Result int64  `json:"result"`
		ErrMsg string `json:"errMsg"`
	}
	Message string `json:"message"`
	Wording string `json:"wording"`
	Echo    string `json:"echo"`
}

func (w *WebSocketMessage) handleMessage(msg []byte) {
	var messStruct messageStruct
	if err := json.Unmarshal(msg, &messStruct); err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			var apiResponse apiResponse
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
		w.handleMessageMessage(messStruct)
		log.Println(string(msg))
	case "meta_event":
		w.metaEvent(msg)
	default:
		log.Printf("接收到未定义的信息:%s\n", string(msg))
	}
}

func (w *WebSocketMessage) handleMessageMessage(ms messageStruct) {
	var messageType string
	if err := json.Unmarshal(ms.MessageType, &messageType); err != nil {
		log.Println(err)
	}
	switch messageType {
	case "group":
		//TODO: 应该通过 eventBus 广播到所有功能中
		w.AutoSetMsgEmojiLike(ms)
	}
}
