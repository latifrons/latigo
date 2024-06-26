package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type DistributedTask struct {
	CheckInterval  time.Duration
	ActionInterval time.Duration
	LockKey        string
}

const CronJobTypeCron = "cron"
const CronJobTypeInterval = "interval"

type CronJob struct {
	Name            string
	Type            string
	Cron            string
	WaitForSchedule bool
	Singleton       bool
	Interval        time.Duration
	Function        interface{}
	Params          []interface{}
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
			log.Info().Str("name", job.Name).Msg("cron job disabled")
		} else {
			log.Info().Str("name", job.Name).Msg("cron job enabled")
			s.jobs = append(s.jobs, job)
		}
	}
}

func (c *CronService) Start() {
	c.cr = gocron.NewScheduler(time.UTC)

	for _, job := range c.jobs {
		if job.Type == CronJobTypeCron {
			scheduler := c.cr.CronWithSeconds(job.Cron).SingletonMode()
			_, err := scheduler.Do(job.Function, job.Params...)
			if err != nil {
				log.Fatal().Err(err).Str("name", job.Name).Msg("failed to start cron job")
			} else {
				log.Info().Str("name", job.Name).Msg("cron job started")
			}
			continue
		} else {
			scheduler := c.cr.Every(job.Interval).SingletonMode()

			if job.WaitForSchedule {
				scheduler = scheduler.WaitForSchedule()
			} else {
				scheduler = scheduler.StartImmediately()
			}

			_, err := scheduler.Do(job.Function, job.Params...)

			if err != nil {
				log.Fatal().Err(err).Str("name", job.Name).Msg("failed to start cron job")
			} else {
				log.Info().Str("name", job.Name).Msg("cron job started")
			}
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
