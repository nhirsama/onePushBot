package config

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nhirsama/onePushBot/internal/store"
	"github.com/spf13/viper"
)

var (
	consoleReader       = bufio.NewReader(os.Stdin)
	generatedAdminToken bool

	storeMu     sync.RWMutex
	activeStore store.Store
)

// OpenDefaultStore 打开默认 SQLite 存储并加载 settings 到运行时配置。
func OpenDefaultStore() error {
	kv, err := store.OpenSQLite(DBPath())
	if err != nil {
		return err
	}
	if err := UseStore(context.Background(), kv); err != nil {
		_ = kv.Close()
		return err
	}
	return nil
}

// UseStore 允许 cmd 或测试注入存储，config 包只保留运行时读取适配。
func UseStore(ctx context.Context, kv store.Store) error {
	storeMu.Lock()
	previous := activeStore
	activeStore = kv
	storeMu.Unlock()
	if previous != nil && previous != kv {
		_ = previous.Close()
	}

	setDefaults()
	if err := ensureAdminToken(ctx); err != nil {
		return err
	}
	if err := Reload(ctx); err != nil {
		return err
	}
	return nil
}

// Reload 从 KV 存储重新加载配置。
func Reload(ctx context.Context) error {
	kv := Store()
	if kv == nil {
		return fmt.Errorf("config store 未初始化")
	}

	resetViper()
	items, err := kv.List(ctx, "")
	if err != nil {
		return err
	}
	for _, item := range items {
		viper.Set(item.Key, decodeSetting(item.Value))
	}
	return nil
}

func Store() store.Store {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return activeStore
}

func CloseStore() error {
	storeMu.Lock()
	kv := activeStore
	activeStore = nil
	storeMu.Unlock()
	if kv == nil {
		return nil
	}
	return kv.Close()
}

func ensureAdminToken(ctx context.Context) error {
	kv := Store()
	if kv == nil {
		return fmt.Errorf("config store 未初始化")
	}
	token, ok, err := kv.Get(ctx, "admin.token")
	if err != nil {
		return err
	}
	if ok {
		if decoded, ok := decodeSetting(token).(string); ok && strings.TrimSpace(decoded) != "" {
			return nil
		}
	}
	generatedAdminToken = true
	return kv.Set(ctx, "admin.token", encodeSetting(randomToken()))
}

func randomToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(buf)
}

func setDefaults() {
	viper.SetDefault("admin.addr", "127.0.0.1:8090")
	viper.SetDefault("router.broker_buffer", 128)
	viper.SetDefault("qq.heartbeat_timeout", 90)
	viper.SetDefault("feishu.http_addr", ":8080")
	viper.SetDefault("feishu.webhook_path", "/feishu/events")
	viper.SetDefault("feishu.receive_id_type", "chat_id")
	viper.SetDefault("telegram_user.session_path", "./data/telegram_user.session")
	viper.SetDefault("telegram_user.auth_mode", "qr")
	viper.SetDefault("llm.base_url", "https://open.bigmodel.cn/api/paas/v4/chat/completions")
	viper.SetDefault("llm.model", "glm-4.5-flash")
	viper.SetDefault("riddle.http_addr", ":12396")
	viper.SetDefault("nowcoder_daily.hour", 18)
	viper.SetDefault("nowcoder_daily.minute", 0)
}

func resetViper() {
	viper.Reset()
	setDefaults()
}

func AdminTokenGenerated() bool {
	return generatedAdminToken
}

func readRuntimeString(ctx context.Context, label string) (string, error) {
	type result struct {
		value string
		err   error
	}

	ch := make(chan result, 1)
	go func() {
		fmt.Print(label)
		value, err := consoleReader.ReadString('\n')
		ch <- result{value: strings.TrimSpace(value), err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-ch:
		return result.value, result.err
	}
}
