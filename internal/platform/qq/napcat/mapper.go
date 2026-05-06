package napcat

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

func mapEnvelopeToEvent(raw rawEnvelope, fallbackSelfID string) (base.Event, bool, error) {
	switch raw.PostType {
	case "message", "message_sent":
		msg, err := mapMessage(raw, fallbackSelfID)
		if err != nil {
			return base.Event{}, false, err
		}
		subType := messageEventSubType(raw)
		if raw.PostType == "message_sent" {
			subType = "sent_" + subType
			msg.SentBySelf = true
		}
		return base.Event{
			ID:       buildEventID(raw, msg.ID, subType),
			Platform: base.PlatformQQ,
			Kind:     base.EventKindMessage,
			SubType:  subType,
			Time:     msg.Time,
			SelfID:   normalizeID(raw.SelfID),
			Message:  &msg,
			Raw:      raw,
		}, true, nil
	case "notice":
		notice := mapNotice(raw)
		return base.Event{
			ID:       buildEventID(raw, "", notice.Type),
			Platform: base.PlatformQQ,
			Kind:     base.EventKindNotice,
			SubType:  notice.Type,
			Time:     eventTime(raw.Time),
			SelfID:   normalizeID(raw.SelfID),
			Notice:   &notice,
			Raw:      raw,
		}, true, nil
	case "request":
		request := mapRequest(raw)
		return base.Event{
			ID:       buildEventID(raw, "", request.Type),
			Platform: base.PlatformQQ,
			Kind:     base.EventKindRequest,
			SubType:  request.Type,
			Time:     eventTime(raw.Time),
			SelfID:   normalizeID(raw.SelfID),
			Request:  &request,
			Raw:      raw,
		}, true, nil
	case "meta_event":
		return base.Event{}, false, nil
	case "":
		return base.Event{}, false, nil
	default:
		return base.Event{}, false, nil
	}
}

func mapMessage(raw rawEnvelope, fallbackSelfID string) (base.Message, error) {
	segments, text := parseSegments(raw.Message, raw.RawMessage)
	sender, _ := parseSender(raw.Sender)

	chat := base.Chat{
		ID:   normalizeID(raw.GroupID),
		Type: base.ChatTypeGroup,
		Name: raw.GroupName,
	}
	if raw.MessageType == "private" {
		chat = base.Chat{
			ID:   normalizeID(raw.UserID),
			Type: base.ChatTypePrivate,
		}
	}

	senderID := normalizeID(raw.UserID)
	if sender.ID != "" {
		senderID = sender.ID
	}
	if raw.PostType == "message_sent" && fallbackSelfID != "" {
		senderID = fallbackSelfID
		sender.ID = fallbackSelfID
	}
	if senderID == "" && fallbackSelfID != "" && raw.PostType == "message_sent" {
		senderID = fallbackSelfID
	}

	msg := base.Message{
		ID:           normalizeID(raw.MessageID),
		Chat:         chat,
		Sender:       enrichUserID(sender, senderID),
		Text:         text,
		RawText:      raw.RawMessage,
		Segments:     segments,
		Time:         eventTime(raw.Time),
		Target:       base.User{ID: normalizeID(raw.TargetID)},
		Font:         raw.Font,
		DetailType:   raw.SubType,
		Anonymous:    rawAnonymous(raw.Anonymous),
		PlatformData: raw,
	}
	if msg.Text == "" && raw.RawMessage != "" {
		msg.Text = raw.RawMessage
	}
	return msg, nil
}

func mapNotice(raw rawEnvelope) base.Notice {
	chat := base.Chat{
		ID:   normalizeID(raw.GroupID),
		Type: base.ChatTypeGroup,
		Name: raw.GroupName,
	}
	if chat.ID == "" {
		chat = base.Chat{
			ID:   normalizeID(raw.UserID),
			Type: base.ChatTypePrivate,
		}
	}

	noticeType := raw.NoticeType
	if noticeType == "" {
		noticeType = raw.SubType
	}
	if raw.NoticeType == "notify" && raw.SubType != "" {
		noticeType = raw.SubType
	}

	notice := base.Notice{
		Type:       noticeType,
		Chat:       chat,
		User:       base.User{ID: normalizeID(raw.UserID)},
		Target:     base.User{ID: normalizeID(raw.TargetID)},
		Operator:   base.User{ID: normalizeID(raw.OperatorID)},
		MessageID:  normalizeID(raw.MessageID),
		Duration:   time.Duration(raw.Duration) * time.Second,
		DetailType: raw.NoticeType,
		PlatformData: map[string]any{
			"sub_type": raw.SubType,
			"raw":      raw,
		},
	}
	if file := parseFile(raw.File); file != nil {
		notice.FileID = file.ID
		notice.FileName = file.Name
		notice.FileSize = file.Size
	}
	return notice
}

func mapRequest(raw rawEnvelope) base.Request {
	chat := base.Chat{
		ID:   normalizeID(raw.GroupID),
		Type: base.ChatTypeGroup,
		Name: raw.GroupName,
	}
	if chat.ID == "" {
		chat.Type = base.ChatTypeUnknown
	}

	reqType := raw.RequestType
	if reqType == "" {
		reqType = raw.SubType
	}

	return base.Request{
		Type:       reqType,
		Chat:       chat,
		User:       base.User{ID: normalizeID(raw.UserID)},
		Comment:    raw.Comment,
		Flag:       raw.Flag,
		DetailType: raw.SubType,
		PlatformData: map[string]any{
			"raw": raw,
		},
	}
}

func messageEventSubType(raw rawEnvelope) string {
	if raw.MessageType == "" {
		return raw.SubType
	}
	return raw.MessageType
}

func parseFile(raw json.RawMessage) *struct {
	ID   string
	Name string
	Size int64
} {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var file rawFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil
	}
	id := normalizeID(file.ID)
	if id == "" {
		id = normalizeID(file.BusID)
	}
	return &struct {
		ID   string
		Name string
		Size int64
	}{
		ID:   id,
		Name: file.Name,
		Size: file.Size,
	}
}

func parseSender(raw json.RawMessage) (base.User, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return base.User{}, nil
	}
	var sender rawSender
	if err := json.Unmarshal(raw, &sender); err != nil {
		return base.User{}, err
	}
	name := sender.Card
	if name == "" {
		name = sender.Remark
	}
	if name == "" {
		name = sender.Nickname
	}
	return base.User{
		ID:       normalizeID(sender.UserID),
		Name:     name,
		Nickname: sender.Nickname,
		Card:     sender.Card,
		Remark:   sender.Remark,
		Role:     sender.Role,
		Title:    sender.Title,
		Level:    sender.Level,
	}, nil
}

func enrichUserID(user base.User, id string) base.User {
	user.ID = id
	return user
}

func rawAnonymous(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return string(raw)
	}
	return payload
}

func parseSegments(raw json.RawMessage, fallback string) ([]base.Segment, string) {
	if len(raw) == 0 || string(raw) == "null" {
		if fallback == "" {
			return nil, ""
		}
		return []base.Segment{{Type: "text", Text: fallback, Data: map[string]any{"text": fallback}}}, fallback
	}

	var textPayload string
	if err := json.Unmarshal(raw, &textPayload); err == nil {
		return parseStringMessage(textPayload, fallback)
	}

	var payload []rawSegment
	if err := json.Unmarshal(raw, &payload); err != nil {
		if fallback == "" {
			return nil, ""
		}
		return []base.Segment{{Type: "text", Text: fallback, Data: map[string]any{"text": fallback}}}, fallback
	}

	var (
		segments []base.Segment
		builder  strings.Builder
	)
	for _, segment := range payload {
		text := extractSegmentText(segment)
		if text != "" {
			builder.WriteString(text)
		}
		segments = append(segments, base.Segment{
			Type: segment.Type,
			Text: text,
			Data: segment.Data,
		})
	}

	fullText := builder.String()
	if fullText == "" {
		fullText = fallback
	}
	return segments, fullText
}

func parseStringMessage(payload string, fallback string) ([]base.Segment, string) {
	if payload == "" {
		if fallback == "" {
			return nil, ""
		}
		return []base.Segment{{Type: "text", Text: fallback, Data: map[string]any{"text": fallback}}}, fallback
	}

	segments := make([]base.Segment, 0, 4)
	fullText := strings.Builder{}
	remaining := payload
	for len(remaining) > 0 {
		start := strings.Index(remaining, "[CQ:")
		if start < 0 {
			appendTextSegment(&segments, &fullText, remaining)
			break
		}
		if start > 0 {
			appendTextSegment(&segments, &fullText, remaining[:start])
			remaining = remaining[start:]
		}

		end := strings.IndexByte(remaining, ']')
		if end < 0 {
			appendTextSegment(&segments, &fullText, remaining)
			break
		}

		segment, ok := parseCQSegment(remaining[4:end])
		if !ok {
			appendTextSegment(&segments, &fullText, remaining[:end+1])
		} else {
			segments = append(segments, segment)
		}
		remaining = remaining[end+1:]
	}

	if len(segments) == 0 {
		return []base.Segment{{Type: "text", Text: payload, Data: map[string]any{"text": payload}}}, payload
	}

	text := fullText.String()
	if text == "" {
		text = fallback
	}
	return segments, text
}

func appendTextSegment(segments *[]base.Segment, fullText *strings.Builder, text string) {
	if text == "" {
		return
	}
	*segments = append(*segments, base.Segment{
		Type: "text",
		Text: text,
		Data: map[string]any{"text": text},
	})
	fullText.WriteString(text)
}

func parseCQSegment(payload string) (base.Segment, bool) {
	if payload == "" {
		return base.Segment{}, false
	}

	parts := strings.Split(payload, ",")
	segmentType := parts[0]
	if segmentType == "" {
		return base.Segment{}, false
	}

	data := make(map[string]any, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return base.Segment{}, false
		}
		data[key] = decodeCQText(value)
	}

	return base.Segment{
		Type: segmentType,
		Data: data,
	}, true
}

func decodeCQText(value string) string {
	value = strings.ReplaceAll(value, "&#91;", "[")
	value = strings.ReplaceAll(value, "&#93;", "]")
	value = strings.ReplaceAll(value, "&#44;", ",")
	value = strings.ReplaceAll(value, "&amp;", "&")
	return value
}

func extractSegmentText(segment rawSegment) string {
	if segment.Type != "text" {
		return ""
	}
	value, ok := segment.Data["text"]
	if !ok {
		return ""
	}
	switch text := value.(type) {
	case string:
		return text
	default:
		return fmt.Sprint(text)
	}
}

func buildEventID(raw rawEnvelope, fallbackID, subType string) string {
	if fallbackID != "" {
		return fallbackID
	}
	parts := []string{"qq", raw.PostType, subType, strconv.FormatInt(raw.Time, 10)}
	if raw.Echo != "" {
		parts = append(parts, raw.Echo)
	}
	if id := normalizeID(raw.MessageID); id != "" {
		parts = append(parts, id)
	}
	return strings.Join(parts, ":")
}

func eventTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Now()
	}
	return time.Unix(ts, 0)
}

func normalizeID(value any) string {
	switch id := value.(type) {
	case nil:
		return ""
	case string:
		return id
	case json.Number:
		return id.String()
	case float64:
		if id == math.Trunc(id) {
			return strconv.FormatInt(int64(id), 10)
		}
		return strconv.FormatFloat(id, 'f', -1, 64)
	case float32:
		f := float64(id)
		if f == math.Trunc(f) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	case int:
		return strconv.Itoa(id)
	case int8:
		return strconv.FormatInt(int64(id), 10)
	case int16:
		return strconv.FormatInt(int64(id), 10)
	case int32:
		return strconv.FormatInt(int64(id), 10)
	case int64:
		return strconv.FormatInt(id, 10)
	case uint:
		return strconv.FormatUint(uint64(id), 10)
	case uint8:
		return strconv.FormatUint(uint64(id), 10)
	case uint16:
		return strconv.FormatUint(uint64(id), 10)
	case uint32:
		return strconv.FormatUint(uint64(id), 10)
	case uint64:
		return strconv.FormatUint(id, 10)
	default:
		return fmt.Sprint(id)
	}
}
