package platform

import "time"

// ChatType 表示会话类型。
type ChatType string

const (
	ChatTypeUnknown ChatType = "unknown"
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
	ChatTypeChannel ChatType = "channel"
)

// Chat 描述一条事件关联的会话。
type Chat struct {
	ID   string
	Type ChatType
	Name string
}

// User 描述一条事件关联的用户。
type User struct {
	ID   string
	Name string
}

// Segment 描述平台消息被拆解后的消息段。
type Segment struct {
	Type string
	Text string
	Data any
}

// Message 表示平台层已经解析完成的消息事件。
type Message struct {
	ID           string
	Chat         Chat
	Sender       User
	Text         string
	Segments     []Segment
	Time         time.Time
	PlatformData any
}

// Notice 表示通知类事件。
type Notice struct {
	Type         string
	Chat         Chat
	User         User
	Target       User
	PlatformData any
}

// Request 表示请求类事件。
type Request struct {
	Type         string
	Chat         Chat
	User         User
	PlatformData any
}
