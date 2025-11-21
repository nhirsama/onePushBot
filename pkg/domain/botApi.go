package domain

import (
	"context"
	"encoding/json"
	"time"
)

//
// 基础类型 & 通用结构
//

// MessageID 消息唯一标识（对应 NapCat / OneBot 的 message_id）
type MessageID int64

// UserID QQ 号
type UserID int64

// GroupID 群号
type GroupID int64

// GuildID 频道 ID（NapCat 扩展）
type GuildID string

// MessageSegmentType 消息段类型，对应 NapCat / OneBot 的 type 字段
type MessageSegmentType string

const (
	MessageSegmentText   MessageSegmentType = "text"
	MessageSegmentAt     MessageSegmentType = "at"
	MessageSegmentFace   MessageSegmentType = "face"
	MessageSegmentImage  MessageSegmentType = "image"
	MessageSegmentReply  MessageSegmentType = "reply"
	MessageSegmentJSON   MessageSegmentType = "json"
	MessageSegmentVoice  MessageSegmentType = "record"
	MessageSegmentVideo  MessageSegmentType = "video"
	MessageSegmentMusic  MessageSegmentType = "music"
	MessageSegmentNode   MessageSegmentType = "node"   // 合并转发用
	MessageSegmentPoke   MessageSegmentType = "poke"   // 戳一戳
	MessageSegmentCustom MessageSegmentType = "custom" // 其他扩展
)

// MessageSegment 抽象一段消息，对应 NapCat message 数组里的每个元素
type MessageSegment struct {
	Type MessageSegmentType `json:"type"` // 消息段类型

	// 文本消息
	Text string `json:"text,omitempty"`

	// 艾特消息
	QQ string `json:"qq,omitempty"` // "all" 表示 @全体成员

	// 图片 / 语音 / 视频 / 文件 等资源标识
	File string `json:"file,omitempty"` // 本地路径 / 网络 URL / NapCat 支持的其他形式

	// 回复消息
	ID MessageID `json:"id,omitempty"` // 被回复的消息 ID（reply 段）

	// JSON 卡片
	Data json.RawMessage `json:"data,omitempty"` // JSON 卡片原始数据

	// 扩展数据，兼容未列举字段
	Extra map[string]any `json:"extra,omitempty"`
}

// Message 一条完整消息抽象，用于历史消息/详情等
type Message struct {
	MessageID MessageID        `json:"message_id"`
	RealID    MessageID        `json:"real_id,omitempty"` // 某些实现会返回 real_id
	GroupID   GroupID          `json:"group_id,omitempty"`
	UserID    UserID           `json:"user_id,omitempty"` // 发送者 QQ
	Time      int64            `json:"time"`              // unix 时间戳
	Raw       json.RawMessage  `json:"raw,omitempty"`     // 原始 data（可选）
	Segments  []MessageSegment `json:"segments,omitempty"`
}

// SendMessageResult 发送消息返回结果，对应 send_group_msg / send_private_msg 的 data
type SendMessageResult struct {
	MessageID MessageID `json:"message_id"` // 新消息 ID
}

// Friend 好友信息，对应 get_friend_list 的单项数据
type Friend struct {
	UserID   UserID `json:"user_id"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
	// 可能还有分组信息等，可按实际数据追加字段
}

// StrangerInfo 获取账号信息 / 陌生人信息，对应 get_stranger_info 的 data 示例
type StrangerInfo struct {
	UserID     UserID `json:"user_id"`
	UID        string `json:"uid"`
	UIN        string `json:"uin"`
	Nickname   string `json:"nickname"`
	Age        int    `json:"age"`
	QID        string `json:"qid"`
	QQLevel    int    `json:"qqLevel"`
	Sex        string `json:"sex"`
	LongNick   string `json:"long_nick"`
	RegTime    int64  `json:"reg_time"`
	IsVIP      bool   `json:"is_vip"`
	IsYearsVIP bool   `json:"is_years_vip"`
	VIPLevel   int    `json:"vip_level"`
	Remark     string `json:"remark"`
	Status     int    `json:"status"`
	LoginDays  int    `json:"login_days"`
}

// LoginInfo 登录号信息，对应 get_login_info
type LoginInfo struct {
	UserID   UserID `json:"user_id"`
	Nickname string `json:"nickname"`
}

// StatusInfo 在线状态信息，对应 get_status
type StatusInfo struct {
	Online bool   `json:"online"`
	Status string `json:"status"` // 具体含义参考 NapCat 文档
}

// Group 群基础信息，对应 get_group_list / get_group_info
type Group struct {
	GroupID GroupID `json:"group_id"`
	Name    string  `json:"group_name"`
	// 其他字段如 group_memo 等可按实际数据追加
}

// GroupMember 群成员信息，对应 get_group_member_info / get_group_member_list
type GroupMember struct {
	GroupID  GroupID `json:"group_id"`
	UserID   UserID  `json:"user_id"`
	Nickname string  `json:"nickname"`
	Card     string  `json:"card"` // 群名片
	Role     string  `json:"role"` // owner / admin / member
}

// GroupHonor 群荣誉信息，对应 get_group_honor_info
type GroupHonor struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw"` // 具体结构比较复杂，用 raw 兜底
}

// EssenceMessage 群精华消息，对应 get_essence_msg_list 的单项
type EssenceMessage struct {
	SenderID   UserID    `json:"sender_id"`
	SenderNick string    `json:"sender_nick"`
	MessageID  MessageID `json:"message_id"`
	Time       int64     `json:"time"`
	OperatorID UserID    `json:"operator_id"`
}

// FileInfo 文件信息，对应 get_file / get_group_root_files / get_group_files_by_folder
type FileInfo struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	BusID       int    `json:"busid,omitempty"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"url,omitempty"`
}

// OnlineClient 在线客户端信息，对应 get_online_clients
type OnlineClient struct {
	Platform string `json:"platform"`
	Device   string `json:"device"`
	Status   string `json:"status"`
}

// VersionInfo 版本信息，对应 get_version_info
type VersionInfo struct {
	AppName    string `json:"app_name,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
	Protocol   string `json:"protocol_version,omitempty"`
}

// AICharacter AI 语音人物，对应 get_ai_characters
type AICharacter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecentContact 最近会话信息，对应 get_recent_contact 的单项
type RecentContact struct {
	Type    string   `json:"type"` // private / group 等
	UserID  *UserID  `json:"user_id,omitempty"`
	GroupID *GroupID `json:"group_id,omitempty"`
	Time    int64    `json:"time"`
}

//
// 总聚合接口：方便其他组件只依赖一个 NapCatAPI
//

// NapCatAPI 聚合 NapCat 所有常用能力的公共接口
type NapCatAPI interface {
	AccountAPI
	FriendAPI
	GroupAPI
	MessageAPI
	FileAPI
	AIAPI
	PersonalAPI
	SystemAPI
}

//
// 账号相关（对应 napcat.wiki 账号相关 + 部分其他功能）
//

// AccountAPI 账号相关接口（登录信息 / 在线状态 / 自身资料等）
type AccountAPI interface {
	// GetLoginInfo 获取登录号信息，对应 NapCat 接口字符串 get_login_info
	GetLoginInfo(ctx context.Context) (*LoginInfo, error)

	// GetStatus 获取机器人在线状态，对应 NapCat 接口字符串 get_status
	GetStatus(ctx context.Context) (*StatusInfo, error)

	// GetStrangerInfo 获取账号信息 / 陌生人信息，对应 NapCat 接口字符串 get_stranger_info
	GetStrangerInfo(ctx context.Context, userID UserID) (*StrangerInfo, error)

	// SetSelfLongNick 设置个性签名，对应 NapCat 接口字符串 set_self_longnick
	SetSelfLongNick(ctx context.Context, longNick string) error

	// SetOnlineStatus 设置在线状态，对应 NapCat 接口字符串 set_online_status
	SetOnlineStatus(ctx context.Context, status string) error

	// SetDIYOnlineStatus 设置自定义在线状态，对应 NapCat 接口字符串 set_diy_online_status
	SetDIYOnlineStatus(ctx context.Context, faceID, faceType int, wording string) error

	// SetQQAvatar 设置 QQ 头像，对应 NapCat 接口字符串 set_qq_avatar
	SetQQAvatar(ctx context.Context, file string) error

	// CleanCache 清除缓存，对应 NapCat 接口字符串 clean_cache
	CleanCache(ctx context.Context) error

	// GetOnlineClients 获取当前账号在线客户端列表，对应 NapCat 接口字符串 get_online_clients
	GetOnlineClients(ctx context.Context) ([]OnlineClient, error)
}

//
// 好友相关
//

// FriendAPI 好友相关接口（好友列表 / 历史消息 / 点赞等等）
type FriendAPI interface {
	// GetFriendList 获取好友列表，对应 NapCat 接口字符串 get_friend_list
	GetFriendList(ctx context.Context, noCache bool) ([]Friend, error)

	// GetUnidirectionalFriendList 获取单向好友列表，对应 NapCat 接口字符串 get_unidirectional_friend_list
	GetUnidirectionalFriendList(ctx context.Context) ([]Friend, error)

	// SendLike 对好友进行点赞，对应 NapCat 接口字符串 send_like
	SendLike(ctx context.Context, userID UserID, times int) error

	// SetFriendRemark 设置好友备注，对应 NapCat 接口字符串 set_friend_remark
	SetFriendRemark(ctx context.Context, userID UserID, remark string) error

	// DeleteFriend 删除好友，对应 NapCat 接口字符串 delete_friend
	DeleteFriend(ctx context.Context, userID UserID) error

	// FriendPoke 私聊戳一戳，对应 NapCat 扩展接口 friend_poke
	FriendPoke(ctx context.Context, userID UserID) error

	// MarkPrivateMsgAsRead 标记私聊消息为已读，对应 NapCat 接口字符串 mark_private_msg_as_read
	MarkPrivateMsgAsRead(ctx context.Context, userID UserID, t time.Time) error

	// GetFriendMsgHistory 获取好友历史消息，对应 NapCat 接口字符串 get_friend_msg_history
	GetFriendMsgHistory(ctx context.Context, userID UserID, count int) ([]Message, error)

	// ForwardFriendSingleMsg 转发单条好友消息，对应 NapCat 接口字符串 forward_friend_single_msg
	ForwardFriendSingleMsg(ctx context.Context, targetUserID UserID, srcMessageID MessageID) (*SendMessageResult, error)
}

//
// 群聊相关
//

// GroupAPI 群聊相关接口（群列表 / 成员 / 管理 / 禁言 / 精华 / 荣誉等）
type GroupAPI interface {
	// GetGroupList 获取群列表，对应 NapCat 接口字符串 get_group_list
	GetGroupList(ctx context.Context, noCache bool) ([]Group, error)

	// GetGroupInfo 获取群基础信息，对应 NapCat 接口字符串 get_group_info
	GetGroupInfo(ctx context.Context, groupID GroupID, noCache bool) (*Group, error)

	// GetGroupInfoEx 获取群扩展信息，对应 NapCat 接口字符串 get_group_info_ex
	GetGroupInfoEx(ctx context.Context, groupID GroupID) (map[string]any, error)

	// GetGroupMemberInfo 获取群成员信息，对应 NapCat 接口字符串 get_group_member_info
	GetGroupMemberInfo(ctx context.Context, groupID GroupID, userID UserID, noCache bool) (*GroupMember, error)

	// GetGroupMemberList 获取群成员列表，对应 NapCat 接口字符串 get_group_member_list
	GetGroupMemberList(ctx context.Context, groupID GroupID, noCache bool) ([]GroupMember, error)

	// GetGroupHonorInfo 获取群荣誉信息，对应 NapCat 接口字符串 get_group_honor_info
	GetGroupHonorInfo(ctx context.Context, groupID GroupID, honorType string) (*GroupHonor, error)

	// GetEssenceMsgList 获取群精华消息列表，对应 NapCat 接口字符串 get_essence_msg_list
	GetEssenceMsgList(ctx context.Context, groupID GroupID) ([]EssenceMessage, error)

	// SetEssenceMsg 设置精华消息，对应 NapCat 接口字符串 set_essence_msg
	SetEssenceMsg(ctx context.Context, messageID MessageID) error

	// DeleteEssenceMsg 删除精华消息，对应 NapCat 接口字符串 delete_essence_msg
	DeleteEssenceMsg(ctx context.Context, messageID MessageID) error

	// SetGroupAddRequest 处理加群请求，对应 NapCat 接口字符串 set_group_add_request
	SetGroupAddRequest(ctx context.Context, flag string, approve bool, reason string) error

	// SetGroupKick 群踢人，对应 NapCat 接口字符串 set_group_kick
	SetGroupKick(ctx context.Context, groupID GroupID, userID UserID, rejectAddRequest bool) error

	// SetGroupBan 设置群成员禁言，对应 NapCat 接口字符串 set_group_ban
	SetGroupBan(ctx context.Context, groupID GroupID, userID UserID, duration time.Duration) error

	// SetGroupWholeBan 设置全员禁言，对应 NapCat 接口字符串 set_group_whole_ban
	SetGroupWholeBan(ctx context.Context, groupID GroupID, enable bool) error

	// SetGroupAdmin 设置群管理员，对应 NapCat 接口字符串 set_group_admin
	SetGroupAdmin(ctx context.Context, groupID GroupID, userID UserID, enable bool) error

	// SetGroupCard 设置群名片，对应 NapCat 接口字符串 set_group_card
	SetGroupCard(ctx context.Context, groupID GroupID, userID UserID, card string) error

	// SetGroupName 设置群名称，对应 NapCat 接口字符串 set_group_name
	SetGroupName(ctx context.Context, groupID GroupID, name string) error

	// SetGroupLeave 退出群聊，对应 NapCat 接口字符串 set_group_leave
	SetGroupLeave(ctx context.Context, groupID GroupID, isDismiss bool) error

	// SetGroupSpecialTitle 设置群专属头衔，对应 NapCat 接口字符串 set_group_special_title
	SetGroupSpecialTitle(ctx context.Context, groupID GroupID, userID UserID, title string) error

	// SetGroupPortrait 设置群头像，对应 NapCat 接口字符串 set_group_portrait
	SetGroupPortrait(ctx context.Context, groupID GroupID, file string, cache int) error

	// SendGroupNotice 发送群公告，对应 NapCat 扩展接口 _send_group_notice
	SendGroupNotice(ctx context.Context, groupID GroupID, content string, extra map[string]any) error

	// GetGroupNotice 获取群公告，对应 NapCat 扩展接口 _get_group_notice
	GetGroupNotice(ctx context.Context, groupID GroupID) (json.RawMessage, error)

	// DelGroupNotice 删除群公告，对应 NapCat 扩展接口 _del_group_notice
	DelGroupNotice(ctx context.Context, groupID GroupID, noticeID string) error

	// GetGroupAtAllRemain 获取 @全体成员 剩余次数，对应 NapCat 接口字符串 get_group_at_all_remain
	GetGroupAtAllRemain(ctx context.Context, groupID GroupID) (int, error)

	// GetGroupSystemMsg 获取群系统消息，对应 NapCat 接口字符串 get_group_system_msg
	GetGroupSystemMsg(ctx context.Context) (json.RawMessage, error)

	// GetGroupShutList 获取群禁言列表，对应 NapCat 接口字符串 get_group_shut_list
	GetGroupShutList(ctx context.Context, groupID GroupID) ([]GroupMember, error)

	// SetGroupRemark 设置群备注，对应 NapCat 接口字符串 set_group_remark
	SetGroupRemark(ctx context.Context, groupID GroupID, remark string) error

	// SetGroupSign 群签到（原生扩展），对应 NapCat 扩展接口 set_group_sign
	SetGroupSign(ctx context.Context, groupID GroupID) error

	// SendGroupSign 发送群签到（扩展），对应 NapCat 扩展接口 send_group_sign
	SendGroupSign(ctx context.Context, groupID GroupID) error

	// GroupPoke 群内戳一戳，对应 NapCat 扩展接口 group_poke
	GroupPoke(ctx context.Context, groupID GroupID, userID UserID) error
}

//
// 消息相关
//

// MessageAPI 消息相关接口（群聊 / 私聊 / 合并转发 / OCR / 历史消息等）
type MessageAPI interface {
	// SendGroupText 发送群文本消息，对应 NapCat 接口字符串 send_group_msg（text 段）
	SendGroupText(ctx context.Context, groupID GroupID, text string) (*SendMessageResult, error)

	// SendGroupAtText 发送群艾特消息，对应 NapCat 接口字符串 send_group_msg（at 段）
	SendGroupAtText(ctx context.Context, groupID GroupID, qq string, text string) (*SendMessageResult, error)

	// SendGroupImage 发送群图片消息，对应 NapCat 接口字符串 send_group_msg（image 段）
	SendGroupImage(ctx context.Context, groupID GroupID, file string) (*SendMessageResult, error)

	// SendGroupJSON 发送群 JSON 卡片，对应 NapCat 接口字符串 send_group_msg（json 段）
	SendGroupJSON(ctx context.Context, groupID GroupID, card json.RawMessage) (*SendMessageResult, error)

	// SendGroupVoice 发送群语音消息，对应 NapCat 接口字符串 send_group_msg（record 段）
	SendGroupVoice(ctx context.Context, groupID GroupID, file string) (*SendMessageResult, error)

	// SendGroupVideo 发送群视频消息，对应 NapCat 接口字符串 send_group_msg（video 段）
	SendGroupVideo(ctx context.Context, groupID GroupID, file string) (*SendMessageResult, error)

	// SendGroupReply 发送群回复消息，对应 NapCat 接口字符串 send_group_msg（reply 段）
	SendGroupReply(ctx context.Context, groupID GroupID, replyTo MessageID, segments []MessageSegment) (*SendMessageResult, error)

	// SendGroupMusicCard 发送群音乐卡片，对应 NapCat 扩展接口（send_group_msg + music 段）
	SendGroupMusicCard(ctx context.Context, groupID GroupID, musicJSON json.RawMessage) (*SendMessageResult, error)

	// SendGroupCustomMusicCard 发送群自定义音乐卡片，对应 NapCat 扩展接口 send_group_msg（自定义 music 段）
	SendGroupCustomMusicCard(ctx context.Context, groupID GroupID, musicJSON json.RawMessage) (*SendMessageResult, error)

	// SendGroupDice 发送群聊超级表情骰子，对应 NapCat 扩展接口（send_group_msg 特定段）
	SendGroupDice(ctx context.Context, groupID GroupID) (*SendMessageResult, error)

	// SendGroupRPS 发送群聊超级表情猜拳，对应 NapCat 扩展接口（send_group_msg 特定段）
	SendGroupRPS(ctx context.Context, groupID GroupID) (*SendMessageResult, error)

	// SendGroupPoke 发送群聊戳一戳，对应 NapCat 接口字符串 send_group_msg / group_poke
	SendGroupPoke(ctx context.Context, groupID GroupID, userID UserID) error

	// SendGroupForwardMsg 发送群合并转发消息，对应 NapCat 接口字符串 send_group_forward_msg
	SendGroupForwardMsg(ctx context.Context, groupID GroupID, nodes []Message) (*SendMessageResult, error)

	// SendPrivateText 发送私聊文本消息，对应 NapCat 接口字符串 send_private_msg（text 段）
	SendPrivateText(ctx context.Context, userID UserID, text string) (*SendMessageResult, error)

	// SendPrivateImage 发送私聊图片消息，对应 NapCat 接口字符串 send_private_msg（image 段）
	SendPrivateImage(ctx context.Context, userID UserID, file string) (*SendMessageResult, error)

	// SendPrivateJSON 发送私聊 JSON 卡片，对应 NapCat 接口字符串 send_private_msg（json 段）
	SendPrivateJSON(ctx context.Context, userID UserID, card json.RawMessage) (*SendMessageResult, error)

	// SendPrivateVoice 发送私聊语音消息，对应 NapCat 接口字符串 send_private_msg（record 段）
	SendPrivateVoice(ctx context.Context, userID UserID, file string) (*SendMessageResult, error)

	// SendPrivateVideo 发送私聊视频消息，对应 NapCat 接口字符串 send_private_msg（video 段）
	SendPrivateVideo(ctx context.Context, userID UserID, file string) (*SendMessageResult, error)

	// SendPrivateReply 发送私聊回复消息，对应 NapCat 接口字符串 send_private_msg（reply 段）
	SendPrivateReply(ctx context.Context, userID UserID, replyTo MessageID, segments []MessageSegment) (*SendMessageResult, error)

	// SendPrivateMusicCard 发送私聊音乐卡片，对应 NapCat 扩展接口
	SendPrivateMusicCard(ctx context.Context, userID UserID, musicJSON json.RawMessage) (*SendMessageResult, error)

	// SendPrivateCustomMusicCard 发送私聊自定义音乐卡片，对应 NapCat 扩展接口
	SendPrivateCustomMusicCard(ctx context.Context, userID UserID, musicJSON json.RawMessage) (*SendMessageResult, error)

	// SendPrivateDice 发送私聊超级表情骰子，对应 NapCat 扩展接口
	SendPrivateDice(ctx context.Context, userID UserID) (*SendMessageResult, error)

	// SendPrivateRPS 发送私聊超级表情猜拳，对应 NapCat 扩展接口
	SendPrivateRPS(ctx context.Context, userID UserID) (*SendMessageResult, error)

	// SendPrivatePoke 发送私聊戳一戳，对应 NapCat 接口字符串 send_poke / friend_poke
	SendPrivatePoke(ctx context.Context, userID UserID) error

	// SendMsg 通用发送消息接口，对应 NapCat 接口字符串 send_msg
	SendMsg(ctx context.Context, messageType string, userID *UserID, groupID *GroupID, segments []MessageSegment) (*SendMessageResult, error)

	// DeleteMsg 撤回消息，对应 NapCat 接口字符串 delete_msg / 撤回消息
	DeleteMsg(ctx context.Context, messageID MessageID) error

	// GetMsg 获取消息详情，对应 NapCat 接口字符串 get_msg / 获取消息详情
	GetMsg(ctx context.Context, messageID MessageID) (*Message, error)

	// GetGroupMsgHistory 获取群历史消息，对应 NapCat 接口字符串 get_group_msg_history
	GetGroupMsgHistory(ctx context.Context, groupID GroupID, count int) ([]Message, error)

	// GetForwardMsg 获取合并转发消息节点列表，对应 NapCat 接口字符串 get_forward_msg
	GetForwardMsg(ctx context.Context, messageID MessageID) ([]Message, error)

	// SendForwardMsg 发送合并转发消息（自动判断上下文），对应 NapCat 接口字符串 send_forward_msg
	SendForwardMsg(ctx context.Context, nodes []Message) (*SendMessageResult, error)

	// SendPrivateForwardMsg 发送私聊合并转发消息，对应 NapCat 接口字符串 send_private_forward_msg
	SendPrivateForwardMsg(ctx context.Context, userID UserID, nodes []Message) (*SendMessageResult, error)

	// SendPoke 上下文相关的戳一戳，对应 NapCat 接口字符串 send_poke
	SendPoke(ctx context.Context, params map[string]any) error

	// MarkMsgAsRead 标记消息已读，对应 NapCat 接口字符串 mark_msg_as_read
	MarkMsgAsRead(ctx context.Context, params map[string]any) error

	// MarkAllAsRead 标记所有消息为已读，对应 NapCat 扩展接口 _mark_all_as_read
	MarkAllAsRead(ctx context.Context) error

	// GetRecentContact 获取最近联系人列表，对应 NapCat 接口字符串 get_recent_contact
	GetRecentContact(ctx context.Context, count int) ([]RecentContact, error)

	// OcrImage 图片 OCR 识别（标准），对应 NapCat 接口字符串 ocr_image
	OcrImage(ctx context.Context, image string) (string, error)

	// OcrImagePro 图片 OCR 识别（增强），对应 NapCat 扩展接口 .ocr_image
	OcrImagePro(ctx context.Context, image string) (string, error)

	// GetRecord 获取语音文件信息，对应 NapCat 接口字符串 get_record
	GetRecord(ctx context.Context, file, outFormat string) (json.RawMessage, error)

	// GetImage 获取图片文件信息，对应 NapCat 接口字符串 get_image
	GetImage(ctx context.Context, file string) (json.RawMessage, error)
}

//
// 文件相关
//

// FileAPI 文件相关接口（群文件 / 私聊文件 / 下载 / 链接）
type FileAPI interface {
	// UploadGroupFile 上传群文件，对应 NapCat 接口字符串 upload_group_file
	UploadGroupFile(ctx context.Context, groupID GroupID, filePath, name, folder string) (*FileInfo, error)

	// DeleteGroupFile 删除群文件，对应 NapCat 接口字符串 delete_group_file
	DeleteGroupFile(ctx context.Context, groupID GroupID, fileID string, busID int) error

	// CreateGroupFileFolder 创建群文件文件夹，对应 NapCat 接口字符串 create_group_file_folder
	CreateGroupFileFolder(ctx context.Context, groupID GroupID, name string) (string, error)

	// DeleteGroupFolder 删除群文件夹，对应 NapCat 接口字符串 delete_group_folder
	DeleteGroupFolder(ctx context.Context, groupID GroupID, folderID string) error

	// GetGroupFileSystemInfo 获取群文件系统信息，对应 NapCat 接口字符串 get_group_file_system_info
	GetGroupFileSystemInfo(ctx context.Context, groupID GroupID) (json.RawMessage, error)

	// GetGroupRootFiles 获取群根目录文件列表，对应 NapCat 接口字符串 get_group_root_files
	GetGroupRootFiles(ctx context.Context, groupID GroupID) ([]FileInfo, error)

	// GetGroupFilesByFolder 获取群子目录文件列表，对应 NapCat 接口字符串 get_group_files_by_folder
	GetGroupFilesByFolder(ctx context.Context, groupID GroupID, folderID string) ([]FileInfo, error)

	// GetGroupFileURL 获取群文件链接，对应 NapCat 接口字符串 get_group_file_url
	GetGroupFileURL(ctx context.Context, groupID GroupID, fileID string, busID int) (string, error)

	// MoveGroupFile 移动群文件，对应 NapCat 接口字符串 move_group_file
	MoveGroupFile(ctx context.Context, groupID GroupID, fileID, targetDir string) error

	// RenameGroupFile 重命名群文件，对应 NapCat 接口字符串 rename_group_file
	RenameGroupFile(ctx context.Context, groupID GroupID, fileID, currentDir, newName string) error

	// UploadPrivateFile 上传私聊文件，对应 NapCat 接口字符串 upload_private_file
	UploadPrivateFile(ctx context.Context, userID UserID, filePath, name string) (*FileInfo, error)

	// GetPrivateFileURL 获取私聊文件链接，对应 NapCat 接口字符串 get_private_file_url
	GetPrivateFileURL(ctx context.Context, userID UserID, fileID string) (string, error)

	// DownloadFile 下载文件到缓存，对应 NapCat 接口字符串 download_file
	DownloadFile(ctx context.Context, url string, threadCount int, headers []string) (string, error)

	// GetFile 获取文件信息，对应 NapCat 接口字符串 get_file
	GetFile(ctx context.Context, file, fileType string) (*FileInfo, error)

	// CleanCache 清空缓存目录，对应 NapCat 扩展接口 clean_cache / clear_cache
	CleanCache(ctx context.Context) error
}

//
// AI / 翻译 / 个人操作
//

// AIAPI AI 相关接口（AI 角色 / 语音）
type AIAPI interface {
	// GetAICharacters 获取 AI 角色列表，对应 NapCat 接口字符串 get_ai_characters
	GetAICharacters(ctx context.Context) ([]AICharacter, error)

	// GetAIRecord 获取 AI 语音文件，对应 NapCat 接口字符串 get_ai_record
	GetAIRecord(ctx context.Context, characterID string, groupID GroupID, text string) (*FileInfo, error)

	// SendGroupAIRecord 发送群 AI 语音，对应 NapCat 接口字符串 send_group_ai_record
	SendGroupAIRecord(ctx context.Context, characterID string, groupID GroupID, text string) (*SendMessageResult, error)
}

// PersonalAPI 其他个人操作（翻译 / 输入状态 / 图片发送检查等）
type PersonalAPI interface {
	// TranslateEn2Zh 英文翻译成中文，对应 NapCat 接口字符串 translate_en2zh
	TranslateEn2Zh(ctx context.Context, text string) (string, error)

	// SetInputStatus 设置输入状态，对应 NapCat 接口字符串 set_input_status
	SetInputStatus(ctx context.Context, userID UserID, eventType int) error

	// CanSendImage 检查是否可以发送图片，对应 NapCat 接口字符串 can_send_image
	CanSendImage(ctx context.Context) (bool, error)

	// CanSendRecord 检查是否可以发送语音，对应 NapCat 接口字符串 can_send_record
	CanSendRecord(ctx context.Context) (bool, error)

	// HandleQuickOperation 对事件执行快速操作，对应 NapCat 扩展接口 .handle_quick_operation
	HandleQuickOperation(ctx context.Context, contextObj, operation map[string]any) error

	// ClickInlineKeyboardButton 点击内联键盘按钮，对应 NapCat 接口字符串 click_inline_keyboard_button
	ClickInlineKeyboardButton(ctx context.Context, groupID GroupID, botAppID, buttonID, callbackData, msgSeq string) error
}

//
// 系统 / 密钥 / 其他
//

// SystemAPI 系统及密钥相关接口（版本信息 / cookies / rkey 等）
type SystemAPI interface {
	// GetVersionInfo 获取 NapCat 版本信息，对应 NapCat 接口字符串 get_version_info
	GetVersionInfo(ctx context.Context) (*VersionInfo, error)

	// BotExit 退出机器人，对应 NapCat 接口字符串 bot_exit
	BotExit(ctx context.Context) error

	// GetClientKey 获取 clientkey，对应 NapCat 接口字符串 get_clientkey
	GetClientKey(ctx context.Context) (string, error)

	// GetCookies 获取 cookies，对应 NapCat 接口字符串 get_cookies
	GetCookies(ctx context.Context, domain string) (map[string]string, error)

	// GetCSRFToken 获取 CSRF Token，对应 NapCat 接口字符串 get_csrf_token
	GetCSRFToken(ctx context.Context) (string, error)

	// GetCredentials 获取 QQ 相关接口凭证，对应 NapCat 接口字符串 get_credentials
	GetCredentials(ctx context.Context, domain string) (map[string]any, error)

	// GetRKey 获取 RKey，对应 NapCat 接口字符串 get_rkey
	GetRKey(ctx context.Context) (string, error)

	// GetRobotUinRange 获取机器人 UIN 范围，对应 NapCat 接口字符串 get_robot_uin_range
	GetRobotUinRange(ctx context.Context) ([]UserID, error)

	// GetGuildList 获取频道列表，对应 NapCat 接口字符串 get_guild_list
	GetGuildList(ctx context.Context) ([]GuildID, error)

	// GetGuildServiceProfile 获取频道资料，对应 NapCat 接口字符串 get_guild_service_profile
	GetGuildServiceProfile(ctx context.Context) (json.RawMessage, error)

	// CheckURLSafely 检查链接安全性，对应 NapCat 接口字符串 check_url_safely
	CheckURLSafely(ctx context.Context, url string) (bool, error)

	// GetPacketStatus 获取数据包状态，对应 NapCat 接口字符串 nc_get_packet_status
	GetPacketStatus(ctx context.Context) (json.RawMessage, error)

	// SendPacket 发送自定义数据包，对应 NapCat 接口字符串 send_packet
	SendPacket(ctx context.Context, packet map[string]any) error
}
