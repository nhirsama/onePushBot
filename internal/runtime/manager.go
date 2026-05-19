package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/internal/admin"
	"github.com/nhirsama/onePushBot/internal/app"
	base "github.com/nhirsama/onePushBot/internal/platform"
)

var (
	snapshotConfig = config.Snapshot
	updateConfig   = config.UpdateSettings
	reloadConfig   = config.Reload
)

type appRunner interface {
	Start(context.Context) error
	Close(context.Context) error
	PlatformStatuses() map[string]string
}

type AppFactory func(controller admin.Controller, logger base.Logger) (appRunner, error)

type Manager struct {
	mu      sync.Mutex
	app     appRunner
	ctx     context.Context
	cancel  context.CancelFunc
	since   time.Time
	factory AppFactory
	logger  base.Logger
}

func NewManager() *Manager {
	return NewManagerWithFactory(nil)
}

func NewManagerWithFactory(factory AppFactory) *Manager {
	if factory == nil {
		factory = defaultAppFactory
	}
	return &Manager{
		factory: factory,
		logger:  newLogger(),
	}
}

func defaultAppFactory(controller admin.Controller, logger base.Logger) (appRunner, error) {
	return app.New(controller, logger)
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.startLocked(ctx)
}

func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeLocked(ctx)
}

func (m *Manager) Status(ctx context.Context) (admin.Status, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	status := admin.Status{
		Running:   m.app != nil,
		Platforms: make(map[string]string),
		Process:   processStatus(m.since),
		Config:    snapshotConfig(true),
	}
	if m.app == nil {
		return status, nil
	}
	for name, state := range m.app.PlatformStatuses() {
		status.Platforms[name] = state
	}
	return status, nil
}

func (m *Manager) Config(ctx context.Context) (map[string]any, error) {
	_ = ctx
	return snapshotConfig(true), nil
}

func (m *Manager) Logs(ctx context.Context, since uint64, limit int) ([]admin.LogEntry, error) {
	_ = ctx
	items := logBuffer.List(since, limit)
	result := make([]admin.LogEntry, 0, len(items))
	for _, item := range items {
		result = append(result, admin.LogEntry{
			ID:      item.ID,
			Time:    item.Time.Format("2006-01-02 15:04:05"),
			Level:   item.Level,
			Message: item.Message,
		})
	}
	return result, nil
}

func (m *Manager) UpdateConfig(ctx context.Context, patch map[string]any) error {
	return updateConfig(ctx, patch)
}

func (m *Manager) Reload(ctx context.Context) error {
	return reloadConfig(ctx)
}

func (m *Manager) Restart(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.closeLocked(ctx); err != nil {
		return err
	}
	if err := reloadConfig(ctx); err != nil {
		return err
	}
	return m.startLocked(context.Background())
}

func (m *Manager) startLocked(parent context.Context) error {
	if m.app != nil {
		return nil
	}
	appRunner, err := m.factory(m, m.logger)
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(parent)
	if err := appRunner.Start(runCtx); err != nil {
		cancel()
		return err
	}
	m.app = appRunner
	m.ctx = runCtx
	m.cancel = cancel
	m.since = time.Now()
	return nil
}

func (m *Manager) closeLocked(ctx context.Context) error {
	if m.app == nil {
		return nil
	}
	if m.cancel != nil {
		m.cancel()
	}
	closeCtx := ctx
	if closeCtx == nil {
		var cancel context.CancelFunc
		closeCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	err := m.app.Close(closeCtx)
	m.app = nil
	m.ctx = nil
	m.cancel = nil
	m.since = time.Time{}
	if err != nil {
		return fmt.Errorf("关闭程序失败: %w", err)
	}
	return nil
}
