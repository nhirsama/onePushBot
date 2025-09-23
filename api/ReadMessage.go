package api

import (
	"log"
	"time"
)

func (w *WebSocketMessage) ReadMessageConcurrent() {
	timeoutLength := 60 * time.Second
	w.readGoroutineClose = make(chan struct{}, 1)
	w.mu.RLock()
	defer w.mu.RUnlock()
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
