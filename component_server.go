package latigo

import (
	"github.com/latifrons/latigo/boot"
	"github.com/latifrons/latigo/cron"
	"github.com/latifrons/latigo/program"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type BasicServer struct {
	Name              string
	EnvPrefix         string
	DumpConfigOnStart bool
	Logger            *zap.SugaredLogger
	LogLevel          string

	bootService      *boot.BootService
	cronService      *cron.CronService
	componentService *program.ComponentService
	injector         boot.Injector
}

func (b *BasicServer) SetupBootJob(bootJobProvider boot.BootJobProvider) {
	b.bootService = &boot.BootService{
		BootJobProvider: bootJobProvider,
		Logger:          b.Logger,
	}
}
func (b *BasicServer) SetupCronJob(cronJobProvider cron.CronJobProvider) {
	b.cronService = &cron.CronService{
		CronJobProvider: cronJobProvider,
		Logger:          b.Logger,
	}
}

func (b *BasicServer) SetupComponentProvider(componentProvider program.ComponentProvider) {
	b.componentService = &program.ComponentService{
		ComponentProvider: componentProvider,
		Logger:            b.Logger,
	}
}
func (b *BasicServer) SetupInjector(injector boot.Injector) {
	b.injector = injector
}

func (b *BasicServer) setup() {
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
}

func (b *BasicServer) Start() {
	b.Logger.Infow("Starting basic server", "name", b.Name)
	b.setup()

	if b.bootService != nil {
		b.bootService.Boot()
	}
	if b.componentService != nil {
		b.componentService.Start()
	}

	// prevent sudden stop. Do your clean up here
	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		sig := <-gracefulStop
		b.Logger.Infow("caught sig", "sig", sig)
		b.Logger.Infow("Exiting... Please do no kill me")
		b.componentService.Stop()
		os.Exit(0)
	}()
}

func NewDefaultBasicServer() BasicServer {
	return BasicServer{
		Name:              "LatiServer",
		EnvPrefix:         "INJ",
		DumpConfigOnStart: true,
		LogLevel:          "INFO",
		Logger:            zap.S(),
	}
}
