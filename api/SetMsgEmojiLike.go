package api

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/nhirsama/onePushBot/global"
)

func (w *WebSocketMessage) SetMsgEmojiLike(message_id int64, emoji_id int, set bool) {
	type params struct {
		Message_id int64 `json:"message_id"`
		Emoji_id   int   `json:"emoji_id"`
		Set        bool  `json:"set"`
	}
	type setMsgEmojiLike struct {
		Action string `json:"action"`
		Echo   string `json:"echo"`
		Parmes params `json:"params"`
	}

	for retryCount := 1; retryCount < 4; retryCount++ {
		if func() bool {
			echo := strconv.FormatInt(time.Now().UnixNano(), 10)
			ch := make(chan []byte, 1)
			global.ResponseMap.Store(echo, ch)
			defer global.ResponseMap.Delete(echo)

			body := setMsgEmojiLike{
				"set_msg_emoji_like",
				echo,
				params{
					message_id,
					emoji_id,
					set,
				}}
			err := w.conn.WriteJSON(body)
			if err != nil {
				log.Println(err)
			}

			select {
			case msg := <-ch:
				type data struct {
					Result int    `json:"result"`
					ErrMsg string `json:"errMsg"`
				}
				type response struct {
					Status  string `json:"status"`
					RetCode int    `json:"retcode"`
					Echo    string `json:"echo"`
					Data    data   `json:"data"`
				}
				resp := response{}
				err = json.Unmarshal(msg, &resp)
				if err != nil {
					log.Println(err)
				}

				if resp.Status == "ok" {
					log.Println("请求 set_msg_emoji_like api 成功")
					return true
				} else {
					jsonBytes, _ := json.Marshal(body)
					log.Printf("请求 set_msg_emoji_like api 错误，错误码:%d，重试次数:%d。请求体:%s \n", resp.RetCode, retryCount, jsonBytes)
				}
			case <-time.After(time.Second * 20):
				jsonBytes, _ := json.Marshal(body)
				log.Printf("请求 set_msg_emoji_like api 超时，重试次数 %d。请求体：%s \n", retryCount, jsonBytes)
			}
			return false
		}() {
			break
		}
	}
}
