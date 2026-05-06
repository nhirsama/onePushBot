package platform

import "strings"

const (
	messagePreviewLimit = 120
	commentPreviewLimit = 80
)

// EventLogFields 返回标准化业务事件的摘要字段，供总线和路由日志复用。
func EventLogFields(event Event) []any {
	fields := []any{
		"platform", event.Platform,
		"event_id", event.ID,
		"kind", event.Kind,
	}

	if event.SubType != "" {
		fields = append(fields, "sub_type", event.SubType)
	}
	if event.SelfID != "" {
		fields = append(fields, "self_id", event.SelfID)
	}

	switch {
	case event.Message != nil:
		fields = appendMessageLogFields(fields, event.Message)
	case event.Notice != nil:
		fields = appendNoticeLogFields(fields, event.Notice)
	case event.Request != nil:
		fields = appendRequestLogFields(fields, event.Request)
	}

	return fields
}

func appendMessageLogFields(fields []any, message *Message) []any {
	if message == nil {
		return fields
	}
	if message.Chat.Type != "" {
		fields = append(fields, "chat_type", message.Chat.Type)
	}
	if message.Chat.ID != "" {
		fields = append(fields, "chat_id", message.Chat.ID)
	}
	if message.Sender.ID != "" {
		fields = append(fields, "user_id", message.Sender.ID)
	}
	if message.Target.ID != "" {
		fields = append(fields, "target_id", message.Target.ID)
	}
	if message.ID != "" {
		fields = append(fields, "message_id", message.ID)
	}
	if message.DetailType != "" {
		fields = append(fields, "detail_type", message.DetailType)
	}
	if message.SentBySelf {
		fields = append(fields, "sent_by_self", true)
	}
	if text := strings.TrimSpace(message.Text); text != "" {
		fields = append(fields, "message_preview", truncateLogText(text, messagePreviewLimit))
	}
	return fields
}

func appendNoticeLogFields(fields []any, notice *Notice) []any {
	if notice == nil {
		return fields
	}
	if notice.Type != "" {
		fields = append(fields, "notice_type", notice.Type)
	}
	if notice.Chat.Type != "" {
		fields = append(fields, "chat_type", notice.Chat.Type)
	}
	if notice.Chat.ID != "" {
		fields = append(fields, "chat_id", notice.Chat.ID)
	}
	if notice.User.ID != "" {
		fields = append(fields, "user_id", notice.User.ID)
	}
	if notice.Target.ID != "" {
		fields = append(fields, "target_id", notice.Target.ID)
	}
	if notice.Operator.ID != "" {
		fields = append(fields, "operator_id", notice.Operator.ID)
	}
	if notice.MessageID != "" {
		fields = append(fields, "message_id", notice.MessageID)
	}
	if notice.DetailType != "" {
		fields = append(fields, "detail_type", notice.DetailType)
	}
	return fields
}

func appendRequestLogFields(fields []any, request *Request) []any {
	if request == nil {
		return fields
	}
	if request.Type != "" {
		fields = append(fields, "request_type", request.Type)
	}
	if request.Chat.Type != "" {
		fields = append(fields, "chat_type", request.Chat.Type)
	}
	if request.Chat.ID != "" {
		fields = append(fields, "chat_id", request.Chat.ID)
	}
	if request.User.ID != "" {
		fields = append(fields, "user_id", request.User.ID)
	}
	if request.Flag != "" {
		fields = append(fields, "request_flag", request.Flag)
	}
	if request.DetailType != "" {
		fields = append(fields, "detail_type", request.DetailType)
	}
	if comment := strings.TrimSpace(request.Comment); comment != "" {
		fields = append(fields, "comment_preview", truncateLogText(comment, commentPreviewLimit))
	}
	return fields
}

func truncateLogText(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
