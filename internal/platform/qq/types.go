package qq

// 常用的 QQ 标识统一使用字符串，避免被不同实现的数值宽度限制。
type (
	GroupID   string
	UserID    string
	MessageID string
)
