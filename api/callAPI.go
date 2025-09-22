package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type commonRequest struct {
	Action string      `json:"action"`
	Echo   string      `json:"echo"`
	Params interface{} `json:"params"`
}

type commonResponse struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Echo    string          `json:"echo"`
	Data    json.RawMessage `json:"data"` // 使用 RawMessage 延迟解析
}

func (w *WebSocketMessage) callAPI(action string, params interface{}) (*commonResponse, error) {
	var retryCount int = 3
	var timeout time.Duration = time.Second * 20
	for i := 1; i <= retryCount; i++ {
		echo := strconv.FormatInt(time.Now().UnixNano(), 10)
		ch := make(chan []byte, 1)
		ResponseMap.Store(echo, ch)
		defer ResponseMap.Delete(echo)

		// 封装请求体
		request := commonRequest{
			Action: action,
			Echo:   echo,
			Params: params,
		}
		//向写协程发送写信息
		w.WriteChan <- request

		select {
		case msg := <-ch:
			var resp commonResponse
			if err := json.Unmarshal(msg, &resp); err != nil {
				return nil, err
			}
			if resp.Status == "ok" {
				return &resp, nil
			}
			log.Printf("API request error. Action: %s, RetCode: %d, Retry: %d.\n", action, resp.RetCode, i)
		case <-time.After(timeout):
			log.Printf("API request timed out. Action: %s, Retry: %d.\n", action, i)
		}
	}
	return nil, fmt.Errorf("API: %s request error", action)
}
