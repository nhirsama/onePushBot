package domain

import (
	"context"
	"time"
)

// Scheduler 调度器接口
type Scheduler interface {
	// AddIntervalJob 添加间隔任务
	AddIntervalJob(name string, interval time.Duration, job func(ctx context.Context) error) error

	// AddCronJob 添加Cron任务
	AddCronJob(name string, cronExpr string, job func(ctx context.Context) error) error

	// AddDailyJob 添加每日任务
	AddDailyJob(name string, hour, minute int, job func(ctx context.Context) error) error

	// AddWeeklyJob 添加每周任务
	AddWeeklyJob(name string, weekday time.Weekday, hour, minute int, job func(ctx context.Context) error) error

	// AddMonthlyJob 添加每月任务
	AddMonthlyJob(name string, day, hour, minute int, job func(ctx context.Context) error) error

	// RemoveJob 移除任务
	RemoveJob(name string) error

	// Start 启动调度器
	Start() error

	// Shutdown 关闭调度器(使用context控制)
	Shutdown(ctx context.Context) error

	// ListJobs 列出所有任务
	ListJobs() []SchedulerJobInfo

	// GetJobInfo 获取任务信息
	GetJobInfo(name string) (SchedulerJobInfo, error)
}

// SchedulerJobInfo JobInfo 任务信息
type SchedulerJobInfo struct {
	Name      string
	Schedule  string
	NextRun   time.Time
	LastRun   time.Time
	IsRunning bool
}
