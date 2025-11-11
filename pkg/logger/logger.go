package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/nhirsama/onePushBot/pkg/domain"
)

type slogLogger struct {
	mu     sync.RWMutex
	logger *slog.Logger
	level  slog.Level
}

func NewLog(level domain.LogLevel) domain.Log {
	l := &slogLogger{}
	l.SetLevel(level)
	return l
}

func (l *slogLogger) SetLevel(level domain.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var slogLevel slog.Level
	switch level {
	case domain.DebugLevel:
		slogLevel = slog.LevelDebug
	case domain.InfoLevel:
		slogLevel = slog.LevelInfo
	case domain.WarnLevel:
		slogLevel = slog.LevelWarn
	case domain.ErrorLevel:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	l.level = slogLevel
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	l.logger = slog.New(handler)
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Error(msg, args...)
}
