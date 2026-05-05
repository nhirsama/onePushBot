package qq

// LoginInfo 表示当前登录账号信息。
type LoginInfo struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

// BotStatus 表示 OneBot/QQ 客户端运行状态。
type BotStatus struct {
	Online bool           `json:"online"`
	Good   bool           `json:"good"`
	Stat   map[string]any `json:"stat,omitempty"`
}

// VersionInfo 表示服务端版本信息。
type VersionInfo struct {
	AppName         string `json:"app_name"`
	AppVersion      string `json:"app_version"`
	ProtocolVersion string `json:"protocol_version"`
	Version         string `json:"version,omitempty"`
}

// FriendInfo 表示好友基础信息。
type FriendInfo struct {
	UserID       string `json:"user_id"`
	Nickname     string `json:"nickname"`
	Remark       string `json:"remark"`
	CategoryID   int    `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

// GroupInfo 表示群基础信息。
type GroupInfo struct {
	GroupID        string `json:"group_id"`
	GroupName      string `json:"group_name"`
	MemberCount    int    `json:"member_count,omitempty"`
	MaxMemberCount int    `json:"max_member_count,omitempty"`
}

// SendMessageRequest 表示通用发消息请求。
type SendMessageRequest struct {
	MessageType string
	UserID      string
	GroupID     string
	Message     any
	AutoEscape  bool
}

// SendForwardMessageRequest 表示通用转发消息请求。
type SendForwardMessageRequest struct {
	MessageType string
	UserID      string
	GroupID     string
	Messages    any
}

// SendMessageResult 表示发消息结果。
type SendMessageResult struct {
	MessageID  string `json:"message_id"`
	ResourceID string `json:"res_id,omitempty"`
}

type (
	GroupID   string
	UserID    string
	MessageID string
)
