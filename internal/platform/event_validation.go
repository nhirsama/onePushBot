package platform

import (
	"fmt"
	"strings"
)

// Validate 检查事件是否满足公共事件契约。
//
// 约束只覆盖路由层依赖的稳定结构，不强制统一各平台的业务语义。
// 例如 SubType 只保证“平台内稳定”，不承诺跨平台同义。
func (e Event) Validate() error {
	if e.Platform == "" {
		return fmt.Errorf("%w: platform is required", ErrEventInvalid)
	}
	if e.Kind == "" {
		return fmt.Errorf("%w: kind is required", ErrEventInvalid)
	}
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrEventInvalid)
	}

	payloads := 0
	if e.Message != nil {
		payloads++
	}
	if e.Notice != nil {
		payloads++
	}
	if e.Request != nil {
		payloads++
	}
	if payloads > 1 {
		return fmt.Errorf("%w: only one payload can be set", ErrEventInvalid)
	}

	switch e.Kind {
	case EventKindMessage:
		if e.Message == nil {
			return fmt.Errorf("%w: message payload is required for message event", ErrEventInvalid)
		}
		return validateMessage(e.Message)
	case EventKindNotice:
		if e.Notice == nil {
			return fmt.Errorf("%w: notice payload is required for notice event", ErrEventInvalid)
		}
		return validateNotice(e.Notice)
	case EventKindRequest:
		if e.Request == nil {
			return fmt.Errorf("%w: request payload is required for request event", ErrEventInvalid)
		}
		return validateRequest(e.Request)
	case EventKindSystem, EventKindRaw:
		if payloads > 0 {
			return fmt.Errorf("%w: %s event should not carry message/notice/request payload", ErrEventInvalid, e.Kind)
		}
		return nil
	default:
		return fmt.Errorf("%w: unsupported event kind %q", ErrEventInvalid, e.Kind)
	}
}

// MustValidate 便于平台实现或测试在构造事件时快速发现问题。
func (e Event) MustValidate() {
	if err := e.Validate(); err != nil {
		panic(err)
	}
}

func validateMessage(message *Message) error {
	if message == nil {
		return fmt.Errorf("%w: message payload is nil", ErrEventInvalid)
	}
	if strings.TrimSpace(message.ID) == "" {
		return fmt.Errorf("%w: message.id is required", ErrEventInvalid)
	}
	if strings.TrimSpace(message.Chat.ID) == "" {
		return fmt.Errorf("%w: message.chat.id is required", ErrEventInvalid)
	}
	if message.Chat.Type == "" {
		return fmt.Errorf("%w: message.chat.type is required", ErrEventInvalid)
	}
	if strings.TrimSpace(message.Sender.ID) == "" {
		return fmt.Errorf("%w: message.sender.id is required", ErrEventInvalid)
	}
	return nil
}

func validateNotice(notice *Notice) error {
	if notice == nil {
		return fmt.Errorf("%w: notice payload is nil", ErrEventInvalid)
	}
	if strings.TrimSpace(notice.Type) == "" {
		return fmt.Errorf("%w: notice.type is required", ErrEventInvalid)
	}
	return nil
}

func validateRequest(request *Request) error {
	if request == nil {
		return fmt.Errorf("%w: request payload is nil", ErrEventInvalid)
	}
	if strings.TrimSpace(request.Type) == "" {
		return fmt.Errorf("%w: request.type is required", ErrEventInvalid)
	}
	if strings.TrimSpace(request.User.ID) == "" {
		return fmt.Errorf("%w: request.user.id is required", ErrEventInvalid)
	}
	return nil
}

// IsValid 是 Validate 的便捷包装。
func (e Event) IsValid() bool {
	return e.Validate() == nil
}

// ContractNotes 返回事件公共契约的简短说明，供上层记录或测试断言使用。
func ContractNotes() []string {
	return []string{
		"Event.ID, Event.Platform, Event.Kind are always required.",
		"Message events require Message.ID, Message.Chat.ID, Message.Chat.Type, Message.Sender.ID.",
		"Notice events require Notice.Type.",
		"Request events require Request.Type and Request.User.ID.",
		"SubType is platform-specific metadata and is not guaranteed to be cross-platform compatible.",
		"PlatformData and Raw preserve platform-specific details for upper layers that need them.",
	}
}
