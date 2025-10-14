package pkg

import (
	"github.com/nhirsama/onePushBot/api"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func (w *WebSocketMessage) test(message *api.MessageStruct) {
	w.SendPrivateMsg(message.UserId, "test", "text")
}
func (w *WebSocketMessage) TestFunc() {
	w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
		w.test(msg)
	}, false)

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.AutoSetMsgEmojiLike(msg)
	}, false)
}
