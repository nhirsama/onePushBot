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
		case <-w.heartbeat:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeoutLength)
		case <-timer.C:
			w.readGoroutineClose <- struct{}{}
			w.reLogin()
			go w.ReadMessageConcurrent()
		}
	}
}
