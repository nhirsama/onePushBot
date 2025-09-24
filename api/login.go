package api

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	mu                 sync.RWMutex
	conn               *websocket.Conn
	url                string
	WriteChan          chan commonRequest
	readGoroutineClose chan struct{}
	heartbeat          chan struct{}
}

func NewWebSocketMessage(url string) *WebSocketMessage {
	var w WebSocketMessage
	w.url = url
	w.reLogin()
	w.WriteChan = make(chan commonRequest, 100)
	w.heartbeat = make(chan struct{}, 1)
	return &w
}
func (w *WebSocketMessage) login() error {
	var err error
	log.Printf("正在连接至 %s\n", w.url)
	w.conn, _, err = websocket.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		return err
	}
	log.Println("WebSocket 已连接")
	return nil
}

func (w *WebSocketMessage) reLogin() {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.login()
	if err != nil {
		for reLoginCount := 1; ; reLoginCount++ {
			time.Sleep(time.Second * 5)
			log.Printf("连接失败，正在重新连接第%d次。%s\n", reLoginCount, err)
			err = w.login()
			if err == nil {
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

// Deprecated: 与并发安全设计冲突，请使用 ReadMessageConcurrent
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
