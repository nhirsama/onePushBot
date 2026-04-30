package platform

import (
	"context"
	"testing"
)

type stubClient struct {
	platform Platform
	events   chan Event
}

func (s *stubClient) Platform() Platform          { return s.platform }
func (s *stubClient) Start(context.Context) error { return nil }
func (s *stubClient) Close(context.Context) error { return nil }
func (s *stubClient) Events() <-chan Event        { return s.events }
func (s *stubClient) Status() Status              { return StatusStopped }

func TestRegistryRegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	client := &stubClient{
		platform: PlatformQQ,
		events:   make(chan Event),
	}

	if err := registry.Register(client); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	got, ok := registry.Get(PlatformQQ)
	if !ok {
		t.Fatal("未找到已注册的平台")
	}
	if got != client {
		t.Fatal("查询结果与注册实例不一致")
	}

	if err := registry.Register(client); err != ErrClientExists {
		t.Fatalf("重复注册应返回 ErrClientExists，got=%v", err)
	}
}

func TestRegistryRejectInvalidClient(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(nil); err != ErrClientNil {
		t.Fatalf("nil client 应返回 ErrClientNil，got=%v", err)
	}

	invalid := &stubClient{
		events: make(chan Event),
	}
	if err := registry.Register(invalid); err != ErrClientInvalid {
		t.Fatalf("空平台应返回 ErrClientInvalid，got=%v", err)
	}
}
