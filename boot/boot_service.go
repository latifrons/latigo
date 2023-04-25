package boot

import (
	"github.com/sirupsen/logrus"
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
			logrus.WithField("name", job.Name).Info("boot job disabled")
		}
	}
}

func (s *BootService) Boot() {
	var err error
	for _, job := range s.jobs {
		logrus.WithField("name", job.Name).Info("starting job")
		err = job.Function()
		if err != nil {
			if job.FaultTolerant {
				logrus.WithError(err).WithField("name", job.Name).Error("failed to execute boot job")
			} else {
				logrus.WithError(err).WithField("name", job.Name).Fatal("failed to execute boot job")
			}
		}
	}
	return
}
