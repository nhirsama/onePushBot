package api

import "time"

func (w *WebSocketMessage) Start() {
	timeoutLength := 60 * time.Second
	timer := time.NewTimer(timeoutLength)
	defer timer.Stop()
	go w.ReadMessageConcurrent()
	go w.WriteMessage()
	for {
		select {
		case <-w.readMessageConcurrentGoroutineDone:
			w.reLogin()
			w.readMessageConcurrentGoroutineDone = make(chan struct{})
			go w.ReadMessageConcurrent()
		case <-w.heartbeat:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeoutLength)
		case <-timer.C:
			w.readGoroutineClose <- struct{}{}
		}
	}
}
