package api

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	conn *websocket.Conn
	url  string
}

func NewWebSocketMessage(url string) *WebSocketMessage {
	var w WebSocketMessage
	w.url = url
	w.reLogin()
	return &w
}
func (w *WebSocketMessage) login() error {
	var err error
	log.Printf("正在连接至 %s", w.url)
	w.conn, _, err = websocket.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		return err
	}
	log.Println("WebSocket 已连接")
	return nil
}

func (w *WebSocketMessage) reLogin() {
	err := w.login()
	if err != nil {
		for reLoginCount := 1; ; reLoginCount++ {
			log.Printf("连接失败，正在重新连接第%d次。%s", reLoginCount, err)
			err = w.login()
			if err != nil {
				return
			}
		}
	}
}
func (w *WebSocketMessage) Close() {
	err := w.conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
}

func (w *WebSocketMessage) ReadMessage() (messageType int, p []byte) {
	var err error
	var errorCount int
	messageType, p, err = w.conn.ReadMessage()
	if err != nil {
		log.Printf("读取消息出错,累计连续错误%d次。错误信息：%s\n", errorCount, err)
		time.Sleep(10 * time.Second)
		errorCount++
		if errorCount > 3 {
			log.Println("读取消息出错,正在尝试重新链接")
			w.reLogin()
		}
	}
	return
}
