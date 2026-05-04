package sudo

import (
	"context"
	"log"
	"strings"

	"github.com/nhirsama/onePushBot/config"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/nhirsama/onePushBot/pkg/auth"
	"github.com/nhirsama/onePushBot/pkg/llm"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
	"github.com/spf13/viper"
)

type authenticator interface {
	Authenticate(message []byte) (string, bool)
}

func Register(rt router.Router, clients platformclient.Source) error {
	if !Enabled() {
		return nil
	}

	authenticator := newAuthenticator()
	return rt.Register(router.Route{
		Name: "sudo",
		Filter: base.EventFilter{
			Kind:     base.EventKindMessage,
			ChatType: base.ChatTypeGroup,
		},
		Handler: router.HandlerFunc(func(ctx context.Context, event base.Event) error {
			if event.Message == nil || !strings.Contains(event.Message.RawText, "-----BEGIN PGP SIGNED MESSAGE-----") {
				return nil
			}

			message, ok := authenticator.Authenticate([]byte(event.Message.RawText))
			if !ok {
				return nil
			}
			if !llmAPIKeyConfigured() {
				log.Println("sudo 已启用，但未配置 llm.api_key/apiKey，跳过回复")
				return nil
			}
			client, ok := clientFromEvent(clients, event)
			if !ok {
				return nil
			}

			reply := llm.Call(message)
			if strings.TrimSpace(reply) == "" {
				return nil
			}
			return client.SendGroupText(ctx, event.Message.Chat.ID, reply)
		}),
	})
}

func newAuthenticator() authenticator {
	keys := publicKeys()
	if len(keys) == 0 {
		return config.Auth
	}

	authenticator := auth.NewAuth()
	for _, key := range keys {
		authenticator.AddPublicKey(key)
	}
	return authenticator
}

func publicKeys() []string {
	keys := append([]string{}, viper.GetStringSlice("sudo.public_keys")...)
	for _, key := range []string{
		viper.GetString("sudo.public_key"),
		viper.GetString("auth.public_key"),
	} {
		if strings.TrimSpace(key) != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

func llmAPIKeyConfigured() bool {
	return firstNonEmpty(
		viper.GetString("llm.api_key"),
		viper.GetString("api_key"),
		viper.GetString("apiKey"),
		config.ApiKey,
	) != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
