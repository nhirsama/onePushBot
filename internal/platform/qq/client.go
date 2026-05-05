package qq

import (
	"context"
	"encoding/json"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
)

// Client 定义 QQ 平台对外暴露的公共能力。
type Client interface {
	base.Client

	// Call 允许上层直接访问底层 OneBot action。
	Call(ctx context.Context, action string, params any) (json.RawMessage, error)

	SendGroupText(ctx context.Context, groupID string, text string) error
	SendPrivateText(ctx context.Context, userID string, text string) error
	SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResult, error)
	DeleteMessage(ctx context.Context, messageID string) error
	GetMessage(ctx context.Context, messageID string) (*base.Message, error)
	GetForwardMessage(ctx context.Context, messageID string) (json.RawMessage, error)
	SendForwardMessage(ctx context.Context, req SendForwardMessageRequest) (*SendMessageResult, error)
	SendGroupForwardMessage(ctx context.Context, groupID string, messages any) (*SendMessageResult, error)
	SendPrivateForwardMessage(ctx context.Context, userID string, messages any) (*SendMessageResult, error)

	GetLoginInfo(ctx context.Context) (*LoginInfo, error)
	GetBotStatus(ctx context.Context) (*BotStatus, error)
	GetVersionInfo(ctx context.Context) (*VersionInfo, error)
	CleanCache(ctx context.Context) error

	GetFriendList(ctx context.Context, noCache bool) ([]FriendInfo, error)
	GetGroupList(ctx context.Context, noCache bool) ([]GroupInfo, error)
	GetGroupInfo(ctx context.Context, groupID string, noCache bool) (*GroupInfo, error)
	SendLike(ctx context.Context, userID string, times int) error
	SendPoke(ctx context.Context, groupID string, userID string) error
	SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error
	GetGroupMemberInfo(ctx context.Context, groupID string, userID string, noCache bool) (*base.GroupMemberInfo, error)
	GetGroupMemberList(ctx context.Context, groupID string, noCache bool) ([]base.GroupMemberInfo, error)

	SetGroupKick(ctx context.Context, groupID string, userID string, rejectAddRequest bool) error
	SetGroupBan(ctx context.Context, groupID string, userID string, duration time.Duration) error
	SetGroupWholeBan(ctx context.Context, groupID string, enable bool) error
	SetGroupAdmin(ctx context.Context, groupID string, userID string, enable bool) error
	SetGroupCard(ctx context.Context, groupID string, userID string, card string) error
	SetGroupName(ctx context.Context, groupID string, name string) error
	SetGroupLeave(ctx context.Context, groupID string, dismiss bool) error
	SetGroupSpecialTitle(ctx context.Context, groupID string, userID string, title string, duration time.Duration) error

	SetFriendAddRequest(ctx context.Context, flag string, approve bool, remark string) error
	SetGroupAddRequest(ctx context.Context, flag string, subType string, approve bool, reason string) error

	GroupPoke(ctx context.Context, groupID string, userID string) error
	FriendPoke(ctx context.Context, userID string) error
	SetGroupSign(ctx context.Context, groupID string) error
	MarkPrivateMsgAsRead(ctx context.Context, userID string) error
	MarkGroupMsgAsRead(ctx context.Context, groupID string) error
}
