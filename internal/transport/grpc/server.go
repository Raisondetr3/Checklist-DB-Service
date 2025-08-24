package grpc

import (
	"context"
	"log/slog"
	"net"

	"github.com/Raisondetr3/checklist-db-service/internal/config"
	"github.com/Raisondetr3/checklist-db-service/internal/service"
	"github.com/Raisondetr3/checklist-db-service/internal/transport/grpc/middleware"
	pb "github.com/Raisondetr3/checklist-db-service/pkg/pb"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	pb.UnimplementedTaskServiceServer
	taskService service.TaskService
	server      *grpc.Server
	config      *config.Config
}

func NewGRPCServer(cfg *config.Config, taskService service.TaskService) *GRPCServer {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			middleware.ChainUnaryInterceptors(
				middleware.PanicRecoveryUnaryInterceptor,
				middleware.RequestIDUnaryInterceptor,
				middleware.LoggingUnaryInterceptor,
			),
		),
	)

	grpcServer := &GRPCServer{
		taskService: taskService,
		server:      server,
		config:      cfg,
	}

	pb.RegisterTaskServiceServer(server, grpcServer)

	return grpcServer
}

func (s *GRPCServer) StartServer() error {
	address := ":" + s.config.Server.GRPCPort

	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Failed to listen on gRPC port",
			slog.String("address", address),
			slog.String("error", err.Error()))
		return err
	}

	slog.Info("gRPC server starting", slog.String("address", address))

	if err := s.server.Serve(listener); err != nil {
		slog.Error("gRPC server error", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	slog.Info("Stopping gRPC server")

	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		slog.Warn("gRPC server shutdown timeout, forcing stop")
		s.server.Stop()
		return ctx.Err()
	}
}