package api

import "log"

func (w *WebSocketMessage) SendLike(userId int, times int) {
	type params struct {
		UserId int `json:"user_id"`
		Times  int `json:"times"`
	}

	resp, err := w.callAPI("send_like", params{userId, times})
	if err != nil {
		log.Println("callAPI error:", err)
	}
	if resp.Status == "ok" {
		log.Println("请求 send_like api 成功")
	}
}

/*
{
  "action": "send_like",
  "params":{
    "user_id" : xxxxxx,
    "times": 10
  }
}
*/
