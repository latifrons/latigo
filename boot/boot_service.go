package boot

import (
	"github.com/rs/zerolog/log"
	"strings"
)

type BootJobProvider interface {
	ProvideAllJobs() []BootJob
	ProvideDisabledJobs() map[string]bool
}

type BootJob struct {
	Name          string
	Function      func() error
	FaultTolerant bool
}

type BootService struct {
	BootJobProvider BootJobProvider
	jobs            []BootJob
	jobDisabled     map[string]bool
}

func (s *BootService) InitJobs() {
	if s.BootJobProvider == nil {
		s.jobs = []BootJob{}
		return
	}

	jobs := s.BootJobProvider.ProvideAllJobs()
	s.jobDisabled = s.BootJobProvider.ProvideDisabledJobs()

	for _, job := range jobs {
		if v, ok := s.jobDisabled[strings.ToLower(job.Name)]; v && ok {
			log.Info().Str("name", job.Name).Msg("boot job disabled")
		} else {
			log.Info().Str("name", job.Name).Msg("boot job enabled")

			s.jobs = append(s.jobs, job)
		}
	}
}

func (s *BootService) Boot() {
	var err error
	for _, job := range s.jobs {
		log.Info().Str("name", job.Name).Msg("executing boot job")
		err = job.Function()
		if err != nil {
			if job.FaultTolerant {
				log.Error().Err(err).Str("name", job.Name).Msg("failed to execute boot job")
			} else {
				log.Fatal().Err(err).Str("name", job.Name).Msg("failed to execute boot job")
			}
		} else {
			log.Info().Str("name", job.Name).Msg("boot job executed")
		}
	}
	return
}
