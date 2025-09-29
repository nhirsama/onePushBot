package api

import (
	"log"
)

func (w *WebSocketMessage) SetMsgEmojiLike(messageId int64, emojiId int, set bool) {
	type params struct {
		MessageId int64 `json:"message_id"`
		EmojiId   int   `json:"emoji_id"`
		Set       bool  `json:"set"`
	}

	resp, err := w.callAPI("set_msg_emoji_like", params{messageId, emojiId, set})
	if err != nil {
		log.Println("callAPI error:", err)
		return
	}
	if resp.Status == "ok" {
		log.Println("请求 set_msg_emoji_like api 成功")
	}
}
