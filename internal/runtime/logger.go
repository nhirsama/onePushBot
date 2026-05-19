package runtime

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/nhirsama/onePushBot/internal/logs"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/spf13/viper"
)

var logBuffer = logs.NewRing(2000)

type platformLogger struct {
	mu     sync.RWMutex
	logger *slog.Logger
}

func newLogger() base.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(viper.GetString("log.level"))) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return &platformLogger{
		logger: slog.New(slog.NewTextHandler(logWriter(), &slog.HandlerOptions{
			Level: level,
		})),
	}
}

func logWriter() io.Writer {
	if logBuffer == nil {
		return os.Stdout
	}
	return io.MultiWriter(os.Stdout, logBuffer)
}

func LogBuffer() io.Writer {
	return logBuffer
}

func (l *platformLogger) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Debug(msg, args...)
}

func (l *platformLogger) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Info(msg, args...)
}

func (l *platformLogger) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Warn(msg, args...)
}

func (l *platformLogger) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Error(msg, args...)
}
