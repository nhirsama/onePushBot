package pkg

import (
	"strconv"

	"github.com/nhirsama/onePushBot/api"
)

func (w *WebSocketMessage) AutoSetMsgEmojiLike(message *api.MessageStruct) {
	if value, ok := api.AutoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go w.SetMsgEmojiLike(message.MessageId, i, true)
		}
	}
}
