package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
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
	u := url.URL{Scheme: "ws", Host: config.ApiUrl, Path: "/ws", RawQuery: "access_token=" + config.Token}
	log.Printf("Connecting to %s", u.String())

	// 建立连接
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer c.Close()

	log.Println("WebSocket 已连接")

	// 循环读取消息
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("读取消息出错:", err)
			time.Sleep(1 * time.Second)
		}
		//log.Printf("收到消息: %s", message)
		pkg.HandleMessage(message, c)
	}
}
