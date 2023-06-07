package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
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
	ProvideDisabledJobs() map[string]bool
}

type CronService struct {
	CronJobProvider CronJobProvider
	cr              *gocron.Scheduler
	jobs            []CronJob
	jobDisabled     map[string]bool
}

func (s *CronService) InitJobs() {
	if s.CronJobProvider == nil {
		s.jobs = []CronJob{}
		return
	}
	jobs := s.CronJobProvider.ProvideAllJobs()
	s.jobDisabled = s.CronJobProvider.ProvideDisabledJobs()

	for _, job := range jobs {
		if v, ok := s.jobDisabled[strings.ToLower(job.Name)]; ok && v {
			s.jobs = append(s.jobs, job)
		} else {
			log.Info().Str("name", job.Name).Msg("boot job disabled")
		}
	}
}

func (c *CronService) Start() {
	c.cr = gocron.NewScheduler(time.UTC)

	for _, job := range c.jobs {
		_, err := c.cr.Every(job.Interval).Do(job.Function, job.Params...)
		if err != nil {
			log.Fatal().Err(err).Str("name", job.Name).Msg("failed to start cron job")
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
