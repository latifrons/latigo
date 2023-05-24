package cron

import (
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"strings"
	"time"
)

type CronJob struct {
	Name     string
	Interval time.Duration
	Function interface{}
	Params   []interface{}
}

type CronJobProvider interface {
	ProvideAllJobs() []CronJob
	ProvideEnabledJobs() map[string]bool
}

type CronService struct {
	CronJobProvider CronJobProvider
	Logger          *zap.SugaredLogger
	cr              *gocron.Scheduler
	jobs            []CronJob
	jobEnabled      map[string]bool
}

func (s *CronService) InitJobs() {
	if s.CronJobProvider == nil {
		s.jobs = []CronJob{}
		return
	}
	jobs := s.CronJobProvider.ProvideAllJobs()
	s.jobEnabled = s.CronJobProvider.ProvideEnabledJobs()

	for _, job := range jobs {
		if v, ok := s.jobEnabled[strings.ToLower(job.Name)]; ok && v {
			s.jobs = append(s.jobs, job)
		} else {
			s.Logger.Infow("boot job disabled", "name", job.Name)
		}
	}
}

func (c *CronService) Start() {
	c.cr = gocron.NewScheduler(time.UTC)

	for _, job := range c.jobs {
		_, err := c.cr.Every(job.Interval).Do(job.Function, job.Params...)
		if err != nil {
			c.Logger.Fatalw("failed to start cron job", "name", job.Name, "error", err)
		}
	}
	c.cr.StartAsync()
}

func (c *CronService) Stop() {
	c.cr.Stop()
}

func (c *CronService) Name() string {
	return "CronService"
}
