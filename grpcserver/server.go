package grpcserver

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
)

type GrpcService struct {
	Desc *grpc.ServiceDesc
	SS   any
}

type ServiceProvider interface {
	ProvideAllServices() []GrpcService
}

type GrpcServer struct {
	ServiceProvider ServiceProvider
	Port            string
	server          *grpc.Server
}

func (srv *GrpcServer) Start() {
	srv.server = grpc.NewServer()
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
