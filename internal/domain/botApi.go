package domain

import "encoding/json"

// CommonRequest 通用请求结构
type CommonRequest struct {
	Action string      `json:"action"`
	Echo   string      `json:"echo"`
	Params interface{} `json:"params"`
}

// CommonResponse 通用响应结构
type CommonResponse struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Echo    string          `json:"echo"`
	Data    json.RawMessage `json:"data"`
}
