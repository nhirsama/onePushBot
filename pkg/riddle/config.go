package riddle

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("riddle.enabled")
}

func groupID() string {
	return firstNonEmpty(viper.GetString("riddle.group_id"))
}

func listenAddr() string {
	addr := firstNonEmpty(viper.GetString("riddle.http_addr"))
	if addr == "" {
		return ":12396"
	}
	if _, err := strconv.Atoi(addr); err == nil {
		return ":" + addr
	}
	return addr
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
