package platform

import "time"

// EventKind 描述平台事件的大类。
type EventKind string

const (
	EventKindMessage EventKind = "message"
	EventKindNotice  EventKind = "notice"
	EventKindRequest EventKind = "request"
	EventKindSystem  EventKind = "system"
	EventKindRaw     EventKind = "raw"
)

// Event 是平台层发布到统一总线的标准事件。
// 这层只负责承载各平台已经解析好的事件，不负责兼容发送或插件路由。
type Event struct {
	ID       string
	Platform Platform
	Kind     EventKind
	SubType  string
	Time     time.Time

	Message *Message
	Notice  *Notice
	Request *Request

	Raw any
}
