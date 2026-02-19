package grpc

import (
	"log/slog"

	"google.golang.org/grpc"
)

// NewServer creates a gRPC server with tenant and auth interceptors.
func NewServer() *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			TenantUnaryInterceptor(),
		),
	)
	slog.Info("gRPC server created")
	return srv
}
