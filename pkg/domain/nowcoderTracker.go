package domain

import "encoding/json"

type Response struct {
	Msg  string          `json:"msg"`  // 返回消息
	Code int64           `json:"code"` // 状态码
	Data json.RawMessage `json:"data"` // 实际数据
}

type TrackerTodayInfo struct {
	QuestionId    int64  `json:"questionId"`
	QuestionTitle string `json:"questionTitle"`
	QuestionUrl   string `json:"questionUrl"`
	ProblemId     int64  `json:"problemId"`
}

type TrackerMemberInfo struct {
	Total int64 `json:"total"` // 总人数
	List  []struct {
		ContinueDays int64  `json:"continueDays"` // 连续打卡天数
		Count        int64  `json:"count"`        // 总打卡次数
		Name         string `json:"name"`         // 用户名
		HeadUrl      string `json:"headUrl"`      // 用户头像 URL
		Rank         int64  `json:"rank"`         // 排名
		UserId       int64  `json:"userId"`       // 用户 ID
		CheckedToday bool   `json:"checkedToday"` // 今日是否已打卡
	} `json:"list"` // 列表数据
}

type RankInfo struct {
	TotalCount int `json:"totalCount"` // 排行人数
	Ranks      []struct {
		Uid          int    `json:"uid"`                    // 用户 ID
		HeadUrl      string `json:"headUrl"`                // 用户头像 URL
		Name         string `json:"name"`                   // 用户名
		Count        int    `json:"count"`                  // 数量（过题数或打卡次数）
		Place        int    `json:"place"`                  // 排名
		ContinueDays int    `json:"continueDays,omitempty"` // 连续打卡天数（仅打卡榜有）
	} `json:"ranks"` // 排行榜条目

}
type GetTrackerInfo interface {
	GetTodayInfo() (*TrackerTodayInfo, error)
	GetGroupMemberInfo(groupId int64) (*TrackerMemberInfo, error)
	GetProblemRankInfo(userId int64) (*RankInfo, error)
	GetCheckinRankInfo(userId int64) (*RankInfo, error)
}
