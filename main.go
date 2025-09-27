package main

import (
	"net/url"

	"github.com/nhirsama/onePushBot/api"
	_ "github.com/nhirsama/onePushBot/global"
	"github.com/spf13/viper"
)

func main() {
	// 链接 api WebSocket
	webSocketUrl := url.URL{Scheme: "wss", Host: viper.GetString("apiUrl"), Path: "/ws", RawQuery: "access_token=" + viper.GetString("token")}
	w := api.NewWebSocketMessage(webSocketUrl.String())
	defer w.Close()
	w.Start()
	select {}
}
