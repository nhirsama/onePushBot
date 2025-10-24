package pkg

import (
	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
)

func (w *WebSocketMessage) replyPoke(message *api.MessageStruct) {
	if message.TargetId == config.SelfId {
		w.SendPoke(message.GroupId, message.UserId)
	}
}
