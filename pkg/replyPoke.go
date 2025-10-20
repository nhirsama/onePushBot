package pkg

import (
	"github.com/nhirsama/onePushBot/api"
	"github.com/spf13/viper"
)

func (w *WebSocketMessage) replyPoke(message *api.MessageStruct) {
	var selfId int64
	viper.UnmarshalKey("selfId", &selfId)
	if message.TargetId == selfId {
		w.SendPoke(message.GroupId, message.UserId)
	}
}
