package config

import (
	"log"
	"net/url"

	"github.com/spf13/viper"
)

var AutoSetMsgEmojiLikeSet map[string][]int
var SelfId int64
var ApiKey string

var WebSocketUrl string

func ReadConfig() {
	webSocketUrl := url.URL{Scheme: "wss", Host: viper.GetString("apiUrl"), Path: "/ws", RawQuery: "access_token=" + viper.GetString("token")}
	WebSocketUrl = webSocketUrl.String()
	err := viper.UnmarshalKey("AutoSetMsgEmojiLikeSet", &AutoSetMsgEmojiLikeSet)
	if err != nil {
		log.Println(err)
	}

	err = viper.UnmarshalKey("selfId", &SelfId)
	if err != nil {
		log.Println(err)
	}

	err = viper.UnmarshalKey("apiKey", &ApiKey)
	if err != nil {
		log.Println(err)
	}
}
