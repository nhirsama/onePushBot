package pkg

import (
	"github.com/nhirsama/onePushBot/api"
	_ "github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
	_ "github.com/nhirsama/onePushBot/pkg/autoSetMsgEmojiLike"
	_ "github.com/nhirsama/onePushBot/pkg/riddle"
	"github.com/spf13/viper"
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

	for _, funcHandle := range TaskFunc.ModuleList {
		if viper.GetBool(funcHandle.ModuleName + "Enable") {
			go funcHandle.TaskFunc(apiWsm)
		}
	}
}
