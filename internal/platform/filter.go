package platform

// EventFilter 描述订阅时的筛选条件。
// 空字段表示不限制该维度。
type EventFilter struct {
	Platform Platform
	Kind     EventKind
	SubType  string
	ChatID   string
	UserID   string
}

func (f EventFilter) Match(event Event) bool {
	if f.Platform != "" && event.Platform != f.Platform {
		return false
	}
	if f.Kind != "" && event.Kind != f.Kind {
		return false
	}
	if f.SubType != "" && event.SubType != f.SubType {
		return false
	}
	if f.ChatID != "" && f.ChatID != eventChatID(event) {
		return false
	}
	if f.UserID != "" && f.UserID != eventUserID(event) {
		return false
	}
	return true
}

func eventChatID(event Event) string {
	switch {
	case event.Message != nil:
		return event.Message.Chat.ID
	case event.Notice != nil:
		return event.Notice.Chat.ID
	case event.Request != nil:
		return event.Request.Chat.ID
	default:
		return ""
	}
}

func eventUserID(event Event) string {
	switch {
	case event.Message != nil:
		return event.Message.Sender.ID
	case event.Notice != nil:
		return event.Notice.User.ID
	case event.Request != nil:
		return event.Request.User.ID
	default:
		return ""
	}
}
