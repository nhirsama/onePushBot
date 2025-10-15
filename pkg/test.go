package pkg

import (
	"log"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/pkg/llm"
)

func (w *WebSocketMessage) reply(message *api.MessageStruct) {
	data := llm.Call(string(message.Message))
	log.Println(data)
	w.SendPrivateMsg(message.UserId, data, "text")
}

func (w *WebSocketMessage) replyGroup(message *api.MessageStruct) {
	data := llm.Call(string(message.Message))
	log.Println(data)
	w.SendGroupMsg(message.GroupId, data, "text")
}
