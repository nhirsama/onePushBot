package pkg

import (
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
)

func (w *WebSocketMessage) sudo(message *api.MessageStruct) {
	if strings.Contains(message.RawMessage, "-----BEGIN PGP SIGNED MESSAGE-----") {
		_, ok := config.Auth.Authenticate([]byte(message.RawMessage))
		if ok {
			switch message.MessageType {
			case "group":
				w.replyGroup(message)
			case "private":
				w.reply(message)
			}
		}
	}

}
