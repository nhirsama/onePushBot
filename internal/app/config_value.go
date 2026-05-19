package app

import (
	"time"

	"github.com/spf13/viper"
)

func secondsToDuration(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func durationFromConfig(key string) time.Duration {
	if value := viper.GetDuration(key); value > 0 {
		return value
	}
	return secondsToDuration(viper.GetInt64(key))
}
