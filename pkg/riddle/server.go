package riddle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
	"github.com/spf13/viper"
)

type Response struct {
	Code    int    `json:"code"`
	UserID  string `json:"userid"`
	Message string `json:"message"`
}

type Message struct {
	Mu      sync.RWMutex
	UserId  string
	Message string
}

var Mes Message

type WebSocketMessage struct {
	*api.WebSocketMessage
}

func init() {
	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{Server, "riddle"})
}
func (w *WebSocketMessage) UpdateMessage() {
	w.WebSocketMessage.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
		if msg.GroupId != viper.GetInt64("riddleConfigGroupId") {
			return
		}
		type structMessageArray struct {
			Type string `json:"type"`
			Data struct {
				Text string `json:"text"`
				Qq   string `json:"qq"`
			} `json:"data"`
		}
		var smess []structMessageArray
		json.Unmarshal(msg.Message, &smess)

		type sender struct {
			Nickname string `json:"nickname"`
			Card     string `json:"card"`
		}
		var senders sender
		json.Unmarshal(msg.Sender, &senders)
		var newMessage string
		for _, ms := range smess {
			if ms.Type == "text" {
				newMessage = newMessage + ms.Data.Text
			} else if ms.Type == "at" {
				req, err := w.GetGroupMemberInfo(viper.GetInt64("riddleConfigGroupId"), func(s string) int64 {
					userId, _ := strconv.ParseInt(s, 10, 64)
					return userId
				}(ms.Data.Qq), false)
				if err != nil {
					return
				}
				newMessage = newMessage + fmt.Sprintf("@%s", func(req *api.GroupMemberInfoData) string {
					if req.Card != "" {
						return req.Card
					} else {
						return req.Nickname
					}
				}(req))
			}
		}
		if newMessage == "" {
			return
		}
		Mes.Mu.Lock()
		Mes.Message = newMessage
		if senders.Card != "" {
			Mes.UserId = senders.Card
		} else {
			Mes.UserId = senders.Nickname
		}
		defer Mes.Mu.Unlock()
	}, false)
}

func riddleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	Mes.Mu.RLock()
	response := Response{
		Code:    http.StatusOK,
		UserID:  Mes.UserId,
		Message: Mes.Message,
	}
	Mes.Mu.RUnlock()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// 发送给客户端的错误信息
		http.Error(w, "JSON 编码错误", http.StatusInternalServerError)

		log.Printf("内部错误：编码 JSON 响应失败 (200 响应): %v", err)
		return
	}
	//log.Printf("已处理请求 %s，返回 HTTP 状态码: %d\n", r.URL.Path, http.StatusOK)
}
func riddleGuestHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Code:    http.StatusOK,
		UserID:  "作者",
		Message: "flag{84983c60f7daadc1cb8698621f802c0d9f9a3c3c295c810748fb048115c186ec}",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// 发送给客户端的错误信息
		http.Error(w, "JSON 编码错误", http.StatusInternalServerError)

		log.Printf("内部错误：编码 JSON 响应失败 (200 响应): %v", err)
		return
	}
}
func Server(web *api.WebSocketMessage) {
	serverPort := 12396
	// student 的 sha256 值
	http.HandleFunc("/riddle-264c8c381bf16c982a4e59b0dd4c6f7808c51a05f64c35db42cc78a2a72875bb", riddleHandler)
	// guest 的 sha256 值
	http.HandleFunc("/riddle-84983c60f7daadc1cb8698621f802c0d9f9a3c3c295c810748fb048115c186ec", riddleGuestHandler)
	log.Printf("服务器启动于 http://localhost:%d.", serverPort)
	w := WebSocketMessage{WebSocketMessage: web}
	go w.UpdateMessage()
	// 启动 HTTP 服务器并监听端口
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		log.Fatal("启动服务器错误: ", err)
	}
}
