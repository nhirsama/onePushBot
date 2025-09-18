package act

import (
	"strconv"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/global"
)

func AutoSetMsgEmojiLike(message global.Message) {
	if value, ok := global.AutoSetMsgEmojiLikeSet[strconv.FormatInt(message.UserId, 10)]; ok {
		for _, i := range value {
			go api.Set_msg_emoji_like(message.MessageId, i, true)
		}
	}
}
