package pkg

import (
	"log"

	"github.com/nhirsama/onePushBot/api"
	"github.com/spf13/viper"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
	AutoSetMsgEmojiLikeSet map[string][]int
}

func init() {
}
func Start(apiWsm *api.WebSocketMessage) {
	w := &WebSocketMessage{WebSocketMessage: apiWsm}
	w.ReadConfig()
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

func (w *WebSocketMessage) ReadConfig() {
	err := viper.UnmarshalKey("AutoSetMsgEmojiLikeSet", &w.AutoSetMsgEmojiLikeSet)
	if err != nil {
		log.Println(err)
	}
}
