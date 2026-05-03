package cmd

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/nhirsama/onePushBot/config"
	"github.com/nhirsama/onePushBot/internal/admin"
	"github.com/nhirsama/onePushBot/internal/logs"
)

var logBuffer = logs.NewRing(2000)

type runtimeController struct {
	mu     sync.Mutex
	app    *app
	ctx    context.Context
	cancel context.CancelFunc
	since  time.Time
}

func newRuntimeController() *runtimeController {
	return &runtimeController{}
}

func (r *runtimeController) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.startLocked(ctx)
}

func (r *runtimeController) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closeLocked(ctx)
}

func (r *runtimeController) Status(ctx context.Context) (admin.Status, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()

	status := admin.Status{
		Running:   r.app != nil,
		Platforms: make(map[string]string),
		Process:   r.processStatusLocked(),
		Config:    config.Snapshot(true),
	}
	if r.app == nil {
		return status, nil
	}
	for _, client := range r.app.hub.All() {
		status.Platforms[string(client.Platform())] = string(client.Status())
	}
	return status, nil
}

func (r *runtimeController) Config(ctx context.Context) (map[string]any, error) {
	_ = ctx
	return config.Snapshot(true), nil
}

func (r *runtimeController) Logs(ctx context.Context, since uint64, limit int) ([]admin.LogEntry, error) {
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

func (r *runtimeController) UpdateConfig(ctx context.Context, patch map[string]any) error {
	return config.UpdateSettings(ctx, patch)
}

func (r *runtimeController) Reload(ctx context.Context) error {
	return config.Reload(ctx)
}

func (r *runtimeController) Restart(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.closeLocked(ctx); err != nil {
		return err
	}
	if err := config.Reload(ctx); err != nil {
		return err
	}
	return r.startLocked(context.Background())
}

func (r *runtimeController) startLocked(parent context.Context) error {
	if r.app != nil {
		return nil
	}
	app, err := newApp()
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(parent)
	if err := app.Start(runCtx); err != nil {
		cancel()
		return err
	}
	r.app = app
	r.ctx = runCtx
	r.cancel = cancel
	r.since = time.Now()
	log.Printf("程序已启动，平台数量: %d", len(app.hub.All()))
	return nil
}

func (r *runtimeController) closeLocked(ctx context.Context) error {
	if r.app == nil {
		return nil
	}
	if r.cancel != nil {
		r.cancel()
	}
	closeCtx := ctx
	if closeCtx == nil {
		var cancel context.CancelFunc
		closeCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	err := r.app.Close(closeCtx)
	r.app = nil
	r.ctx = nil
	r.cancel = nil
	r.since = time.Time{}
	if err != nil {
		return fmt.Errorf("关闭程序失败: %w", err)
	}
	return nil
}

func (r *runtimeController) processStatusLocked() admin.ProcessStatus {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	var uptime time.Duration
	if !r.since.IsZero() {
		uptime = time.Since(r.since)
	}
	cpu := cpuTime()

	return admin.ProcessStatus{
		Uptime:      formatDuration(uptime),
		MemoryAlloc: formatBytes(mem.Alloc),
		MemorySys:   formatBytes(mem.Sys),
		HeapObjects: mem.HeapObjects,
		Goroutines:  runtime.NumGoroutine(),
		CPUPercent:  formatCPUPercent(cpu, uptime),
		CPUTime:     formatDuration(cpu),
	}
}

func cpuTime() time.Duration {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0
	}
	user := time.Duration(usage.Utime.Sec)*time.Second + time.Duration(usage.Utime.Usec)*time.Microsecond
	system := time.Duration(usage.Stime.Sec)*time.Second + time.Duration(usage.Stime.Usec)*time.Microsecond
	return user + system
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f PB", value/unit)
}

func formatCPUPercent(cpu time.Duration, uptime time.Duration) string {
	if cpu <= 0 || uptime <= 0 || runtime.NumCPU() <= 0 {
		return "0.0%"
	}
	percent := cpu.Seconds() / uptime.Seconds() / float64(runtime.NumCPU()) * 100
	return fmt.Sprintf("%.1f%%", percent)
}

func formatDuration(duration time.Duration) string {
	if duration <= 0 {
		return "0s"
	}
	duration = duration.Round(time.Second)
	days := duration / (24 * time.Hour)
	duration -= days * 24 * time.Hour
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
