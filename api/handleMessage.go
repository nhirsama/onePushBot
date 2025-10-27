package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nhirsama/onePushBot/config"
)

type MessageStruct struct {
	Time          int64           `json:"time"`
	SelfId        int64           `json:"self_id"`
	PostType      string          `json:"post_type"`
	MetaEventType string          `json:"meta_event_type"`
	Interval      int64           `json:"interval"`
	MessageId     int64           `json:"message_id"`
	MessageSeq    json.RawMessage `json:"message_seq"`
	RealId        json.RawMessage `json:"real_id"`
	RealSeq       json.RawMessage `json:"real_seq"`
	MessageType   string          `json:"message_type"`
	Sender        json.RawMessage `json:"sender"`
	RawMessage    string          `json:"raw_message"`
	Font          json.RawMessage `json:"font"`
	SubType       string          `json:"sub_type"`
	Message       json.RawMessage `json:"message"`
	MessageFormat string          `json:"message_format"`
	GroupId       int64           `json:"group_id"`
	GroupName     string          `json:"group_name"`
	UserId        int64           `json:"user_id"`
	Status        json.RawMessage `json:"status"`
	RetCode       int64           `json:"retcode"`
	TargetId      int64           `json:"target_id"`
	Wording       string          `json:"wording"`
	Echo          string          `json:"echo"`
}

func (w *WebSocketMessage) handleMessage(msg []byte) {
	var messStruct MessageStruct
	if err := json.Unmarshal(msg, &messStruct); err != nil {
		log.Println("json解析失败：", string(msg))
	}

	switch messStruct.PostType {
	case "message":
		w.handleMessageMessage(messStruct)
		log.Println(string(msg))
	case "meta_event":
		w.metaEvent(msg)
	case "notice":
		w.Bus.Publish("notice", &messStruct)
		switch messStruct.SubType {
		case "poke":
			w.Bus.Publish("poke", &messStruct)
			w.pokeLog(msg)
		}
	case "message_sent":
		w.Bus.Publish("messageSent", &messStruct)
	default:
		switch messStruct.Echo {
		case "":
			log.Printf("接收到未定义的信息:%s\n", string(msg))
		default:
			if ch, ok := w.responseMap.Load(messStruct.Echo); ok {
				ch.(chan []byte) <- msg
			} else {
				log.Printf("接收到未追踪的返回值：%s", msg)
			}
		}

	}
}

func (w *WebSocketMessage) handleMessageMessage(ms MessageStruct) {
	switch ms.MessageType {
	case "group":
		w.Bus.Publish("groupMessage", &ms)
		if strings.Contains(ms.RawMessage, fmt.Sprintf("[CQ:at,qq=%d", config.SelfId)) {
			w.Bus.Publish("atMe", &ms)
			log.Println("接收到艾特信息")
		}

	case "private":
		w.Bus.Publish("privateMessage", &ms)
	}
}

func (w *WebSocketMessage) pokeLog(ms []byte) {
	type RawItem struct {
		Txt  string `json:"txt"`
		Type string `json:"type"`
	}

	type Message struct {
		UserID   int       `json:"user_id"`
		TargetID int       `json:"target_id"`
		RawInfo  []RawItem `json:"raw_info"`
	}
	var msg Message
	if err := json.Unmarshal(ms, &msg); err != nil {
		log.Println("拍一拍信息解析失败", err, string(ms))
	}
	var action, tail string
	for _, item := range msg.RawInfo {
		if item.Type == "nor" {
			if action == "" {
				action = item.Txt
			} else {
				tail = item.Txt
			}
		}
	}

	log.Printf("%d %s %d %s\n", msg.UserID, action, msg.TargetID, tail)
}
