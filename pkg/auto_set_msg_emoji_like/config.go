package autoSetMsgEmojiLike

import (
	"strconv"
	"strings"

	"github.com/nhirsama/onePushBot/config"
	"github.com/spf13/viper"
)

func Enabled() bool {
	return viper.GetBool("autoSetMsgEmojiLike.Enable") || viper.GetBool("autoSetMsgEmojiLike.enable")
}

func emojiLikesForUser(userID string) []int {
	settings := emojiLikeSettings()
	if len(settings) == 0 {
		return nil
	}
	return settings[userID]
}

func emojiLikeSettings() map[string][]int {
	result := make(map[string][]int)
	for _, key := range []string{
		"autoSetMsgEmojiLike.set",
		"autoSetMsgEmojiLikeSet",
		"auto_set_msg_emoji_like.set",
	} {
		mergeEmojiSettings(result, viper.GetStringMap(key))
	}
	for userID, emojiIDs := range config.AutoSetMsgEmojiLikeSet {
		result[userID] = append(result[userID], emojiIDs...)
	}
	return result
}

func mergeEmojiSettings(result map[string][]int, raw map[string]any) {
	for userID, value := range raw {
		result[userID] = append(result[userID], intsFromAny(value)...)
	}
}

func intsFromAny(value any) []int {
	switch typed := value.(type) {
	case []int:
		return typed
	case []any:
		items := make([]int, 0, len(typed))
		for _, item := range typed {
			if value, ok := intFromAny(item); ok {
				items = append(items, value)
			}
		}
		return items
	case []string:
		items := make([]int, 0, len(typed))
		for _, item := range typed {
			if value, err := strconv.Atoi(strings.TrimSpace(item)); err == nil {
				items = append(items, value)
			}
		}
		return items
	default:
		if value, ok := intFromAny(typed); ok {
			return []int{value}
		}
		return nil
	}
}

func intFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return 0, false
	}
}
