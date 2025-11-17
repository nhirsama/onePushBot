package napcat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/internal/domain"
	pkg "github.com/nhirsama/onePushBot/pkg/domain"
)

// MessageDispatcher 消息分发器
type MessageDispatcher struct {
	mu  sync.RWMutex
	bus *EventBus.Bus
	// 统计
	totalMessages   atomic.Int64
	handledMessages atomic.Int64
	errorMessages   atomic.Int64
	totalLatency    atomic.Int64
	log             pkg.Log
	manager         domain.ConnectionManager
}

// NewMessageDispatcher 创建消息分发器

func NewMessageDispatcher(log pkg.Log, manager domain.ConnectionManager, bus *EventBus.Bus) domain.MessageDispatcher {
	return &MessageDispatcher{
		log:     log,
		manager: manager,
		bus:     bus,
	}
}

func (m *MessageDispatcher) GetStats() domain.DispatcherStats {
	return domain.DispatcherStats{
		TotalMessages:   m.totalMessages.Load(),
		HandledMessages: m.handledMessages.Load(),
		ErrorMessages:   m.errorMessages.Load(),
		TotalLatency:    m.totalLatency.Load() / 1000000,
	}
}

// Dispatch 分发消息
func (m *MessageDispatcher) Dispatch(ctx context.Context, rawMsg domain.RawMessage) error {
	start := time.Now()
	defer func() {
		end := time.Now()
		diff := end.Sub(start)
		m.totalLatency.Add(diff.Nanoseconds())
	}()
	m.totalMessages.Add(1)
	msg, err := m.parser(&rawMsg)
	if err != nil {
		m.errorMessages.Add(1)
		return err
	}
	switch msg.PostType {
	case domain.PostTypeMessage:
		err = m.HandlerPostTypeMessage(ctx, msg)
	case domain.PostTypeMetaEvent:
		err = m.HandlerPostTypeMetaEvent(ctx, msg)
	case domain.PostTypeNotice:
		err = m.HandlerPostTypeNotice(ctx, msg)
	case domain.PostTypeMessageSent:
		err = m.HandlerPostTypeMessageSent(ctx, msg)
	case domain.PostTypeRequest:
		err = m.HandlerPostTypeRequest(ctx, msg)
	default:
		err = m.HandlerPostTypeOther(ctx, msg)
	}
	if err != nil {
		m.errorMessages.Add(1)
		return err
	}
	m.handledMessages.Add(1)
	return nil
}

func (m *MessageDispatcher) parser(rawMsg *domain.RawMessage) (*domain.MessageStruct, error) {
	var msgs domain.MessageStruct
	if err := json.Unmarshal(*rawMsg, &msgs); err != nil {
		return nil, fmt.Errorf("json解析原始信息失败: %w", err)
	}
	return &msgs, nil
}

func (m *MessageDispatcher) HandlerPostTypeMessage(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	switch domain.MessageType(msg.MessageType) {
	case domain.MessageTypeGroup:
		m.publish("groupMessage", msg)
		m.log.Debug("接收到群组信息:", "mes", msg)
		// 如果群消息里有 @SelfId，则再发 atMe 事件
		if strings.Contains(msg.RawMessage, fmt.Sprintf("[CQ:at,qq=%d", config.SelfId)) {
			m.publish("atMe", msg)
			m.log.Debug("接收到艾特信息")
		}

	case domain.MessageTypePrivate:
		m.publish("privateMessage", msg)
		m.log.Debug("接收到私聊信息:", msg)

	default:
		m.log.Debug("未识别的 message_type")
	}
	return nil
}

// HandlerPostTypeMetaEvent 处理上报类型为 meta_event 的元事件消息（如心跳、生命周期等）
func (m *MessageDispatcher) HandlerPostTypeMetaEvent(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	switch msg.MetaEventType {
	case string(domain.MessageTypeHeartbeat):
		// 心跳事件：用于维持连接状态
		m.publish("meta.heartbeat", msg)
		m.log.Debug("收到心跳元事件")
		if m.manager != nil {
			m.manager.NotifyHeartbeat()
		}

	case string(domain.MessageTypeLifecycle):
		// 生命周期事件：例如 WebSocket 连接建立等
		m.publish("meta.lifecycle", msg)
		if msg.SubType == "connect" {
			m.log.Debug("WebSocket 生命周期事件：连接成功")
		} else {
			m.log.Debug(fmt.Sprintf("收到生命周期元事件，子类型：%s", msg.SubType))
		}

	default:
		// 未知的 meta_event 类型
		m.publish("meta.unknown", msg)
		m.log.Debug(fmt.Sprintf("收到未定义的元事件类型：%s", msg.MetaEventType))
	}

	return nil
}

// HandlerPostTypeNotice 处理上报类型为 notice 的通知类消息（如群内事件、拍一拍等）
func (m *MessageDispatcher) HandlerPostTypeNotice(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 通用 notice 事件
	m.publish("notice", msg)

	switch domain.MessageType(msg.SubType) {
	case domain.MessageTypePoke:
		// 拍一拍通知
		m.publish("poke", msg)
		m.log.Info("收到拍一拍通知事件")

	default:
		// 其他未识别的 notice 子类型
		m.log.Debug(fmt.Sprintf("收到未识别的 notice 子类型：%s", msg.SubType))
	}

	return nil
}

// HandlerPostTypeMessageSent 处理上报类型为 message_sent 的消息（自身发送消息的回执）
func (m *MessageDispatcher) HandlerPostTypeMessageSent(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	m.publish("messageSent", msg)
	m.log.Debug("处理 message_sent 上报完成")

	return nil
}

// HandlerPostTypeRequest 处理上报类型为 request 的消息（例如好友申请、群邀请等请求）
func (m *MessageDispatcher) HandlerPostTypeRequest(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	m.publish("request", msg)
	m.log.Debug(fmt.Sprintf("收到请求类上报，子类型：%s", msg.SubType))

	return nil
}

// HandlerPostTypeOther 处理无法识别或归类到其它类型的上报消息（兜底逻辑）
func (m *MessageDispatcher) HandlerPostTypeOther(ctx context.Context, msg *domain.MessageStruct) error {
	if msg == nil {
		return fmt.Errorf("*domain.MessageStruct 空指针错误")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if msg.Echo == "" {
		// 没有 Echo，视为未定义的普通上报
		m.publish("unknown", msg)
		m.log.Debug("接收到未定义的上报消息（无 Echo）", msg)
		return nil
	}

	// 有 Echo 的一般是 API 调用的回包，交给上层按 Echo 做进一步路由
	m.publish("api_response", msg)
	m.log.Debug(fmt.Sprintf("接收到带 Echo 的上报消息，Echo：%s", msg.Echo))

	return nil
}

func (m *MessageDispatcher) publish(topic string, data interface{}) {
	if m.bus == nil {
		return
	}
	(*m.bus).Publish(topic, data)
}
