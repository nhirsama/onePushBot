package nowcoderTracker

import (
	"fmt"

	"github.com/nhirsama/onePushBot/pkg/nowcoder_tracker/nowcoder_tracker_api"
)

func PushTodayProblem() (string, error) {
	T, err := nowcoderTrackerApi.NewTracker().GetTodayInfo()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("『每日一题』\n今日题目:%s\n快来挑战自己，点击链接参与: %s", T.QuestionTitle, T.QuestionUrl), nil

}
