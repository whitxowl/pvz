package grpcsrv

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	grpcsrv "github.com/whitxowl/pvz.git/internal/api/grpc"
	"github.com/whitxowl/pvz.git/internal/service/pvz"
	"github.com/whitxowl/pvz.git/pkg/proto/pvz_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/whitxowl/pvz.git/internal/config"
)

type Server struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	cfg        config.GRPCServer
}

func New(log *slog.Logger, service *pvz.Service, cfg config.GRPCServer) *Server {
	grpcServer := grpc.NewServer()

	pvzHandler := grpcsrv.NewPVZHandler(service)
	pvz_v1.RegisterPVZServiceServer(grpcServer, pvzHandler)

	reflection.Register(grpcServer)

	return &Server{
		log:        log,
		grpcServer: grpcServer,
		cfg:        cfg,
	}
}

func (s *Server) MustRun(ctx context.Context) {
	if err := s.Run(ctx); err != nil {
		panic("failed to start grpc server: " + err.Error())
	}
}

func (s *Server) Run(ctx context.Context) error {
	const op = "grpcserver.Run"

	lis, err := net.Listen("tcp", s.cfg.Address())
	if err != nil {
		return fmt.Errorf("%s listen: %w", op, err)
	}

	s.log.InfoContext(ctx, "starting grpc server", slog.String("address", s.cfg.Address()))

	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop(ctx context.Context) {
	s.log.InfoContext(ctx, "stopping grpc server")
	s.grpcServer.GracefulStop()
}
