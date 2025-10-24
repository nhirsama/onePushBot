package pkg

import (
	"github.com/nhirsama/onePushBot/api"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
}
func Start(apiWsm *api.WebSocketMessage) {
	w := &WebSocketMessage{WebSocketMessage: apiWsm}
	w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
		w.reply(msg)
	}, false)

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.autoSetMsgEmojiLike(msg)
	}, false)

	w.Bus.SubscribeAsync("atMe", func(msg *api.MessageStruct) {
		w.replyGroup(msg)
	}, false)

	w.Bus.SubscribeAsync("poke", func(msg *api.MessageStruct) {
		w.replyPoke(msg)
	}, false)
}
