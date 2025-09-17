package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/global"
	"github.com/nhirsama/onePushBot/pkg"
)

func main() {
	//创建配置文件
	data, err := os.ReadFile("data/config.json")
	if err != nil {
		log.Println(err)
		err = pkg.SetConfig()
		if err != nil {
			log.Fatal(err)
		}
		data, err = os.ReadFile("data/config.json")
		if err != nil {
			log.Fatal(err)
		}
	}
	var config pkg.Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	// 构造 WebSocket URL
	u := url.URL{Scheme: "wss", Host: config.ApiUrl, Path: "/ws", RawQuery: "access_token=" + config.Token}
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
		go pkg.HandleMessage(message, global.C)
	}
}
