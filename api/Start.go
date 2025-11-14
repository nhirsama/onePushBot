package api

import (
	"log"
	"time"
)

func (w *WebSocketMessage) Start() {
	timeoutLength := 40 * time.Second
	timer := time.NewTimer(timeoutLength)
	defer timer.Stop()
	go w.ReadMessageConcurrent()
	go w.WriteMessage()
	go func() {
		err := w.Scheduler.Start()
		if err != nil {
			log.Fatalf("Error starting scheduler: %v", err)
		}
	}()
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
			select {
			case <-w.readGoroutineClose:
			default:
			}
			go w.ReadMessageConcurrent()
		}
	}
}
