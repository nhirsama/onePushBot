package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLite, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path 不能为空")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &SQLite{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS runtime_state (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	platform TEXT NOT NULL,
	chat_id TEXT NOT NULL,
	sender_id TEXT,
	text TEXT,
	raw TEXT,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS summaries (
	id TEXT PRIMARY KEY,
	chat_id TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	return err
}

func (s *SQLite) Get(ctx context.Context, key string) (string, bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (s *SQLite) Set(ctx context.Context, key string, value string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO settings(key, value, updated_at)
VALUES(?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET
	value = excluded.value,
	updated_at = CURRENT_TIMESTAMP`, key, value)
	return err
}

func (s *SQLite) List(ctx context.Context, prefix string) ([]KV, error) {
	query := `SELECT key, value FROM settings ORDER BY key`
	args := []any(nil)
	if prefix != "" {
		query = `SELECT key, value FROM settings WHERE key LIKE ? ORDER BY key`
		args = append(args, prefix+"%")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []KV
	for rows.Next() {
		var item KV
		if err := rows.Scan(&item.Key, &item.Value); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
