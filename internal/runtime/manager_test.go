package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/nhirsama/onePushBot/internal/admin"
	base "github.com/nhirsama/onePushBot/internal/platform"
)

type fakeApp struct {
	mu         sync.Mutex
	startCalls int
	closeCalls int
	statuses   map[string]string
	startErr   error
	closeErr   error
}

func (a *fakeApp) Start(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.startCalls++
	return a.startErr
}

func (a *fakeApp) Close(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closeCalls++
	return a.closeErr
}

func (a *fakeApp) PlatformStatuses() map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make(map[string]string, len(a.statuses))
	for k, v := range a.statuses {
		result[k] = v
	}
	return result
}

func TestManagerStartStatusAndClose(t *testing.T) {
	oldSnapshot := snapshotConfig
	oldReload := reloadConfig
	oldUpdate := updateConfig
	t.Cleanup(func() {
		snapshotConfig = oldSnapshot
		reloadConfig = oldReload
		updateConfig = oldUpdate
	})
	snapshotConfig = func(bool) map[string]any { return map[string]any{"admin": map[string]any{"addr": "127.0.0.1:8090"}} }
	reloadConfig = func(context.Context) error { return nil }
	updateConfig = func(context.Context, map[string]any) error { return nil }

	fake := &fakeApp{statuses: map[string]string{"qq": string(base.StatusRunning)}}
	var factoryCalls int
	manager := NewManagerWithFactory(func(controller admin.Controller, logger base.Logger) (appRunner, error) {
		factoryCalls++
		if controller == nil {
			t.Fatal("expected controller to be passed to app factory")
		}
		if logger == nil {
			t.Fatal("expected logger to be passed to app factory")
		}
		return fake, nil
	})

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if factoryCalls != 1 {
		t.Fatalf("unexpected factory calls: %d", factoryCalls)
	}
	if fake.startCalls != 1 {
		t.Fatalf("unexpected start calls: %d", fake.startCalls)
	}

	status, err := manager.Status(context.Background())
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !status.Running {
		t.Fatal("expected running status")
	}
	if status.Platforms["qq"] != string(base.StatusRunning) {
		t.Fatalf("unexpected platform status: %v", status.Platforms)
	}
	if status.Config["admin"].(map[string]any)["addr"] != "127.0.0.1:8090" {
		t.Fatalf("unexpected config snapshot: %v", status.Config)
	}

	if err := manager.Close(context.Background()); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if fake.closeCalls != 1 {
		t.Fatalf("unexpected close calls: %d", fake.closeCalls)
	}

	status, err = manager.Status(context.Background())
	if err != nil {
		t.Fatalf("status after close failed: %v", err)
	}
	if status.Running {
		t.Fatal("expected stopped status after close")
	}
}

func TestManagerRestartRebuildsApp(t *testing.T) {
	oldSnapshot := snapshotConfig
	oldReload := reloadConfig
	oldUpdate := updateConfig
	t.Cleanup(func() {
		snapshotConfig = oldSnapshot
		reloadConfig = oldReload
		updateConfig = oldUpdate
	})
	snapshotConfig = func(bool) map[string]any { return map[string]any{} }
	reloadConfig = func(context.Context) error { return nil }
	updateConfig = func(context.Context, map[string]any) error { return nil }

	first := &fakeApp{statuses: map[string]string{"first": "running"}}
	second := &fakeApp{statuses: map[string]string{"second": "running"}}
	var factoryCalls int
	manager := NewManagerWithFactory(func(admin.Controller, base.Logger) (appRunner, error) {
		factoryCalls++
		switch factoryCalls {
		case 1:
			return first, nil
		case 2:
			return second, nil
		default:
			t.Fatalf("unexpected factory call: %d", factoryCalls)
			return nil, nil
		}
	})

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := manager.Restart(context.Background()); err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	if first.closeCalls != 1 {
		t.Fatalf("expected first app to close once, got %d", first.closeCalls)
	}
	if second.startCalls != 1 {
		t.Fatalf("expected second app to start once, got %d", second.startCalls)
	}
	status, err := manager.Status(context.Background())
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if status.Platforms["second"] != "running" {
		t.Fatalf("unexpected status after restart: %v", status.Platforms)
	}
}

func TestManagerReloadUpdateAndLogs(t *testing.T) {
	oldSnapshot := snapshotConfig
	oldReload := reloadConfig
	oldUpdate := updateConfig
	t.Cleanup(func() {
		snapshotConfig = oldSnapshot
		reloadConfig = oldReload
		updateConfig = oldUpdate
	})
	snapshotConfig = func(bool) map[string]any { return map[string]any{"ok": true} }

	var reloadCalls int
	reloadConfig = func(context.Context) error {
		reloadCalls++
		return nil
	}
	var updateCalls int
	updateConfig = func(context.Context, map[string]any) error {
		updateCalls++
		return nil
	}

	manager := NewManagerWithFactory(func(admin.Controller, base.Logger) (appRunner, error) {
		return &fakeApp{statuses: map[string]string{}}, nil
	})

	if _, err := manager.Config(context.Background()); err != nil {
		t.Fatalf("config failed: %v", err)
	}
	if err := manager.Reload(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if err := manager.UpdateConfig(context.Background(), map[string]any{"ok": true}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if reloadCalls != 1 || updateCalls != 1 {
		t.Fatalf("unexpected hook calls: reload=%d update=%d", reloadCalls, updateCalls)
	}
}

func TestManagerStartFailsWithoutFactoryApp(t *testing.T) {
	manager := NewManagerWithFactory(func(admin.Controller, base.Logger) (appRunner, error) {
		return nil, errors.New("boom")
	})
	if err := manager.Start(context.Background()); err == nil {
		t.Fatal("expected start failure")
	}
}
