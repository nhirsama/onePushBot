package api

import (
	"strconv"
)

var autoSetMsgEmojiLikeSet map[string][]int

func (w *WebSocketMessage) AutoSetMsgEmojiLike(message messageStruct) {
	if value, ok := autoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go w.SetMsgEmojiLike(message.MessageId, i, true)
		}
	}
}
