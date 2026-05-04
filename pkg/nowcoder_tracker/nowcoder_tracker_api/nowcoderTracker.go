package nowcoderTrackerApi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nhirsama/onePushBot/pkg/domain"
)

type Tracker struct {
	*http.Client
}

func NewTracker() *Tracker {
	return &Tracker{
		&http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *Tracker) api(api string) (*domain.Response, error) {
	resp, err := t.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var r domain.Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	if r.Code != 0 {
		return nil, fmt.Errorf("API 返回错误: %s", r.Msg)
	}

	return &r, nil
}

func (t *Tracker) GetTodayInfo() (*domain.TrackerTodayInfo, error) {
	api := "https://www.nowcoder.com/problem/tracker/clock/todayinfo"
	resp, err := t.api(api)
	if err != nil {
		return nil, fmt.Errorf("获取今日一题失败: %s", err)
	}
	var info domain.TrackerTodayInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, fmt.Errorf("json解析失败: %s", err)
	}
	info.QuestionUrl = "https://www.nowcoder.com" + info.QuestionUrl
	return &info, nil
}
func (t *Tracker) GetGroupMemberInfo(groupId int64) (*domain.TrackerMemberInfo, error) {
	api := fmt.Sprintf(
		"https://www.nowcoder.com/problem/tracker/team/leaderboard/clock?teamId=%d",
		groupId,
	)

	resp, err := t.api(api)
	if err != nil {
		return nil, fmt.Errorf("获取团队成员失败: %s", err)
	}

	var info domain.TrackerMemberInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, fmt.Errorf("json解析失败: %s", err)
	}

	return &info, nil
}
func (t *Tracker) GetProblemRankInfo(userId int64) (*domain.RankInfo, error) {
	api := fmt.Sprintf(
		"https://www.nowcoder.com/problem/tracker/ranks/problem?userId=%d",
		userId,
	)

	resp, err := t.api(api)
	if err != nil {
		return nil, fmt.Errorf("获取过题排名失败: %s", err)
	}

	var info domain.RankInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, fmt.Errorf("json解析失败: %s", err)
	}

	return &info, nil
}
func (t *Tracker) GetCheckinRankInfo(userId int64) (*domain.RankInfo, error) {
	api := fmt.Sprintf(
		"https://www.nowcoder.com/problem/tracker/ranks/checkin?userId=%d",
		userId,
	)

	resp, err := t.api(api)
	if err != nil {
		return nil, fmt.Errorf("获取打卡排名失败: %s", err)
	}

	var info domain.RankInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, fmt.Errorf("json解析失败: %s", err)
	}

	return &info, nil
}
