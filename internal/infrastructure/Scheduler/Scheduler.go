package Scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nhirsama/onePushBot/internal/domain"
)

// jobMetadata 任务元数据
type jobMetadata struct {
	name     string
	schedule string
	jobFunc  func(ctx context.Context) error
}

type Scheduler struct {
	scheduler gocron.Scheduler
	jobs      map[string]gocron.Job // 任务名称到Job的映射
	jobMeta   map[string]*jobMetadata
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	started   bool
}

func NewScheduler() (domain.Scheduler, error) {
	// 创建gocron调度器
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("创建调度器失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		scheduler: s,
		jobs:      make(map[string]gocron.Job),
		jobMeta:   make(map[string]*jobMetadata),
		ctx:       ctx,
		cancel:    cancel,
		started:   false,
	}, nil
}

// AddIntervalJob 添加间隔任务
func (s *Scheduler) AddIntervalJob(name string, interval time.Duration, job func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	// 包装任务函数,传入context
	wrappedJob := s.wrapJob(name, job)

	// 创建间隔任务
	j, err := s.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(wrappedJob),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to create interval job: %w", err)
	}

	s.jobs[name] = j
	s.jobMeta[name] = &jobMetadata{
		name:     name,
		schedule: fmt.Sprintf("Every %v", interval),
		jobFunc:  job,
	}

	return nil
}

// AddCronJob 添加Cron任务
// cronExpr格式: "0 30 * * * *" (秒 分 时 日 月 周)
func (s *Scheduler) AddCronJob(name string, cronExpr string, job func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	wrappedJob := s.wrapJob(name, job)

	// 创建Cron任务
	j, err := s.scheduler.NewJob(
		gocron.CronJob(cronExpr, true), // true表示使用秒级精度
		gocron.NewTask(wrappedJob),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to create cron job: %w", err)
	}

	s.jobs[name] = j
	s.jobMeta[name] = &jobMetadata{
		name:     name,
		schedule: fmt.Sprintf("Cron: %s", cronExpr),
		jobFunc:  job,
	}

	return nil
}

// AddDailyJob 添加每日任务
func (s *Scheduler) AddDailyJob(name string, hour, minute int, job func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	wrappedJob := s.wrapJob(name, job)

	// 创建每日任务
	j, err := s.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(uint(hour), uint(minute), 0),
		)),
		gocron.NewTask(wrappedJob),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to create daily job: %w", err)
	}

	s.jobs[name] = j
	s.jobMeta[name] = &jobMetadata{
		name:     name,
		schedule: fmt.Sprintf("Daily at %02d:%02d", hour, minute),
		jobFunc:  job,
	}

	return nil
}

// AddWeeklyJob 添加每周任务
func (s *Scheduler) AddWeeklyJob(name string, weekday time.Weekday, hour, minute int, job func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	wrappedJob := s.wrapJob(name, job)

	// 创建每周任务
	j, err := s.scheduler.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(weekday), gocron.NewAtTimes(
			gocron.NewAtTime(uint(hour), uint(minute), 0),
		)),
		gocron.NewTask(wrappedJob),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to create weekly job: %w", err)
	}

	s.jobs[name] = j
	s.jobMeta[name] = &jobMetadata{
		name:     name,
		schedule: fmt.Sprintf("Weekly %s at %02d:%02d", weekday, hour, minute),
		jobFunc:  job,
	}

	return nil
}

// AddMonthlyJob 添加每月任务
func (s *Scheduler) AddMonthlyJob(name string, day, hour, minute int, job func(ctx context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	wrappedJob := s.wrapJob(name, job)

	// 创建每月任务
	j, err := s.scheduler.NewJob(
		gocron.MonthlyJob(1, gocron.NewDaysOfTheMonth(day), gocron.NewAtTimes(
			gocron.NewAtTime(uint(hour), uint(minute), 0),
		)),
		gocron.NewTask(wrappedJob),
		gocron.WithName(name),
	)
	if err != nil {
		return fmt.Errorf("failed to create monthly job: %w", err)
	}

	s.jobs[name] = j
	s.jobMeta[name] = &jobMetadata{
		name:     name,
		schedule: fmt.Sprintf("Monthly day %d at %02d:%02d", day, hour, minute),
		jobFunc:  job,
	}

	return nil
}

// RemoveJob 移除任务
func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	// 从gocron中移除任务
	if err := s.scheduler.RemoveJob(job.ID()); err != nil {
		return fmt.Errorf("failed to remove job: %w", err)
	}

	delete(s.jobs, name)
	delete(s.jobMeta, name)

	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("scheduler already started")
	}

	s.scheduler.Start()
	s.started = true

	return nil
}

// Shutdown 关闭调度器(使用context控制超时)
func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	s.mu.Unlock()

	// 取消所有任务的context
	s.cancel()

	// 等待调度器关闭
	done := make(chan error, 1)
	go func() {
		done <- s.scheduler.Shutdown()
	}()

	// 等待关闭完成或超时
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// ListJobs 列出所有任务
func (s *Scheduler) ListJobs() []domain.SchedulerJobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]domain.SchedulerJobInfo, 0, len(s.jobs))
	for name, job := range s.jobs {
		meta := s.jobMeta[name]

		nextRun, _ := job.NextRun()
		lastRun, _ := job.LastRun()

		infos = append(infos, domain.SchedulerJobInfo{
			Name:      name,
			Schedule:  meta.schedule,
			NextRun:   nextRun,
			LastRun:   lastRun,
			IsRunning: false, // gocron v2 不直接提供运行状态
		})
	}

	return infos
}

// GetJobInfo 获取任务信息
func (s *Scheduler) GetJobInfo(name string) (domain.SchedulerJobInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[name]
	if !exists {
		return domain.SchedulerJobInfo{}, fmt.Errorf("job %s not found", name)
	}

	meta := s.jobMeta[name]
	nextRun, _ := job.NextRun()
	lastRun, _ := job.LastRun()

	return domain.SchedulerJobInfo{
		Name:      name,
		Schedule:  meta.schedule,
		NextRun:   nextRun,
		LastRun:   lastRun,
		IsRunning: false,
	}, nil
}

// wrapJob 包装任务函数,使其能接收context并处理取消
func (s *Scheduler) wrapJob(name string, job func(ctx context.Context) error) func() {
	return func() {
		// 使用调度器的context
		if err := job(s.ctx); err != nil {
			fmt.Printf("Job %s execution error: %v\n", name, err)
		}
	}
}
