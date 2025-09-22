package api

import (
	"log"
	"time"
)

func (w *WebSocketMessage) ReadMessageConcurrent() {
	w.readMessageConcurrentGoroutineDone = make(chan struct{})
	defer close(w.readMessageConcurrentGoroutineDone)
	timeoutLength := 60 * time.Second
	w.readGoroutineClose = make(chan struct{}, 1)
	for {
		_ = w.conn.SetReadDeadline(time.Now().Add(timeoutLength))
		select {
		case <-w.readGoroutineClose:
			return
		default:
			_, msg, err := w.conn.ReadMessage()
			if err != nil {
				log.Println("ReadMessageConcurrent:", err)
				return
			}
			go w.handleMessage(msg)
		}
	}
}
