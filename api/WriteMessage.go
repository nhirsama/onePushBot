package api

import "log"

func (w *WebSocketMessage) WriteMessage() {
	for {
		select {
		case request := <-w.WriteChan:
			for retryCount := 0; retryCount < 3; retryCount++ {
				w.mu.Lock()
				err := w.conn.WriteJSON(request)
				w.mu.Unlock()
				if err != nil {
					log.Println("WriteJSON error:", err)
				} else {
					break
				}
			}
		}
	}
}
