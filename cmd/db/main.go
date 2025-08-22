package main

import (
	"checklist-db-service/internal/config"
	"checklist-db-service/pkg/logger"
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	} else {
		slog.Info("Loaded configuration from .env file")
	}

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
		"grpc_port":     cfg.Server.GRPCPort,
		"db_host":       cfg.Database.Host,
		"db_name":       cfg.Database.Name,
		"redis_enabled": cfg.Redis.Enabled,
	})

	db, err := initDatabase(cfg)
	if err != nil {
		logger.LogError(context.Background(), err, "database_initialization")
		os.Exit(1)
	}
	defer closeDatabase(db)
}
