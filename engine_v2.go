package latigo

import (
	"github.com/go-co-op/gocron"
	"github.com/latifrons/latigo/boot"
	"github.com/latifrons/latigo/cron"
	"github.com/latifrons/latigo/program"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type BootType string

const BootTypeOnce BootType = "once"
const BootTypeCron BootType = "cron"
const BootTypeComponent BootType = "component"

type BootSequence struct {
	Type BootType
	Job  interface{}
}

type EngineV2 struct {
	Name                 string
	EnvPrefix            string
	DumpConfigOnStart    bool
	LogLevel             string
	Jobs                 []BootSequence
	registeredCrons      []cron.CronJob
	registeredComponents []program.Component
	cr                   *gocron.Scheduler
}

func (b *EngineV2) setup() {
	b.registeredCrons = []cron.CronJob{}
	b.registeredComponents = []program.Component{}
}

func (b *EngineV2) Start() {
	log.Info().Str("name", b.Name).Msg("Starting basic server")
	b.setup()

	var err error

	for _, job := range b.Jobs {

		switch job.Type {
		case BootTypeOnce:
			bootJob := job.Job.(boot.BootJob)
			log.Info().Str("name", bootJob.Name).Str("type", string(job.Type)).Msg("executing job")

			err = bootJob.Function()
			if err != nil {
				if bootJob.FaultTolerant {
					log.Error().Err(err).Str("name", bootJob.Name).Msg("failed to execute boot job")
				} else {
					log.Fatal().Err(err).Str("name", bootJob.Name).Msg("failed to execute boot job")
				}
			}
		case BootTypeComponent:
			component := job.Job.(program.Component)
			b.registeredComponents = append(b.registeredComponents, component)
			log.Info().Str("name", component.Name()).Msg("starting component")
			component.Start()
			log.Info().Str("name", component.Name()).Msg("started component")
		case BootTypeCron:
			cronJob := job.Job.(cron.CronJob)
			b.registeredCrons = append(b.registeredCrons, cronJob)
			log.Info().Str("name", cronJob.Name).Msg("registered cron job")
			// run later
		}
	}

	b.cr = gocron.NewScheduler(time.UTC)
	for _, job := range b.registeredCrons {
		if job.Type == cron.CronJobTypeCron {
			scheduler := b.cr.CronWithSeconds(job.Cron)
			_, err := scheduler.Do(job.Function, job.Params...)
			if err != nil {
				log.Fatal().Err(err).Str("name", job.Name).Msg("failed to start cron job")
			} else {
				log.Info().Str("name", job.Name).Msg("cron job started")
			}
			continue
		} else {
			scheduler := b.cr.Every(job.Interval)
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
	b.cr.StartAsync()

	// prevent sudden stop. Do your clean up here
	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		sig := <-gracefulStop
		log.Info().Str("name", b.Name).Str("sig", sig.String()).Msg("caught sig")
		log.Info().Str("name", b.Name).Msg("Exiting... Please do no kill me")
		// stop crons
		log.Info().Msg("stopping cron jobs")
		b.cr.Stop()
		log.Info().Msg("stopped cron jobs")
		// stop components
		for _, component := range b.registeredComponents {
			log.Info().Str("name", component.Name()).Msg("stopping component")
			component.Stop()
			log.Info().Str("name", component.Name()).Msg("stopped component")
		}
		os.Exit(0)
	}()
}

func NewDefaultEngineV2() EngineV2 {
	return EngineV2{
		Name:              "LatiEngineV2",
		EnvPrefix:         "INJ",
		DumpConfigOnStart: true,
		LogLevel:          "INFO",
	}
}
