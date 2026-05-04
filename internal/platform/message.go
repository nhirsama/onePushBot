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
	ID       string
	Name     string
	Nickname string
	Card     string
	Remark   string
	Role     string
	Title    string
	Level    string
}

// GroupMemberInfo describes group member details shared by platform-specific
// clients that can query group roster data.
type GroupMemberInfo struct {
	GroupID         string `json:"group_id"`
	UserID          string `json:"user_id"`
	Nickname        string `json:"nickname"`
	Card            string `json:"card"`
	Sex             string `json:"sex"`
	Age             int    `json:"age"`
	JoinTime        int64  `json:"join_time"`
	LastSentTime    int64  `json:"last_sent_time"`
	Level           string `json:"level"`
	QQLevel         int    `json:"qq_level"`
	Role            string `json:"role"`
	Title           string `json:"title"`
	Area            string `json:"area"`
	Unfriendly      bool   `json:"unfriendly"`
	TitleExpireTime int64  `json:"title_expire_time"`
	CardChangeable  bool   `json:"card_changeable"`
	ShutUpTimestamp int64  `json:"shut_up_timestamp"`
	IsRobot         bool   `json:"is_robot"`
	QAge            string `json:"qage"`
}

// Segment 描述平台消息被拆解后的消息段。
type Segment struct {
	Type string
	Text string
	Data any
}

// Message 表示平台层已经解析完成的消息事件。
type Message struct {
	ID       string
	Chat     Chat
	Sender   User
	Text     string
	RawText  string
	Segments []Segment
	Time     time.Time
	Target   User
	Font     int
	// SentBySelf 表示该消息由当前平台账号发出，对应 OneBot 的 message_sent。
	SentBySelf bool
	// DetailType 保留平台消息细分类型，例如 QQ normal/friend、飞书 text/image。
	DetailType string
	// Anonymous 保留匿名群消息信息，平台无此能力时为空。
	Anonymous    any
	PlatformData any
}

// Notice 表示通知类事件。
type Notice struct {
	Type   string
	Chat   Chat
	User   User
	Target User
	// Operator 表示触发操作的用户，例如群管理变更、禁言、文件上传等场景。
	Operator User
	// MessageID 关联被撤回、被回应等通知所指向的消息。
	MessageID string
	// FileID/Name/Size 用于文件上传等通知。
	FileID   string
	FileName string
	FileSize int64
	// Duration 用于禁言等带持续时间的通知。
	Duration     time.Duration
	DetailType   string
	PlatformData any
}

// Request 表示请求类事件。
type Request struct {
	Type         string
	Chat         Chat
	User         User
	Comment      string
	Flag         string
	DetailType   string
	PlatformData any
}
