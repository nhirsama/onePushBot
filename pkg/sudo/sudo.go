package sudo

import (
	"log"
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
	"github.com/nhirsama/onePushBot/pkg/llm"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{start, "sudo"})
}
func (w *WebSocketMessage) sudo(message *api.MessageStruct) {
	if strings.Contains(message.RawMessage, "-----BEGIN PGP SIGNED MESSAGE-----") {
		mes, ok := config.Auth.Authenticate([]byte(message.RawMessage))
		if ok {
			switch message.MessageType {
			case "group":
				w.sudoReplyGroup(mes, message.GroupId)
				//case "private":
				//	w.reply(message)
			}
		}
	}

}

func (w *WebSocketMessage) sudoReplyGroup(mes string, groupId int64) {
	data := llm.Call(mes)
	log.Println(data)
	w.SendGroupMsg(groupId, data, "text")
}

func start(wsm *api.WebSocketMessage) {
	w := WebSocketMessage{WebSocketMessage: wsm}
	//回调sudo函数
	//w.Bus.SubscribeAsync("privateMessage", func(msg *api.MessageStruct) {
	//	w.sudo(msg)
	//}, false)
	w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		w.sudo(msg)
	}, false)
}
