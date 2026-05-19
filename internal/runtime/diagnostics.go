package runtime

import (
	"fmt"
	goruntime "runtime"
	"syscall"
	"time"

	"github.com/nhirsama/onePushBot/internal/admin"
)

func processStatus(since time.Time) admin.ProcessStatus {
	var mem goruntime.MemStats
	goruntime.ReadMemStats(&mem)
	var uptime time.Duration
	if !since.IsZero() {
		uptime = time.Since(since)
	}
	cpu := cpuTime()

	return admin.ProcessStatus{
		Uptime:      formatDuration(uptime),
		MemoryAlloc: formatBytes(mem.Alloc),
		MemorySys:   formatBytes(mem.Sys),
		HeapObjects: mem.HeapObjects,
		Goroutines:  goruntime.NumGoroutine(),
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
	if cpu <= 0 || uptime <= 0 || goruntime.NumCPU() <= 0 {
		return "0.0%"
	}
	percent := cpu.Seconds() / uptime.Seconds() / float64(goruntime.NumCPU()) * 100
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
