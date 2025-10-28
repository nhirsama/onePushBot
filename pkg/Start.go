package pkg

import (
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
}
func Start(apiWsm *api.WebSocketMessage) {
	w := &WebSocketMessage{WebSocketMessage: apiWsm}
	//w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
	//	w.reply(msg)
	//}, false)

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.autoSetMsgEmojiLike(msg)
	}, false)

	w.Bus.SubscribeAsync("atMe", func(msg *api.MessageStruct) {
		w.replyGroup(msg)
	}, false)

	w.Bus.SubscribeAsync("poke", func(msg *api.MessageStruct) {
		w.replyPoke(msg)
	}, false)
	//回调sudo函数
	//w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
	//	w.sudo(msg)
	//}, false)

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.sudo(msg)
	}, false)

	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		strings.Contains(string(msg.Message), "\"type\":\"text\"")
		config.DB.UpdateUser(msg.UserId, msg.RawMessage)
	}, false)
}
