package reply

import (
	"strconv"
	"strings"

	"github.com/nhirsama/onePushBot/config"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/router"
	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("reply.Enable") || viper.GetBool("reply.enable")
}

func configuredSelfID() string {
	selfID := firstNonEmpty(viper.GetString("bot_config.self_id"), viper.GetString("selfId"))
	if selfID != "" {
		return selfID
	}
	if config.SelfId != 0 {
		return strconv.FormatInt(config.SelfId, 10)
	}
	return ""
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
