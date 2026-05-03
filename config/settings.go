package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const RedactedValue = "******"

// Snapshot 返回当前配置快照。redact 为 true 时隐藏 token/secret/hash/key 等敏感字段。
func Snapshot(redact bool) map[string]any {
	settings := cloneMap(viper.AllSettings())
	if redact {
		redactMap(settings)
	}
	return settings
}

// UpdateSettings 合并 Web 管理面板提交的配置并写入 KV 存储。
func UpdateSettings(ctx context.Context, patch map[string]any) error {
	kv := Store()
	if kv == nil {
		return fmt.Errorf("config store 未初始化")
	}
	flattened := make(map[string]any)
	flatten("", patch, flattened)
	for key, value := range flattened {
		if text, ok := value.(string); ok && text == RedactedValue {
			continue
		}
		if key == "admin.token" && strings.TrimSpace(fmt.Sprint(value)) == "" {
			continue
		}
		if err := kv.Set(ctx, key, encodeSetting(value)); err != nil {
			return err
		}
	}
	return Reload(ctx)
}

func encodeSetting(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}

func decodeSetting(value string) any {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return value
	}
	return decoded
}

func flatten(prefix string, values map[string]any, output map[string]any) {
	for key, value := range values {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if nested, ok := value.(map[string]any); ok {
			flatten(path, nested, output)
			continue
		}
		output[path] = value
	}
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		if nested, ok := value.(map[string]any); ok {
			output[key] = cloneMap(nested)
			continue
		}
		output[key] = value
	}
	return output
}

func redactMap(values map[string]any) {
	for key, value := range values {
		if nested, ok := value.(map[string]any); ok {
			redactMap(nested)
			continue
		}
		if sensitiveKey(key) && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
			values[key] = RedactedValue
		}
	}
}

func sensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	return strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "hash") ||
		strings.Contains(normalized, "key")
}
