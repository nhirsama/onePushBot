package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// --- 模拟环境结构 (假设这些结构已存在于您的项目中) ---

// APIResponse 模拟基础的 API 响应结构
type APIResponse struct {
	Status  string          `json:"status"`
	Retcode int             `json:"retcode"`
	Data    json.RawMessage `json:"data"` // 使用 RawMessage 延迟解析数据
	Message string          `json:"message"`
	Wording string          `json:"wording"`
	Echo    *string         `json:"echo,omitempty"`
}

// --- /get_group_member_info 接口定义 ---

// GetGroupMemberInfoParams 定义了请求 /get_group_member_info API 所需的参数
type GetGroupMemberInfoParams struct {
	GroupID int64 `json:"group_id"` // 群号
	UserID  int64 `json:"user_id"`  // 成员 QQ 号
	NoCache bool  `json:"no_cache"` // 是否不使用缓存
}

// GroupMemberInfoData 对应于 OpenAPI 规范中的 "群成员信息" schema
// 包含 API 响应中 data 字段的详细结构
type GroupMemberInfoData struct {
	GroupID         int64  `json:"group_id"`
	UserID          int64  `json:"user_id"`
	Nickname        string `json:"nickname"`
	Card            string `json:"card"`              // 群昵称
	Sex             string `json:"sex"`               // 性别
	Age             int    `json:"age"`               // 年龄
	JoinTime        int64  `json:"join_time"`         // 加群时间戳
	LastSentTime    int64  `json:"last_sent_time"`    // 最后发言时间戳
	Level           string `json:"level"`             // 群等级
	QQLevel         int    `json:"qq_level"`          // 账号等级
	Role            string `json:"role"`              // 权限 (owner/admin/member)
	Title           string `json:"title"`             // 头衔
	Area            string `json:"area"`              // 地区
	Unfriendly      bool   `json:"unfriendly"`        // 是否不良记录成员
	TitleExpireTime int64  `json:"title_expire_time"` // 头衔过期时间戳
	CardChangeable  bool   `json:"card_changeable"`   // 群昵称是否可修改
	ShutUpTimestamp int64  `json:"shut_up_timestamp"` // 禁言时间戳
	IsRobot         bool   `json:"is_robot"`          // 是否机器人
	QAge            string `json:"qage"`              // Q龄
}

// GetGroupMemberInfoResult 是该 API 调用的最终响应结构
type GetGroupMemberInfoResult struct {
	APIResponse
	// Data 字段在结构体切片后会被单独解析
	Data GroupMemberInfoData `json:"data"`
}

// GetGroupMemberInfo 调用 /get_group_member_info API 获取单个群成员的详细信息
// 参数:
//   - groupID: 群号
//   - userID: 成员 QQ 号
//   - noCache: 是否不使用缓存
//
// 返回:
//   - *GroupMemberInfoData: 成员信息数据，如果成功
//   - error: 错误信息
func (w *WebSocketMessage) GetGroupMemberInfo(groupID, userID int64, noCache bool) (*GroupMemberInfoData, error) {
	// 1. 构造请求参数结构体
	params := GetGroupMemberInfoParams{
		GroupID: groupID,
		UserID:  userID,
		NoCache: noCache,
	}

	// 2. 调用底层的 callAPI 方法
	resp, err := w.callAPI("get_group_member_info", params)
	if err != nil {
		log.Println("callAPI get_group_member_info 错误:", err)
		return nil, err
	}

	// 3. 检查 API 状态
	if resp.Status != "ok" {
		log.Printf("请求 get_group_member_info API 失败, 状态: %s, RetCode: %d", resp.Status, resp.RetCode)
		return nil, fmt.Errorf("API 调用失败: %s, retcode: %d", resp.Status, resp.RetCode)
	}

	// 4. 解析 data 字段 (这是关键步骤，需要将 json.RawMessage 再次解析到目标结构体)
	var memberInfo GroupMemberInfoData
	if err := json.Unmarshal(resp.Data, &memberInfo); err != nil {
		log.Printf("解析群成员信息数据失败: %v", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	log.Println("请求 get_group_member_info API 成功")
	return &memberInfo, nil
}
