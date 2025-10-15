package api

import (
	"log"
)

func (w *WebSocketMessage) SendGroupMsg(groupId int64, text string, _type string) {
	type messageData struct {
		Text string `json:"text"`
	}

	type message struct {
		Type string      `json:"type"`
		Data messageData `json:"data"`
	}

	type params struct {
		GroupId int64     `json:"group_id"`
		Message []message `json:"message"`
	}

	resp, err := w.callAPI("send_group_msg", params{
		GroupId: groupId,
		Message: []message{
			{Type: _type, Data: messageData{Text: text}},
		},
	})
	if err != nil {
		log.Println("callAPI send_group_msg error:", err)
		return
	}

	if resp.Status == "ok" {
		log.Println("请求 send_group_msg api 成功")
	}
}
