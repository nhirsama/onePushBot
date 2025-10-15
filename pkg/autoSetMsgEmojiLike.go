package pkg

import (
	"strconv"

	"github.com/nhirsama/onePushBot/api"
)

func (w *WebSocketMessage) autoSetMsgEmojiLike(message *api.MessageStruct) {
	if value, ok := w.AutoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go w.SetMsgEmojiLike(message.MessageId, i, true)
		}
	}
}
