package riddle

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("riddle.Enable") || viper.GetBool("riddle.enable")
}

func groupID() string {
	return firstNonEmpty(viper.GetString("riddle.group_id"), viper.GetString("riddleConfigGroupId"))
}

func listenAddr() string {
	addr := firstNonEmpty(viper.GetString("riddle.http_addr"), viper.GetString("riddle.addr"))
	if addr == "" {
		if port := viper.GetInt("riddle.port"); port > 0 {
			addr = strconv.Itoa(port)
		}
	}
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
