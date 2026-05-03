package api

import (
	"context"
	"fmt"
	"time"

	"github.com/nhirsama/onePushBot/internal/domain"
)

type noopScheduler struct{}

func (noopScheduler) AddIntervalJob(string, time.Duration, func(context.Context) error) error {
	return nil
}

func (noopScheduler) AddCronJob(string, string, func(context.Context) error) error {
	return nil
}

func (noopScheduler) AddDailyJob(string, int, int, func(context.Context) error) error {
	return nil
}

func (noopScheduler) AddWeeklyJob(string, time.Weekday, int, int, func(context.Context) error) error {
	return nil
}

func (noopScheduler) AddMonthlyJob(string, int, int, int, func(context.Context) error) error {
	return nil
}

func (noopScheduler) RemoveJob(string) error {
	return nil
}

func (noopScheduler) Start() error {
	return nil
}

func (noopScheduler) Shutdown(context.Context) error {
	return nil
}

func (noopScheduler) ListJobs() []domain.SchedulerJobInfo {
	return nil
}

func (noopScheduler) GetJobInfo(name string) (domain.SchedulerJobInfo, error) {
	return domain.SchedulerJobInfo{}, fmt.Errorf("job %s not found", name)
}
