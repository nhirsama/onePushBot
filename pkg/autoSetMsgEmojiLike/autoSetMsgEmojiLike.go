package autoSetMsgEmojiLike

import (
	"strconv"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
)

func init() {
	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{start, "autoSetMsgEmojiLike"})
}

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func (w *WebSocketMessage) autoSetMsgEmojiLike(message *api.MessageStruct) {
	if value, ok := config.AutoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go w.SetMsgEmojiLike(message.MessageId, i, true)
		}
	}
}

func start(wsm *api.WebSocketMessage) {
	w := WebSocketMessage{wsm}

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.autoSetMsgEmojiLike(msg)
	}, false)
}
