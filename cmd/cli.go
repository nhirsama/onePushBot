package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/pkg"
)

func Cli() {
	// 连接 api WebSocket
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	w := api.NewWebSocketMessage(config.WebSocketUrl)
	defer w.Close()
	go w.Start()
	go pkg.Start(w)
	// 阻塞程序结束，等待终止信号
	<-sig
	config.SaveConfig()
	log.Println("程序正常结束，正在保存与释放资源")
	return
}
