package api

import (
	"log"

	"github.com/spf13/viper"
)

func (w *WebSocketMessage) SendPoke(groupId, targetId int64) {
	var selfId int64
	viper.UnmarshalKey("selfId", &selfId)
	type params struct {
		UserId   int64 `json:"user_id"`
		GroupId  int64 `json:"group_id"`
		TargetId int64 `json:"target_id"`
	}

	resp, err := w.callAPI("send_poke", params{
		UserId:   selfId,
		GroupId:  groupId,
		TargetId: targetId,
	})
	if err != nil {
		log.Println("callAPI send_poke error:", err)
		return
	}

	if resp.Status == "ok" {
		log.Println("请求 send_poke api 成功")
	}
}
