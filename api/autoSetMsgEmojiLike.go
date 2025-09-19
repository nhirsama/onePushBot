package api

import (
	"strconv"

	"github.com/nhirsama/onePushBot/global"
)

func (w *WebSocketMessage) AutoSetMsgEmojiLike(message global.Message) {
	if value, ok := global.AutoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go w.SetMsgEmojiLike(message.MessageId, i, true)
		}
	}
}
