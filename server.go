package latigo

import (
	"github.com/latifrons/latigo/boot"
	"github.com/latifrons/latigo/cron"
	"github.com/latifrons/latigo/program"
	"github.com/latifrons/latigo/rpcserver"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type BasicRpcServer struct {
	Name              string
	EnvPrefix         string
	DumpConfigOnStart bool
	bootService       *boot.BootService
	cronService       *cron.CronService
	componentService  *program.ComponentService
	server            *rpcserver.RpcServer
	injector          boot.Injector
}

func (b *BasicRpcServer) SetupBootJob(bootJobProvider boot.BootJobProvider) {
	b.bootService = &boot.BootService{
		BootJobProvider: bootJobProvider,
	}
}
func (b *BasicRpcServer) SetupCronJob(cronJobProvider cron.CronJobProvider) {
	b.cronService = &cron.CronService{
		CronJobProvider: cronJobProvider,
	}
}

func (b *BasicRpcServer) SetupComponentProvider(componentProvider program.ComponentProvider) {
	b.componentService = &program.ComponentService{
		ComponentProvider: componentProvider,
	}
}

func (b *BasicRpcServer) SetupInjector(injector boot.Injector) {
	b.injector = injector
}

func (b *BasicRpcServer) setup() {
	if b.bootService != nil {
		b.bootService.InitJobs()
	}

	if b.componentService == nil {
		b.componentService = &program.ComponentService{}
	}
	b.componentService.InitComponents()

	if b.server != nil {
		b.server.InitDefault()
		b.componentService.AddComponent(b.server)
	}
	if b.cronService != nil {
		b.cronService.InitJobs()

		b.componentService.AddComponent(b.cronService)
	}
}

func (b *BasicRpcServer) Start() {
	log.Info().Str("name", b.Name).Msg("Starting basic rpc server")
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
		log.Info().Str("name", b.Name).Str("signal", sig.String()).Msg("caught sig")
		log.Info().Str("name", b.Name).Msg("Exiting... Please do no kill me")
		b.componentService.Stop()
		os.Exit(0)
	}()
}

func NewDefaultBasicRpcServer() BasicRpcServer {
	return BasicRpcServer{
		Name:              "LatiServer",
		EnvPrefix:         "INJ",
		DumpConfigOnStart: true,
	}
}
