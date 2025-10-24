package cmd

import (
	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg"
)

func Cli() {
	// 连接 api WebSocket
	w := api.NewWebSocketMessage(config.WebSocketUrl)
	defer w.Close()
	go w.Start()
	go pkg.Start(w)

	select {}
}
