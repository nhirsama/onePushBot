package domain

import (
	"context"
)

type Bot interface {
	// SendGroupMsg 发送群消息
	SendGroupMsg(ctx context.Context, groupID int64, text string, msgType string) error
	// SendPrivateMsg 发送私聊消息
	SendPrivateMsg(ctx context.Context, userID int64, text string, msgType string) error
	// SendPoke 发送戳一戳
	SendPoke(ctx context.Context, groupID, targetID int64) error
	// SetMsgEmojiLike 设置消息表情回应
	SetMsgEmojiLike(ctx context.Context, messageID int64, emojiID int, set bool) error
	// SendLike 发送点赞
	SendLike(ctx context.Context, userID, times int) error
}
