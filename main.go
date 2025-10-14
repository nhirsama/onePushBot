package main

import (
	"net/url"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/pkg"
	"github.com/spf13/viper"
)

func main() {
	// 链接 api WebSocket
	webSocketUrl := url.URL{Scheme: "wss", Host: viper.GetString("apiUrl"), Path: "/ws", RawQuery: "access_token=" + viper.GetString("token")}
	w := api.NewWebSocketMessage(webSocketUrl.String())
	defer w.Close()
	go w.Start()
	pkgW := pkg.WebSocketMessage{WebSocketMessage: w}
	go pkgW.TestFunc()
	select {}
}
