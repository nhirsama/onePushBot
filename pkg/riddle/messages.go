package riddle

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type messages struct {
	mu      sync.RWMutex
	userID  string
	message string
}

var latest messages

func (m *messages) Update(userID string, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userID = userID
	m.message = message
}

func (m *messages) Handle(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "application/json")

	m.mu.RLock()
	resp := response{
		Code:    http.StatusOK,
		UserID:  m.userID,
		Message: m.message,
	}
	m.mu.RUnlock()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "JSON 编码错误", http.StatusInternalServerError)
		log.Printf("编码 riddle 响应失败: %v", err)
	}
}
