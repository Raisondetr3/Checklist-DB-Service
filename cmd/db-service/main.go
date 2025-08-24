package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/config"
	"github.com/Raisondetr3/checklist-db-service/internal/repository"
	"github.com/Raisondetr3/checklist-db-service/internal/service"
	grpcTransport "github.com/Raisondetr3/checklist-db-service/internal/transport/grpc"
	httpTransport "github.com/Raisondetr3/checklist-db-service/internal/transport/http"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/joho/godotenv"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	loggerCfg := logger.Config{
		Level:    cfg.Logging.Level,
		FilePath: cfg.Logging.FilePath,
		FileName: cfg.Logging.FileName,
	}

	if err := logger.SetupLogger(loggerCfg, "db-service"); err != nil {
		panic("Failed to setup logger: " + err.Error())
	}

	logger.LogServiceStart("db-service", map[string]interface{}{
		"http_port":     cfg.Server.HTTPPort,
		"grpc_port":     cfg.Server.GRPCPort,
		"db_host":       cfg.Database.Host,
		"db_name":       cfg.Database.Name,
		"log_level":     cfg.Logging.Level,
		"redis_enabled": cfg.Redis.Enabled,
	})

	defer logger.LogServiceStop("db-service", "shutdown")

	dbPool, err := initDatabaseWithRetry(cfg, 10, 5*time.Second)
	if err != nil {
		slog.Error("Failed to initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	healthRepo := repository.NewHealthRepository(dbPool)
	taskRepo := repository.NewTaskRepository(dbPool)

	healthService := service.NewHealthService(healthRepo)
	taskService := service.NewTaskService(taskRepo)

	handlers := httpTransport.NewHTTPHandlers(cfg, healthService)
	httpServer := httpTransport.NewHTTPServer(cfg, handlers)
	grpcServer := grpcTransport.NewGRPCServer(cfg, taskService)

	var wg sync.WaitGroup

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting HTTP server", slog.String("port", cfg.Server.HTTPPort))

		if err := httpServer.StartServer(); err != nil {
			slog.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting gRPC server", slog.String("port", cfg.Server.GRPCPort))

		if err := grpcServer.StartServer(); err != nil {
			slog.Error("gRPC server error", slog.String("error", err.Error()))
		}
	}()

	<-quit
	slog.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Stopping HTTP server...")
	if err := httpServer.Stop(ctx); err != nil {
		slog.Error("Error stopping HTTP server", slog.String("error", err.Error()))
	}

	slog.Info("Stopping gRPC server...")
	if err := grpcServer.Stop(ctx); err != nil {
		slog.Error("Error stopping gRPC server", slog.String("error", err.Error()))
	}

	slog.Info("Waiting for servers to stop...")
	wg.Wait()

	slog.Info("All servers stopped successfully")
}

func initDatabaseWithRetry(cfg *config.Config, maxRetries int, delay time.Duration) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	for i := 0; i < maxRetries; i++ {
		slog.Info("Attempting to connect to database",
			slog.Int("attempt", i+1),
			slog.Int("max_attempts", maxRetries))

		pool, err = initDatabase(cfg)
		if err == nil {
			slog.Info("Successfully connected to database")
			return pool, nil
		}

		slog.Warn("Database connection failed, retrying...",
			slog.String("error", err.Error()),
			slog.Duration("retry_in", delay))

		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}

	return nil, err
}

func initDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	logger.LogDatabaseConnection(ctx, cfg.Database.DSN(), "connect", nil)

	return pool, nil
}
