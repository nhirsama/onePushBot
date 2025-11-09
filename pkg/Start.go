package pkg

import (
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	_ "github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
	_ "github.com/nhirsama/onePushBot/pkg/autoSetMsgEmojiLike"
	_ "github.com/nhirsama/onePushBot/pkg/reply"
	_ "github.com/nhirsama/onePushBot/pkg/riddle"
	"github.com/spf13/viper"
)

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{TaskFunc: func(w *api.WebSocketMessage) {
		w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
			if strings.Contains(string(msg.Message), "\"type\":\"text\"") {
				config.DB.UpdateUser(msg.UserId, msg.RawMessage)
			}
		}, false)
	}, ModuleName: "group_message_token_DB"})
}
func Start(apiWsm *api.WebSocketMessage) {
	for _, funcHandle := range TaskFunc.ModuleList {
		if viper.GetBool(funcHandle.ModuleName + ".Enable") {
			go funcHandle.TaskFunc(apiWsm)
		} else {
			viper.Set(funcHandle.ModuleName+".Enable", false)
		}
	}
}
