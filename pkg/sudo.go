package pkg

import (
	"log"
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/llm"
)

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
