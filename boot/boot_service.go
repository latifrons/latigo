package boot

import (
	"go.uber.org/zap"
	"strings"
)

type BootJobProvider interface {
	ProvideAllJobs() []BootJob
	ProvideEnabledJobs() map[string]bool
}

type BootJob struct {
	Name          string
	Function      func() error
	FaultTolerant bool
}

type BootService struct {
	BootJobProvider BootJobProvider
	Logger          *zap.SugaredLogger
	jobs            []BootJob
	jobEnabled      map[string]bool
}

func (s *BootService) InitJobs() {
	if s.BootJobProvider == nil {
		s.jobs = []BootJob{}
		return
	}

	jobs := s.BootJobProvider.ProvideAllJobs()
	s.jobEnabled = s.BootJobProvider.ProvideEnabledJobs()

	for _, job := range jobs {
		if v, ok := s.jobEnabled[strings.ToLower(job.Name)]; ok && v {
			s.jobs = append(s.jobs, job)
		} else {
			s.Logger.Infow("boot job disabled", "name", job.Name)
		}
	}
}

func (s *BootService) Boot() {
	var err error
	for _, job := range s.jobs {
		s.Logger.Infow("starting boot job", "name", job.Name)
		err = job.Function()
		if err != nil {
			if job.FaultTolerant {
				s.Logger.Errorw("failed to execute boot job", "name", job.Name, "error", err)
			} else {
				s.Logger.Fatalw("failed to execute boot job", "name", job.Name, "error", err)
			}
		}
	}
	return
}
