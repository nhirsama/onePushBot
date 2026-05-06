package nowcoderTracker

import (
	"context"
	"log"
	"strings"
	"time"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/plat"
	"github.com/spf13/viper"
)

type Client interface {
	SendGroupText(ctx context.Context, groupID string, text string) error
}

func StartDaily(ctx context.Context, clients plat.Clients) {
	groupIDs := dailyGroupIDs()
	if len(groupIDs) == 0 {
		log.Println("nowcoder_daily 已启用，但 nowcoder_daily.groups 为空，跳过")
		return
	}

	hour := viper.GetInt("nowcoder_daily.hour")
	if hour == 0 && !viper.IsSet("nowcoder_daily.hour") {
		hour = 18
	}
	minute := viper.GetInt("nowcoder_daily.minute")

	go runDaily(ctx, clients, groupIDs, hour, minute)
}

func clientFromSource(source plat.Clients) (Client, bool) {
	if platform := dailyPlatform(); platform != "" {
		return clientFromPlatform(source, platform)
	}
	if client, ok := clientFromPlatform(source, base.PlatformQQ); ok {
		return client, true
	}
	if client, ok := clientFromPlatform(source, base.PlatformTelegramBot); ok {
		return client, true
	}
	if client, ok := clientFromPlatform(source, base.PlatformTelegramUser); ok {
		return client, true
	}
	if client, ok := clientFromPlatform(source, base.PlatformFeishu); ok {
		return client, true
	}
	return nil, false
}

func dailyGroupIDs() []string {
	var result []string
	seen := map[string]struct{}{}
	for _, groupID := range viper.GetStringSlice("nowcoder_daily.groups") {
		addUniqueString(&result, seen, groupID)
	}
	return result
}

func dailyPlatform() base.Platform {
	return base.Platform(strings.TrimSpace(viper.GetString("nowcoder_daily.platform")))
}

func addUniqueString(result *[]string, seen map[string]struct{}, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if _, ok := seen[value]; ok {
		return
	}
	seen[value] = struct{}{}
	*result = append(*result, value)
}

func runDaily(ctx context.Context, clients plat.Clients, groupIDs []string, hour int, minute int) {
	for {
		next := nextDailyRun(time.Now(), hour, minute)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		text, err := PushTodayProblem()
		if err != nil {
			log.Printf("获取牛客每日一题失败: %v", err)
			continue
		}
		client, ok := clientFromSource(clients)
		if !ok {
			continue
		}
		for _, groupID := range groupIDs {
			if err := client.SendGroupText(ctx, groupID, text); err != nil {
				log.Printf("发送牛客每日一题失败: group=%s err=%v", groupID, err)
			}
		}
	}
}

func clientFromPlatform(clients plat.Clients, platform base.Platform) (Client, bool) {
	switch platform {
	case base.PlatformQQ:
		if clients.QQ == nil {
			return nil, false
		}
		client, ok := clients.QQ.(Client)
		return client, ok
	case base.PlatformFeishu:
		if clients.Feishu == nil {
			return nil, false
		}
		client, ok := clients.Feishu.(Client)
		return client, ok
	case base.PlatformTelegramBot:
		if clients.Telegram == nil || clients.Telegram.Bot == nil {
			return nil, false
		}
		client, ok := clients.Telegram.Bot.(Client)
		return client, ok
	case base.PlatformTelegramUser:
		if clients.Telegram == nil || clients.Telegram.User == nil {
			return nil, false
		}
		client, ok := clients.Telegram.User.(Client)
		return client, ok
	default:
		return nil, false
	}
}

func nextDailyRun(now time.Time, hour int, minute int) time.Time {
	if hour < 0 || hour > 23 {
		hour = 18
	}
	if minute < 0 || minute > 59 {
		minute = 0
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
