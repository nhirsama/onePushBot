package riddle

import (
	"strings"

	"github.com/spf13/viper"
)

func groupID() string {
	return strings.TrimSpace(viper.GetString("riddle.group_id"))
}
