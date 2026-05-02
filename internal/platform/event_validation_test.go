package platform

import "testing"

func TestEventValidateMessageContract(t *testing.T) {
	valid := Event{
		ID:       "event-1",
		Platform: PlatformQQ,
		Kind:     EventKindMessage,
		SubType:  "group",
		Message: &Message{
			ID:     "msg-1",
			Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
			Sender: User{ID: "user-1"},
			Text:   "hello",
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("合法消息事件校验失败: %v", err)
	}

	invalid := valid
	invalid.Message = &Message{
		ID:     "msg-1",
		Chat:   Chat{ID: "chat-1", Type: ChatTypeGroup},
		Sender: User{},
	}
	if err := invalid.Validate(); err == nil {
		t.Fatal("缺少 sender.id 的消息事件应校验失败")
	}
}

func TestEventValidateNoticeAndRequestContract(t *testing.T) {
	notice := Event{
		ID:       "notice-1",
		Platform: PlatformFeishu,
		Kind:     EventKindNotice,
		Notice: &Notice{
			Type: "message_read",
		},
	}
	if err := notice.Validate(); err != nil {
		t.Fatalf("合法 notice 事件校验失败: %v", err)
	}

	request := Event{
		ID:       "request-1",
		Platform: PlatformQQ,
		Kind:     EventKindRequest,
		Request: &Request{
			Type: "friend",
			User: User{ID: "user-1"},
		},
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("合法 request 事件校验失败: %v", err)
	}

	request.Request = &Request{Type: "friend"}
	if err := request.Validate(); err == nil {
		t.Fatal("缺少 request.user.id 的请求事件应校验失败")
	}
}

func TestEventValidateRejectsMismatchedPayload(t *testing.T) {
	event := Event{
		ID:       "system-1",
		Platform: PlatformTelegramBot,
		Kind:     EventKindSystem,
		Message: &Message{
			ID:     "msg-1",
			Chat:   Chat{ID: "chat-1", Type: ChatTypePrivate},
			Sender: User{ID: "user-1"},
		},
	}
	if err := event.Validate(); err == nil {
		t.Fatal("system 事件携带 message payload 应校验失败")
	}
}

func TestContractNotes(t *testing.T) {
	notes := ContractNotes()
	if len(notes) == 0 {
		t.Fatal("事件契约说明不应为空")
	}
}
