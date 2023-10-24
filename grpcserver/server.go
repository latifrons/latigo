package grpcserver

import (
	"fmt"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"runtime/debug"
)

type ServiceProvider interface {
	ProvideAllServices() []GrpcService
}

type GrpcService struct {
	Desc *grpc.ServiceDesc
	SS   any
}

type DebugFlags struct {
	GRpcDebug   bool
	RequestLog  bool
	ResponseLog bool
}

type GrpcServer struct {
	ServiceProvider ServiceProvider
	Port            string
	DebugFlags      DebugFlags
	server          *grpc.Server
	logger          zerolog.Logger
}

func (srv *GrpcServer) Start() {
	//srv.logger = zerolog.New(os.Stdout)
	//output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "01-02 15:04:05.000"}
	//srv.logger = zerolog.New(output).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	srv.logger = log.Logger

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			debug.PrintStack()
			log.Error().Interface("panic", p).Msg("panic")
			return fmt.Errorf("%s", p)
		})),
	}
	streamInterceptos := []grpc.StreamServerInterceptor{}

	if srv.DebugFlags.GRpcDebug {
		events := []logging.LoggableEvent{}
		if srv.DebugFlags.RequestLog {
			events = append(events, logging.PayloadReceived)
			events = append(events, logging.StartCall)
		} else {
			events = append(events, logging.StartCall)
		}
		if srv.DebugFlags.ResponseLog {
			events = append(events, logging.PayloadSent)
			events = append(events, logging.FinishCall)
		} else {
			events = append(events, logging.FinishCall)
		}

		opts := []logging.Option{
			logging.WithLogOnEvents(events...),
			//logging.WithDisableLoggingFields("method_type", "protocol"),
			// Add any other option (check functions starting with logging.With).
		}
		unaryInterceptors = append(unaryInterceptors, logging.UnaryServerInterceptor(InterceptorLogger(srv.logger), opts...))
		streamInterceptos = append(streamInterceptos, logging.StreamServerInterceptor(InterceptorLogger(srv.logger), opts...))
	}

	srv.server = grpc.NewServer(
		//grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		//
		//	//grpc_validator.UnaryServerInterceptor(),
		//	//grpc_zap.UnaryServerInterceptor(zapLogger),
		//	//grpc_prometheus.UnaryServerInterceptor,
		//	//grpc_opentracing.UnaryServerInterceptor(),
		//	//grpc_ctxtags.UnaryServerInterceptor(),
		//	//grpc_auth.UnaryServerInterceptor(authFunc),
		//	//grpc_ratelimit.UnaryServerInterceptor(rateLimiter),
		//	//grpc_zap.PayloadUnaryServerInterceptor(zapLogger, func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
		//	//	return strings.HasPrefix(fullMethodName, "/grpc.health.v1.Health/")
		//	//}),
		//)),
		grpc.ChainUnaryInterceptor(
			unaryInterceptors...,
		// Add any other interceptor you want.
		),
		grpc.ChainStreamInterceptor(
			streamInterceptos...,
		),
	)
	for _, service := range srv.ServiceProvider.ProvideAllServices() {
		srv.server.RegisterService(service.Desc, service.SS)
	}
	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+srv.Port)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to listen")
	}
	log.Info().Str("port", srv.Port).Msg("listening gRPC on " + srv.Port)
	go func() {
		if err := srv.server.Serve(lis); err != nil {
			log.Fatal().Stack().Err(err).Msg("failed to serve")
		}
	}()
}

func (srv *GrpcServer) Stop() {
	srv.server.Stop()
}

func (srv *GrpcServer) Name() string {
	return fmt.Sprintf("grpcServer at port %s", srv.Port)

}
