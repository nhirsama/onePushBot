package reply

import (
	"strings"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("reply.enabled")
}

func configuredSelfID() string {
	return strings.TrimSpace(viper.GetString("qq.self_id"))
}

func matchMentionedSelf() router.MatchFunc {
	return func(event base.Event) bool {
		selfID := configuredSelfID()
		if selfID == "" {
			selfID = event.SelfID
		}
		if selfID == "" {
			return false
		}
		return router.MatchMentionedUser(selfID)(event)
	}
}

func llmAPITokenConfigured() bool {
	return strings.TrimSpace(viper.GetString("llm.api_token")) != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
