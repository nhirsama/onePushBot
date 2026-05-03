package logs

import (
	"bytes"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	ID      uint64    `json:"id"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type Ring struct {
	mu      sync.RWMutex
	entries []Entry
	next    int
	size    int
	lastID  uint64
	partial []byte
}

func NewRing(size int) *Ring {
	if size <= 0 {
		size = 1000
	}
	return &Ring{
		entries: make([]Entry, 0, size),
		size:    size,
	}
}

func (r *Ring) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	written := len(data)
	r.partial = append(r.partial, data...)
	for {
		index := bytes.IndexByte(r.partial, '\n')
		if index < 0 {
			break
		}
		line := strings.TrimSpace(string(r.partial[:index]))
		r.partial = r.partial[index+1:]
		if line != "" {
			r.appendLocked(line)
		}
	}
	return written, nil
}

func (r *Ring) List(since uint64, limit int) []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := r.orderedLocked()
	if since > 0 {
		filtered := entries[:0]
		for _, entry := range entries {
			if entry.ID > since {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	result := make([]Entry, len(entries))
	copy(result, entries)
	return result
}

func (r *Ring) appendLocked(message string) {
	r.lastID++
	entry := Entry{
		ID:      r.lastID,
		Time:    time.Now(),
		Level:   detectLevel(message),
		Message: message,
	}
	if len(r.entries) < r.size {
		r.entries = append(r.entries, entry)
		return
	}
	r.entries[r.next] = entry
	r.next = (r.next + 1) % r.size
}

func (r *Ring) orderedLocked() []Entry {
	if len(r.entries) < r.size || r.next == 0 {
		result := make([]Entry, len(r.entries))
		copy(result, r.entries)
		return result
	}
	result := make([]Entry, 0, len(r.entries))
	result = append(result, r.entries[r.next:]...)
	result = append(result, r.entries[:r.next]...)
	return result
}

func detectLevel(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "level=error"), strings.Contains(lower, "error"), strings.Contains(message, "失败"), strings.Contains(message, "错误"):
		return "error"
	case strings.Contains(lower, "level=warn"), strings.Contains(lower, "warn"), strings.Contains(message, "警告"):
		return "warn"
	case strings.Contains(lower, "level=debug"), strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}
