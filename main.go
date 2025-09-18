package main

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/global"
	"github.com/nhirsama/onePushBot/pkg"
	"github.com/spf13/viper"
)

func main() {
	const configFile = "./data/config.yaml"
	viper.SetConfigName(configFile)
	// 构造 WebSocket URL
	u := url.URL{Scheme: "wss", Host: viper.GetString("apiUrl"), Path: "/ws", RawQuery: "access_token=" + viper.GetString("token")}
	log.Printf("Connecting to %s", u.String())

	// 建立连接
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer c.Close()
	global.C = c
	log.Println("WebSocket 已连接")
	// 循环读取消息
	readLoop()
}

func readLoop() {

	var errorCount int
	for {
		_, message, err := global.C.ReadMessage()
		if err != nil {
			log.Printf("读取消息出错,累计连续错误:%d次。错误信息：%s\n", errorCount, err)
			time.Sleep(10 * time.Second)
			errorCount++
			if errorCount > 10 {
				log.Fatalf("读取消息出错")
			}
		}
		go pkg.HandleMessage(message)
	}
}
