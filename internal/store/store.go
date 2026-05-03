package store

import "context"

type KV struct {
	Key   string
	Value string
}

// Store 收口本地持久化能力，上层不直接依赖 SQLite 细节。
type Store interface {
	Close() error
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string) error
	List(ctx context.Context, prefix string) ([]KV, error)
}
