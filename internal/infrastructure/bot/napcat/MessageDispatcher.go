package napcat

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nhirsama/onePushBot/internal/domain"
)

type MessageDispatcher struct {
	mu sync.RWMutex

	// 统计
	totalMessages   atomic.Int64
	handledMessages atomic.Int64
	errorMessages   atomic.Int64
	totalLatency    atomic.Int64
}

// NewMessageDispatcher 创建消息分发器

func NewMessageDispatcher() domain.MessageDispatcher {
	return &MessageDispatcher{}
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

	m.mu.RLock()
	switch msg.PostType {
	case domain.PostTypeMessage:
		err := m.HandlerPostTypeMessage(ctx, msg)
		if err != nil {
			return err
		}
	case domain.PostTypeMetaEvent:
		err := m.HandlerPostTypeMetaEvent(ctx, msg)
		if err != nil {
			return err
		}
	case domain.PostTypeNotice:
		err := m.HandlerPostTypeNotice(ctx, msg)
		if err != nil {
			return err
		}
	case domain.PostTypeMessageSent:
		err := m.HandlerPostTypeMessageSent(ctx, msg)
		if err != nil {
			return err
		}
	case domain.PostTypeRequest:
		err := m.HandlerPostTypeRequest(ctx, msg)
		if err != nil {
			return err
		}
	case domain.PostTypeOther:
		err := m.HandlerPostTypeOther(ctx, msg)
		if err != nil {
			return err
		}
	}
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
	//TODO implement me
	panic("implement me")
}

func (m *MessageDispatcher) HandlerPostTypeMetaEvent(ctx context.Context, msg *domain.MessageStruct) error {
	//TODO implement me
	panic("implement me")
}

func (m *MessageDispatcher) HandlerPostTypeNotice(ctx context.Context, msg *domain.MessageStruct) error {
	//TODO implement me
	panic("implement me")
}

func (m *MessageDispatcher) HandlerPostTypeMessageSent(ctx context.Context, msg *domain.MessageStruct) error {
	//TODO implement me
	panic("implement me")
}

func (m *MessageDispatcher) HandlerPostTypeRequest(ctx context.Context, msg *domain.MessageStruct) error {
	//TODO implement me
	panic("implement me")
}

func (m *MessageDispatcher) HandlerPostTypeOther(ctx context.Context, msg *domain.MessageStruct) error {
	//TODO implement me
	panic("implement me")
}
