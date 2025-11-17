package napcat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/nhirsama/onePushBot/internal/domain"
	pkg "github.com/nhirsama/onePushBot/pkg/domain"
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

type Distribute struct {
	domain.ConnectionManager
	retry       int
	log         pkg.Log
	responseMap sync.Map
	bus         *EventBus.Bus
}

func (d *Distribute) Start(bus *EventBus.Bus) error {
	d.bus = bus
	// 订阅来自 EventBus 的 API 响应
	err := (*d.bus).SubscribeAsync("api_response", func(msg *domain.MessageStruct) {
		d.handleAPIResponse(msg)
	},
		false,
	)
	if err != nil {
		return fmt.Errorf("订阅 api_response 失败：%w", err)
	}

	d.log.Info("订阅 api_response")
	return nil
}

func NewDistribute(c *domain.ConnectionManager, log pkg.Log) domain.ApiDistribute {
	var d Distribute
	d.ConnectionManager = *c
	d.retry = 3
	d.log = log
	d.responseMap = sync.Map{}
	return &d
}
func (d *Distribute) Call(ctx context.Context, action string, params any) (resp []byte, err error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx2, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		ctx = ctx2
	}

	for i := 1; i <= d.retry; i++ {
		echo := strconv.FormatInt(time.Now().UnixNano(), 10)
		ch := make(chan []byte, 1)
		d.responseMap.Store(echo, ch)

		// 构造请求
		req := commonRequest{
			Action: action,
			Echo:   echo,
			Params: params,
		}

		jsonReq, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("序列化Json失败：%w", err)
		}

		select {
		case d.WriteChan() <- jsonReq:
			d.log.Debug(fmt.Sprintf("调用%s成功，发送内容：%s", action, req))
		case <-ctx.Done():
			d.responseMap.Delete(echo)
			d.log.Debug(fmt.Sprintf("发送 API 超时。Action：%s，第 %d 次尝试，错误：%v", action, i, ctx.Err()))
			return nil, ctx.Err()
		}

		select {
		case msg := <-ch:
			d.responseMap.Delete(echo)
			d.log.Debug(fmt.Sprintf("收到 API 响应。Action：%s，Echo：%s，原始消息：%s", action, echo, string(msg)))

			var r commonResponse
			if err := json.Unmarshal(msg, &r); err != nil {
				d.log.Debug(fmt.Sprintf("解析 API 响应失败。Action：%s，错误：%v", action, err))
				continue
			}

			if r.Status == "ok" {
				d.log.Info(fmt.Sprintf("API 调用成功。Action：%s，Echo：%s", action, echo))
				return r.Data, nil
			}
			d.log.Debug(fmt.Sprintf("API 返回错误。Action：%s，RetCode：%d，第 %d 次尝试失败", action, r.RetCode, i))

		case <-ctx.Done():
			d.responseMap.Delete(echo)
			d.log.Debug(fmt.Sprintf("等待 API 响应超时。Action：%s，第 %d 次尝试，错误：%v", action, i, ctx.Err()))
			return nil, ctx.Err()
		}

	}

	return nil, fmt.Errorf("API：%s 调用连续失败）", action)
}

func (d *Distribute) handleAPIResponse(ms *domain.MessageStruct) {
	if ms == nil {
		d.log.Error("收到空的 MessageStruct")
		return
	}

	echo := ms.Echo

	d.log.Debug(fmt.Sprintf("收到 API 响应：Echo=%s", echo))

	// 根据 Echo 读取对应的 channel
	val, ok := d.responseMap.Load(echo)
	if !ok {
		d.log.Error(fmt.Sprintf("未找到 Echo=%s 对应的回调通道，可能已超时或已清理", echo))
		return
	}

	ch, ok := val.(chan []byte)
	if !ok {
		d.log.Error("responseMap 中存储的不是 chan []byte 类型")
		return
	}

	// 将结构体编码成 JSON（Call() 期望的是 []byte）
	raw, err := json.Marshal(ms)
	if err != nil {
		d.log.Error(fmt.Sprintf("MessageStruct JSON 编码失败：%v", err))
		return
	}

	// 将 JSON 数据写入对应通道
	select {
	case ch <- raw:
		d.log.Info(fmt.Sprintf("已将 API 响应分发到 Echo=%s 的回调通道", echo))
	default:
		d.log.Error(fmt.Sprintf("回调通道已阻塞，无法写入 Echo=%s", echo))
	}
}
