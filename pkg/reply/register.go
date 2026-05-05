package reply

import (
	"context"
	"log"
	"strings"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/llm"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

func Register(rt router.Router, clients platformclient.Source) error {
	if !Enabled() {
		return nil
	}

	if err := rt.Register(router.Route{
		Name: "reply_at_me",
		Filter: base.EventFilter{
			Kind:     base.EventKindMessage,
			ChatType: base.ChatTypeGroup,
		},
		Match: matchMentionedSelf(),
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			if event.Message == nil || strings.TrimSpace(event.Message.Text) == "" {
				return nil
			}
			if !llmAPITokenConfigured() {
				log.Println("reply 已启用，但未配置 llm.api_token，跳过回复")
				return nil
			}
			client, ok := clientFromEvent(clients, event)
			if !ok {
				return nil
			}

			reply := llm.Call(event.Message.Text)
			if strings.TrimSpace(reply) == "" {
				return nil
			}
			return client.SendGroupText(ctx, event.Message.Chat.ID, reply)
		}),
	}); err != nil {
		return err
	}

	return rt.Register(router.Route{
		Name: "reply_poke",
		Filter: base.EventFilter{
			Kind:    base.EventKindNotice,
			SubType: "poke",
		},
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			if event.Notice == nil || event.Notice.Chat.ID == "" {
				return nil
			}

			selfID := configuredSelfID()
			if selfID == "" {
				selfID = event.SelfID
			}
			if selfID == "" || event.Notice.Target.ID != selfID {
				return nil
			}
			client, ok := clientFromEvent(clients, event)
			if !ok {
				return nil
			}

			return client.SendPoke(ctx, event.Notice.Chat.ID, event.Notice.User.ID)
		}),
	})
}
