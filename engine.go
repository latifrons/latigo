package latigo

import (
	"github.com/latifrons/latigo/boot"
	"github.com/latifrons/latigo/cron"
	"github.com/latifrons/latigo/program"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type BasicEngine struct {
	Name              string
	EnvPrefix         string
	DumpConfigOnStart bool
	LogLevel          string
	PostBootLatency   time.Duration

	bootService      *boot.BootService
	cronService      *cron.CronService
	componentService *program.ComponentService
	postBootService  *boot.BootService
	injector         boot.Injector
}

func (b *BasicEngine) SetupBootJob(bootJobProvider boot.BootJobProvider) {
	b.bootService = &boot.BootService{
		BootJobProvider: bootJobProvider,
	}
}
func (b *BasicEngine) SetupCronJob(cronJobProvider cron.CronJobProvider) {
	b.cronService = &cron.CronService{
		CronJobProvider: cronJobProvider,
	}
}

func (b *BasicEngine) SetupComponentProvider(componentProvider program.ComponentProvider) {
	b.componentService = &program.ComponentService{
		ComponentProvider: componentProvider,
	}
}

func (b *BasicEngine) SetupPostBootJob(bootJobProvider boot.BootJobProvider) {
	b.postBootService = &boot.BootService{
		BootJobProvider: bootJobProvider,
	}
}

func (b *BasicEngine) SetupInjector(injector boot.Injector) {
	b.injector = injector
}

func (b *BasicEngine) setup() {
	if b.bootService != nil {
		b.bootService.InitJobs()
	}

	if b.componentService == nil {
		b.componentService = &program.ComponentService{}
	}
	b.componentService.InitComponents()

	if b.cronService != nil {
		b.cronService.InitJobs()

		b.componentService.AddComponent(b.cronService)
	}
	if b.postBootService != nil {
		b.postBootService.InitJobs()
	}
}

func (b *BasicEngine) Start() {
	log.Info().Str("name", b.Name).Msg("Starting basic server")
	b.setup()

	if b.bootService != nil {
		b.bootService.Boot()
	}
	if b.componentService != nil {
		b.componentService.Start()
	}
	log.Info().Dur("sleep", b.PostBootLatency).Msg("wait to start post boot jobs")
	time.Sleep(b.PostBootLatency)
	if b.postBootService != nil {
		b.postBootService.Boot()
	}

	// prevent sudden stop. Do your clean up here
	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		sig := <-gracefulStop
		log.Info().Str("name", b.Name).Str("sig", sig.String()).Msg("caught sig")
		log.Info().Str("name", b.Name).Msg("Exiting... Please do no kill me")
		b.componentService.Stop()
		os.Exit(0)
	}()
}

func NewDefaultEngine() BasicEngine {
	return BasicEngine{
		Name:              "LatiEngine",
		EnvPrefix:         "INJ",
		DumpConfigOnStart: true,
		LogLevel:          "INFO",
	}
}
