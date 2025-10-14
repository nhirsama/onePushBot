package api

import (
	"log"
)

func (w *WebSocketMessage) SendPrivateMsg(userId int64, text string, _type string) {
	type messageData struct {
		Text string `json:"text"`
	}

	type message struct {
		Type string      `json:"type"`
		Data messageData `json:"data"`
	}

	type params struct {
		UserId  int64     `json:"user_id"`
		Message []message `json:"message"`
	}

	resp, err := w.callAPI("send_private_msg", params{userId, []message{{_type, messageData{text}}}})
	if err != nil {
		log.Println("callAPI send_private_msg error:", err)
	}
	if resp.Status == "ok" {
		log.Println("请求 send_private_msg api 成功")
	}
}
