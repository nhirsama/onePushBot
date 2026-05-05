package groupmessagetokendb

import (
	"context"
	"log"
	"strconv"
	"strings"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/info_entropy"
	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("group_message_token_db.enabled")
}

func Register(rt router.Router) error {
	if !Enabled() {
		return nil
	}

	db := infoEntropy.LoadDB()
	if db == nil {
		log.Println("group_message_token_db 初始化失败，跳过")
		return nil
	}

	return rt.Register(router.Route{
		Name: "group_message_token_db",
		Filter: base.EventFilter{
			Kind:     base.EventKindMessage,
			ChatType: base.ChatTypeGroup,
		},
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			_ = ctx
			if event.Message == nil || !hasTextSegment(event.Message) {
				return nil
			}

			userID, err := strconv.ParseInt(event.Message.Sender.ID, 10, 64)
			if err != nil {
				return nil
			}

			text := event.Message.RawText
			if text == "" {
				text = event.Message.Text
			}
			if text == "" {
				return nil
			}

			db.UpdateUser(userID, text)
			db.Save()
			return nil
		}),
	})
}

func hasTextSegment(message *base.Message) bool {
	if message == nil {
		return false
	}
	for _, segment := range message.Segments {
		if segment.Type == "text" && strings.TrimSpace(segment.Text) != "" {
			return true
		}
	}
	return len(message.Segments) == 0 && strings.TrimSpace(message.Text) != ""
}
