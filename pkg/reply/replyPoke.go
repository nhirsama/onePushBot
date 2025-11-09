package reply

import (
	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
)

func (w *WebSocketMessage) replyPoke(message *api.MessageStruct) {
	if message.TargetId == config.SelfId {
		w.SendPoke(message.GroupId, message.UserId)
	}
}

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{start, "reply"})
}
func start(wsm *api.WebSocketMessage) {
	w := &WebSocketMessage{wsm}

	//w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
	//	w.reply(msg)
	//}, false)
	w.Bus.SubscribeAsync("atMe", func(msg *api.MessageStruct) {
		w.replyGroup(msg)
	}, false)

	w.Bus.SubscribeAsync("poke", func(msg *api.MessageStruct) {
		w.replyPoke(msg)
	}, false)
}
