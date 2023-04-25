package latigo

import (
	"github.com/latifrons/latigo/boot"
	"github.com/latifrons/latigo/cron"
	"github.com/latifrons/latigo/program"
	"github.com/latifrons/latigo/rpcserver"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type BasicRpcServer struct {
	Name              string
	FolderConfig      program.FolderConfig
	EnvPrefix         string
	DumpConfigOnStart bool
	LogLevel          string

	bootService      *boot.BootService
	cronService      *cron.CronService
	componentService *program.ComponentService
	server           *rpcserver.RpcServer
	injector         boot.Injector
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
func (b *BasicRpcServer) SetupServer(provider rpcserver.RouterProvider, port string, debugFlags rpcserver.DebugFlags) {
	b.server = &rpcserver.RpcServer{
		RouterProvider: provider,
		Port:           port,
		DebugFlags:     debugFlags,
	}
}

func (b *BasicRpcServer) SetupInjector(injector boot.Injector) {
	b.injector = injector
}

func (b *BasicRpcServer) setup() {
	// init logger first.
	b.FolderConfig = program.EnsureFolders(b.FolderConfig)

	program.ReadNormalConfig(b.FolderConfig.Config)
	program.ReadPrivate(b.FolderConfig.Private)
	program.ReadEnvConfig(b.EnvPrefix)
	program.DumpConfig()
	program.SetupLogger(b.LogLevel)

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
	logrus.Info(b.Name + " Starting")
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
		logrus.Infof("caught sig: %+v", sig)
		logrus.Info("Exiting... Please do no kill me")
		b.componentService.Stop()
		os.Exit(0)
	}()
}

func NewDefaultBasicRpcServer() BasicRpcServer {
	return BasicRpcServer{
		Name: "LatiServer",
		FolderConfig: program.FolderConfig{
			Root:    "data",
			Log:     "",
			Data:    "",
			Config:  "",
			Private: "",
		},
		EnvPrefix:         "INJ",
		DumpConfigOnStart: true,
		LogLevel:          "INFO",
	}
}
