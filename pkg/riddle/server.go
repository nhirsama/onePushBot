package riddle

import (
	"context"
	"fmt"
	"log"
	"strings"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

func Register(rt router.Router, clients platformclient.Source) error {
	if !Enabled() {
		return nil
	}

	groupID := groupID()
	if groupID == "" {
		log.Println("riddle 已启用，但未配置 riddle.group_id/riddleConfigGroupId，仅启动 HTTP 查询接口")
		return nil
	}

	return rt.Register(router.Route{
		Name: "riddle_update_message",
		Filter: base.EventFilter{
			Kind:     base.EventKindMessage,
			ChatType: base.ChatTypeGroup,
			ChatID:   groupID,
		},
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			if event.Message == nil {
				return nil
			}
			client, ok := clientFromEvent(clients, event)
			if !ok {
				return nil
			}

			message, err := buildMessage(ctx, client, groupID, event.Message)
			if err != nil {
				return err
			}
			if message == "" {
				return nil
			}

			latest.Update(senderDisplayName(event.Message.Sender), message)
			return nil
		}),
	})
}

func buildMessage(ctx context.Context, client Client, groupID string, message *base.Message) (string, error) {
	var builder strings.Builder
	for _, segment := range message.Segments {
		switch segment.Type {
		case "text":
			builder.WriteString(segment.Text)
		case "at", "mention":
			targetID := segmentDataString(segment.Data, "qq", "user_id", "userId", "id")
			if targetID == "" || targetID == "all" {
				continue
			}
			member, err := client.GetGroupMemberInfo(ctx, groupID, targetID, false)
			if err != nil {
				return "", err
			}
			builder.WriteString("@")
			builder.WriteString(memberDisplayName(member, targetID))
		}
	}
	if builder.Len() == 0 {
		return message.Text, nil
	}
	return builder.String(), nil
}

func segmentDataString(data any, keys ...string) string {
	fields, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		value, ok := fields[key]
		if !ok || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func memberDisplayName(member *base.GroupMemberInfo, fallback string) string {
	if member == nil {
		return fallback
	}
	if member.Card != "" {
		return member.Card
	}
	if member.Nickname != "" {
		return member.Nickname
	}
	return fallback
}

func senderDisplayName(user base.User) string {
	if user.Card != "" {
		return user.Card
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Name != "" {
		return user.Name
	}
	return user.ID
}
