package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	HealthCheck(ctx context.Context) error
}

type healthRepository struct {
	db *pgxpool.Pool
}

func NewHealthRepository(db *pgxpool.Pool) HealthRepository {
	return &healthRepository{
		db: db,
	}
}

func (r *healthRepository) HealthCheck(ctx context.Context) error {
	start := time.Now()

	var result int
	err := r.db.QueryRow(ctx, "SELECT 1").Scan(&result)

	duration := time.Since(start)

	if err != nil {
		r.logHealthCheckError(ctx, "basic_health_check", duration, err)
		return HandlePgxError("health_check", err)
	}

	if duration > 100*time.Millisecond {
		logger.LogSlowOperation(ctx, "health_check", duration, 100*time.Millisecond)
	}

	slog.DebugContext(ctx, "Health check successful",
		slog.Duration("duration", duration),
		slog.Int("result", result),
	)

	return nil
}

func (r *healthRepository) logHealthCheckError(ctx context.Context, operation string, duration time.Duration, err error) {
	logger.LogDatabaseQuery(ctx, "SELECT 1", []interface{}{}, duration, err)

	slog.ErrorContext(ctx, "Health check failed",
		slog.String("operation", operation),
		slog.String("error", err.Error()),
		slog.Duration("duration", duration),
		slog.String("type", "health_check_failure"),
	)
}
