package napcat

import "encoding/json"

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
	Message       json.RawMessage `json:"message"`
	RawMessage    string          `json:"raw_message"`
	GroupID       any             `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserID        any             `json:"user_id"`
	TargetID      any             `json:"target_id"`
	NoticeType    string          `json:"notice_type"`
	RequestType   string          `json:"request_type"`
	SubType       string          `json:"sub_type"`
	Interval      int64           `json:"interval"`
	Status        string          `json:"status"`
	RetCode       int             `json:"retcode"`
	Data          json.RawMessage `json:"data"`
	Echo          string          `json:"echo"`
	Comment       string          `json:"comment"`
	Flag          string          `json:"flag"`
}

type rawSender struct {
	UserID   any    `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Remark   string `json:"remark"`
}

type rawSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
