package sudo

import "github.com/spf13/viper"

func Enabled() bool {
	return viper.GetBool("sudo.enabled")
}
