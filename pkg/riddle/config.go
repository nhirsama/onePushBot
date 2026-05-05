package riddle

import (
	"strings"

	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("riddle.enabled")
}

func groupID() string {
	return firstNonEmpty(viper.GetString("riddle.group_id"))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
