package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	tguser "github.com/nhirsama/onePushBot/internal/platform/telegram/user"
	"rsc.io/qr"
)

type consoleTelegramAuth struct{}

// NewTelegramAuthProvider 返回 Telegram 用户态运行时认证输入源。
func NewTelegramAuthProvider() tguser.AuthProvider {
	return consoleTelegramAuth{}
}

// Phone 从控制台读取 Telegram 用户态首次授权手机号。
func (consoleTelegramAuth) Phone(ctx context.Context) (string, error) {
	return readRuntimeString(ctx, "请输入 Telegram 手机号: ")
}

// Code 从控制台读取 Telegram 用户态登录验证码。
func (consoleTelegramAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	_ = sentCode
	return readRuntimeString(ctx, "请输入 Telegram 验证码: ")
}

// Password 从控制台读取 Telegram 二步验证密码。
func (consoleTelegramAuth) Password(ctx context.Context) (string, error) {
	return readRuntimeString(ctx, "请输入 Telegram 二步验证密码: ")
}

// QRCode 在控制台输出扫码登录链接；token 只用于本次登录，不写入配置文件。
func (consoleTelegramAuth) QRCode(ctx context.Context, token tguser.QRLoginToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	code, err := qr.Encode(token.URL, qr.M)
	if err != nil {
		return err
	}
	fmt.Println("请使用 Telegram 扫描二维码登录:")
	printQRCode(code)
	fmt.Printf("如果终端无法扫码，打开链接: %s\n", token.URL)
	if !token.ExpiresAt.IsZero() {
		fmt.Printf("二维码过期时间: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func printQRCode(code *qr.Code) {
	const quietZone = 2
	for y := -quietZone; y < code.Size+quietZone; y += 2 {
		var line strings.Builder
		for x := -quietZone; x < code.Size+quietZone; x++ {
			top := qrBlack(code, x, y)
			bottom := qrBlack(code, x, y+1)
			switch {
			case top && bottom:
				line.WriteString("█")
			case top:
				line.WriteString("▀")
			case bottom:
				line.WriteString("▄")
			default:
				line.WriteString(" ")
			}
		}
		fmt.Println(line.String())
	}
}

func qrBlack(code *qr.Code, x, y int) bool {
	if x < 0 || y < 0 || x >= code.Size || y >= code.Size {
		return false
	}
	return code.Black(x, y)
}
