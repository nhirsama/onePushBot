package platform

import "time"

// EventKind 描述平台事件的大类。
type EventKind string

const (
	EventKindMessage EventKind = "message"
	EventKindNotice  EventKind = "notice"
	EventKindRequest EventKind = "request"
)

// Event 是平台层发布到业务总线的事件。
// 总线只承载上层业务需要处理的消息、通知、请求；连接心跳、生命周期、
// API echo 响应等协议运行细节应由平台驱动自行处理。
type Event struct {
	ID       string
	Platform Platform
	Kind     EventKind
	// SubType 保留平台内细分语义，不保证跨平台同名字段表达同样含义。
	SubType string
	Time    time.Time
	// SelfID 是接收该事件的平台账号 ID，平台无法提供时为空。
	SelfID string

	Message *Message
	Notice  *Notice
	Request *Request

	// Raw 保留平台原始数据，用于业务确实需要访问平台专属字段的场景。
	Raw any
}
