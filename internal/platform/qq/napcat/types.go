package napcat

import (
	"encoding/json"

	"github.com/nhirsama/onePushBot/internal/platform/qq"
)

type rawRequest struct {
	Action string `json:"action"`
	Echo   string `json:"echo"`
	Params any    `json:"params"`
}

type rawEnvelope struct {
	Time          int64           `json:"time"`
	SelfID        any             `json:"self_id"`
	PostType      string          `json:"post_type"`
	MetaEventType string          `json:"meta_event_type"`
	MessageType   string          `json:"message_type"`
	MessageID     any             `json:"message_id"`
	MessageSeq    json.RawMessage `json:"message_seq"`
	RealID        json.RawMessage `json:"real_id"`
	RealSeq       json.RawMessage `json:"real_seq"`
	Sender        json.RawMessage `json:"sender"`
	Anonymous     json.RawMessage `json:"anonymous"`
	Message       json.RawMessage `json:"message"`
	RawMessage    string          `json:"raw_message"`
	Font          int             `json:"font"`
	GroupID       any             `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserID        any             `json:"user_id"`
	TargetID      any             `json:"target_id"`
	OperatorID    any             `json:"operator_id"`
	NoticeType    string          `json:"notice_type"`
	RequestType   string          `json:"request_type"`
	SubType       string          `json:"sub_type"`
	Duration      int64           `json:"duration"`
	File          json.RawMessage `json:"file"`
	Interval      int64           `json:"interval"`
	Echo          string          `json:"echo"`
	Comment       string          `json:"comment"`
	Flag          string          `json:"flag"`
}

type rawResponseEnvelope struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Data    json.RawMessage `json:"data"`
	Echo    string          `json:"echo"`
}

type rawHeartbeatStatus struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}

type rawHeartbeatEnvelope struct {
	PostType      string             `json:"post_type"`
	MetaEventType string             `json:"meta_event_type"`
	Status        rawHeartbeatStatus `json:"status"`
	Interval      int64              `json:"interval"`
}

type rawFile struct {
	ID    any    `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	BusID any    `json:"busid"`
}

type rawSender struct {
	UserID   any    `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Remark   string `json:"remark"`
	Role     string `json:"role"`
	Title    string `json:"title"`
	Level    string `json:"level"`
}

type rawSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type rawGroupMemberInfo struct {
	GroupID         any    `json:"group_id"`
	UserID          any    `json:"user_id"`
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

type rawLoginInfo struct {
	UserID   any    `json:"user_id"`
	Nickname string `json:"nickname"`
}

func (m rawLoginInfo) toDomain() *qq.LoginInfo {
	return &qq.LoginInfo{
		UserID:   normalizeID(m.UserID),
		Nickname: m.Nickname,
	}
}

type rawFriendInfo struct {
	UserID       any    `json:"user_id"`
	Nickname     string `json:"nickname"`
	Remark       string `json:"remark"`
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

func (m rawFriendInfo) toDomain() qq.FriendInfo {
	return qq.FriendInfo{
		UserID:       normalizeID(m.UserID),
		Nickname:     m.Nickname,
		Remark:       m.Remark,
		CategoryID:   m.CategoryID,
		CategoryName: m.CategoryName,
	}
}

type rawGroupInfo struct {
	GroupID        any    `json:"group_id"`
	GroupName      string `json:"group_name"`
	MemberCount    int    `json:"member_count"`
	MaxMemberCount int    `json:"max_member_count"`
}

func (m rawGroupInfo) toDomain() qq.GroupInfo {
	return qq.GroupInfo{
		GroupID:        normalizeID(m.GroupID),
		GroupName:      m.GroupName,
		MemberCount:    m.MemberCount,
		MaxMemberCount: m.MaxMemberCount,
	}
}

type rawSendMessageResult struct {
	MessageID  any    `json:"message_id"`
	ResourceID string `json:"res_id"`
}

func (m rawSendMessageResult) toDomain() *qq.SendMessageResult {
	return &qq.SendMessageResult{
		MessageID:  normalizeID(m.MessageID),
		ResourceID: m.ResourceID,
	}
}
