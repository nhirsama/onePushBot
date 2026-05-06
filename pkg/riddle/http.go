package riddle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type response struct {
	Code    int    `json:"code"`
	UserID  string `json:"userid"`
	Message string `json:"message"`
}

func RegisterHTTP(mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}) error {
	if mux == nil {
		return fmt.Errorf("http mux 不能为空")
	}
	mux.HandleFunc("/riddle-264c8c381bf16c982a4e59b0dd4c6f7808c51a05f64c35db42cc78a2a72875bb", latest.Handle)
	mux.HandleFunc("/riddle-84983c60f7daadc1cb8698621f802c0d9f9a3c3c295c810748fb048115c186ec", guestHandler)
	return nil
}

func guestHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "application/json")

	resp := response{
		Code:    http.StatusOK,
		UserID:  "作者",
		Message: "flag{84983c60f7daadc1cb8698621f802c0d9f9a3c3c295c810748fb048115c186ec}",
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "JSON 编码错误", http.StatusInternalServerError)
		log.Printf("编码 riddle guest 响应失败: %v", err)
	}
}
