package api

import (
	"encoding/json"
	"log"
)

type metaEvent struct {
	Time          int64           `json:"time"`
	SelfId        int64           `json:"self_id"`
	MetaEventType string          `json:"meta_event_type"`
	Interval      json.RawMessage `json:"interval"`
	Status        json.RawMessage `json:"status"`
	SubType       json.RawMessage `json:"sub_type"`
}

func (w *WebSocketMessage) metaEvent(msg []byte) {
	var mev metaEvent
	err := json.Unmarshal(msg, &mev)
	if err != nil {
		log.Println("元信息解析失败：", string(msg))
		return
	}
	if mev.MetaEventType == "heartbeat" {
		w.heartbeat <- struct{}{}
	} else {
		log.Println("收到未定义的元信息：", string(msg))
	}
}
