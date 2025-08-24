package middleware

import (
	"context"
	"time"

	"log/slog"

	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LoggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	logger.LogGRPCRequest(ctx, info.FullMethod, duration, err)
	return resp, err
}

func RequestIDUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, "request_id", requestID)

	header := metadata.New(map[string]string{"x-request-id": requestID})
	grpc.SendHeader(ctx, header)

	return handler(ctx, req)
}

func PanicRecoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered in gRPC handler",
				slog.String("method", info.FullMethod),
				slog.Any("panic", r))
			err = status.Error(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}

func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			currentHandler := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, currentHandler)
			}
		}
		return chain(ctx, req)
	}
}
