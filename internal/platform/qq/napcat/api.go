package napcat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/qq"
)

var _ qq.Client = (*client)(nil)

func (c *client) SendGroupText(ctx context.Context, groupID string, text string) error {
	_, err := c.Call(ctx, "send_group_msg", map[string]any{
		"group_id": toNapcatID(groupID),
		"message": []map[string]any{
			{
				"type": "text",
				"data": map[string]any{"text": text},
			},
		},
	})
	return err
}

func (c *client) SendPrivateText(ctx context.Context, userID string, text string) error {
	_, err := c.Call(ctx, "send_private_msg", map[string]any{
		"user_id": toNapcatID(userID),
		"message": []map[string]any{
			{
				"type": "text",
				"data": map[string]any{"text": text},
			},
		},
	})
	return err
}

func (c *client) SendLike(ctx context.Context, userID string, times int) error {
	return c.callNoResult(ctx, "send_like", map[string]any{
		"user_id": toNapcatID(userID),
		"times":   times,
	})
}

func (c *client) SendPoke(ctx context.Context, groupID string, userID string) error {
	params := map[string]any{
		"group_id":  toNapcatID(groupID),
		"target_id": toNapcatID(userID),
	}
	if c.cfg.SelfID != "" {
		params["user_id"] = toNapcatID(c.cfg.SelfID)
	}
	return c.callNoResult(ctx, "send_poke", params)
}

func (c *client) SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error {
	return c.callNoResult(ctx, "set_msg_emoji_like", map[string]any{
		"message_id": messageID,
		"emoji_id":   emojiID,
		"set":        set,
	})
}

func (c *client) GetGroupMemberInfo(ctx context.Context, groupID string, userID string, noCache bool) (*base.GroupMemberInfo, error) {
	data, err := c.Call(ctx, "get_group_member_info", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
		"no_cache": noCache,
	})
	if err != nil {
		return nil, err
	}

	var member rawGroupMemberInfo
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, fmt.Errorf("解析群成员信息失败: %w", err)
	}
	return member.toDomain(), nil
}

func (c *client) GetLoginInfo(ctx context.Context) (*qq.LoginInfo, error) {
	var payload rawLoginInfo
	if err := c.callAndDecode(ctx, "get_login_info", nil, &payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

func (c *client) GetBotStatus(ctx context.Context) (*qq.BotStatus, error) {
	var payload qq.BotStatus
	if err := c.callAndDecode(ctx, "get_status", nil, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *client) GetVersionInfo(ctx context.Context) (*qq.VersionInfo, error) {
	var payload qq.VersionInfo
	if err := c.callAndDecode(ctx, "get_version_info", nil, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *client) CleanCache(ctx context.Context) error {
	return c.callNoResult(ctx, "clean_cache", nil)
}

func (c *client) SendMessage(ctx context.Context, req qq.SendMessageRequest) (*qq.SendMessageResult, error) {
	params := map[string]any{
		"message": req.Message,
	}
	if req.MessageType != "" {
		params["message_type"] = req.MessageType
	}
	if req.UserID != "" {
		params["user_id"] = toNapcatID(req.UserID)
	}
	if req.GroupID != "" {
		params["group_id"] = toNapcatID(req.GroupID)
	}
	if req.AutoEscape {
		params["auto_escape"] = true
	}

	var payload rawSendMessageResult
	if err := c.callAndDecode(ctx, "send_msg", params, &payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

func (c *client) DeleteMessage(ctx context.Context, messageID string) error {
	return c.callNoResult(ctx, "delete_msg", map[string]any{
		"message_id": toNapcatID(messageID),
	})
}

func (c *client) GetMessage(ctx context.Context, messageID string) (*base.Message, error) {
	var payload rawEnvelope
	if err := c.callAndDecode(ctx, "get_msg", map[string]any{
		"message_id": toNapcatID(messageID),
	}, &payload); err != nil {
		return nil, err
	}

	msg, err := mapMessage(payload, c.cfg.SelfID)
	if err != nil {
		return nil, fmt.Errorf("解析消息详情失败: %w", err)
	}
	return &msg, nil
}

func (c *client) GetForwardMessage(ctx context.Context, messageID string) (json.RawMessage, error) {
	return c.Call(ctx, "get_forward_msg", map[string]any{
		"id": messageID,
	})
}

func (c *client) GetFriendList(ctx context.Context, noCache bool) ([]qq.FriendInfo, error) {
	var payload []rawFriendInfo
	if err := c.callAndDecode(ctx, "get_friend_list", map[string]any{
		"no_cache": noCache,
	}, &payload); err != nil {
		return nil, err
	}

	items := make([]qq.FriendInfo, 0, len(payload))
	for _, item := range payload {
		items = append(items, item.toDomain())
	}
	return items, nil
}

func (c *client) GetGroupList(ctx context.Context, noCache bool) ([]qq.GroupInfo, error) {
	var payload []rawGroupInfo
	if err := c.callAndDecode(ctx, "get_group_list", map[string]any{
		"no_cache": noCache,
	}, &payload); err != nil {
		return nil, err
	}

	items := make([]qq.GroupInfo, 0, len(payload))
	for _, item := range payload {
		items = append(items, item.toDomain())
	}
	return items, nil
}

func (c *client) GetGroupInfo(ctx context.Context, groupID string, noCache bool) (*qq.GroupInfo, error) {
	var payload rawGroupInfo
	if err := c.callAndDecode(ctx, "get_group_info", map[string]any{
		"group_id": toNapcatID(groupID),
		"no_cache": noCache,
	}, &payload); err != nil {
		return nil, err
	}
	info := payload.toDomain()
	return &info, nil
}

func (c *client) GetGroupMemberList(ctx context.Context, groupID string, noCache bool) ([]base.GroupMemberInfo, error) {
	var payload []rawGroupMemberInfo
	if err := c.callAndDecode(ctx, "get_group_member_list", map[string]any{
		"group_id": toNapcatID(groupID),
		"no_cache": noCache,
	}, &payload); err != nil {
		return nil, err
	}

	items := make([]base.GroupMemberInfo, 0, len(payload))
	for _, item := range payload {
		items = append(items, *item.toDomain())
	}
	return items, nil
}

func (c *client) SetGroupKick(ctx context.Context, groupID string, userID string, rejectAddRequest bool) error {
	return c.callNoResult(ctx, "set_group_kick", map[string]any{
		"group_id":           toNapcatID(groupID),
		"user_id":            toNapcatID(userID),
		"reject_add_request": rejectAddRequest,
	})
}

func (c *client) SetGroupBan(ctx context.Context, groupID string, userID string, duration time.Duration) error {
	return c.callNoResult(ctx, "set_group_ban", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
		"duration": durationSeconds(duration),
	})
}

func (c *client) SetGroupWholeBan(ctx context.Context, groupID string, enable bool) error {
	return c.callNoResult(ctx, "set_group_whole_ban", map[string]any{
		"group_id": toNapcatID(groupID),
		"enable":   enable,
	})
}

func (c *client) SetGroupAdmin(ctx context.Context, groupID string, userID string, enable bool) error {
	return c.callNoResult(ctx, "set_group_admin", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
		"enable":   enable,
	})
}

func (c *client) SetGroupCard(ctx context.Context, groupID string, userID string, card string) error {
	return c.callNoResult(ctx, "set_group_card", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
		"card":     card,
	})
}

func (c *client) SetGroupName(ctx context.Context, groupID string, name string) error {
	return c.callNoResult(ctx, "set_group_name", map[string]any{
		"group_id":   toNapcatID(groupID),
		"group_name": name,
	})
}

func (c *client) SetGroupLeave(ctx context.Context, groupID string, dismiss bool) error {
	return c.callNoResult(ctx, "set_group_leave", map[string]any{
		"group_id":   toNapcatID(groupID),
		"is_dismiss": dismiss,
	})
}

func (c *client) SetGroupSpecialTitle(ctx context.Context, groupID string, userID string, title string, duration time.Duration) error {
	return c.callNoResult(ctx, "set_group_special_title", map[string]any{
		"group_id":      toNapcatID(groupID),
		"user_id":       toNapcatID(userID),
		"special_title": title,
		"duration":      durationSeconds(duration),
	})
}

func (c *client) SetFriendAddRequest(ctx context.Context, flag string, approve bool, remark string) error {
	return c.callNoResult(ctx, "set_friend_add_request", map[string]any{
		"flag":    flag,
		"approve": approve,
		"remark":  remark,
	})
}

func (c *client) SetGroupAddRequest(ctx context.Context, flag string, subType string, approve bool, reason string) error {
	params := map[string]any{
		"flag":    flag,
		"approve": approve,
		"reason":  reason,
	}
	if subType != "" {
		params["sub_type"] = subType
	}
	return c.callNoResult(ctx, "set_group_add_request", params)
}

func (c *client) GroupPoke(ctx context.Context, groupID string, userID string) error {
	return c.callNoResult(ctx, "group_poke", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
	})
}

func (c *client) FriendPoke(ctx context.Context, userID string) error {
	return c.callNoResult(ctx, "friend_poke", map[string]any{
		"user_id": toNapcatID(userID),
	})
}

func (c *client) SetGroupSign(ctx context.Context, groupID string) error {
	return c.callNoResult(ctx, "set_group_sign", map[string]any{
		"group_id": toNapcatID(groupID),
	})
}

func (c *client) MarkPrivateMsgAsRead(ctx context.Context, userID string) error {
	return c.callNoResult(ctx, "mark_private_msg_as_read", map[string]any{
		"user_id": toNapcatID(userID),
	})
}

func (c *client) MarkGroupMsgAsRead(ctx context.Context, groupID string) error {
	return c.callNoResult(ctx, "mark_group_msg_as_read", map[string]any{
		"group_id": toNapcatID(groupID),
	})
}

func (c *client) SendForwardMessage(ctx context.Context, req qq.SendForwardMessageRequest) (*qq.SendMessageResult, error) {
	params := map[string]any{
		"messages": req.Messages,
	}
	if req.MessageType != "" {
		params["message_type"] = req.MessageType
	}
	if req.UserID != "" {
		params["user_id"] = toNapcatID(req.UserID)
	}
	if req.GroupID != "" {
		params["group_id"] = toNapcatID(req.GroupID)
	}

	var payload rawSendMessageResult
	if err := c.callAndDecode(ctx, "send_forward_msg", params, &payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

func (c *client) SendGroupForwardMessage(ctx context.Context, groupID string, messages any) (*qq.SendMessageResult, error) {
	var payload rawSendMessageResult
	if err := c.callAndDecode(ctx, "send_group_forward_msg", map[string]any{
		"group_id": toNapcatID(groupID),
		"messages": messages,
	}, &payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

func (c *client) SendPrivateForwardMessage(ctx context.Context, userID string, messages any) (*qq.SendMessageResult, error) {
	var payload rawSendMessageResult
	if err := c.callAndDecode(ctx, "send_private_forward_msg", map[string]any{
		"user_id":  toNapcatID(userID),
		"messages": messages,
	}, &payload); err != nil {
		return nil, err
	}
	return payload.toDomain(), nil
}

func (c *client) callNoResult(ctx context.Context, action string, params any) error {
	_, err := c.Call(ctx, action, params)
	return err
}

func (c *client) callAndDecode(ctx context.Context, action string, params any, out any) error {
	data, err := c.Call(ctx, action, params)
	if err != nil {
		return err
	}
	if out == nil || len(data) == 0 || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("解析 napcat action %s 响应失败: %w", action, err)
	}
	return nil
}

func durationSeconds(value time.Duration) int64 {
	if value <= 0 {
		return 0
	}
	return int64(value / time.Second)
}

func toNapcatID(id string) any {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return id
	}
	if value, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return value
	}
	return id
}

func (m rawGroupMemberInfo) toDomain() *base.GroupMemberInfo {
	return &base.GroupMemberInfo{
		GroupID:         normalizeID(m.GroupID),
		UserID:          normalizeID(m.UserID),
		Nickname:        m.Nickname,
		Card:            m.Card,
		Sex:             m.Sex,
		Age:             m.Age,
		JoinTime:        m.JoinTime,
		LastSentTime:    m.LastSentTime,
		Level:           m.Level,
		QQLevel:         m.QQLevel,
		Role:            m.Role,
		Title:           m.Title,
		Area:            m.Area,
		Unfriendly:      m.Unfriendly,
		TitleExpireTime: m.TitleExpireTime,
		CardChangeable:  m.CardChangeable,
		ShutUpTimestamp: m.ShutUpTimestamp,
		IsRobot:         m.IsRobot,
		QAge:            m.QAge,
	}
}
