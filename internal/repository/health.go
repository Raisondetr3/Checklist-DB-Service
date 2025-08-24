package repository

import (
	"context"
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
	var result int
	err := r.db.QueryRow(ctx, "SELECT 1").Scan(&result)
	return err
}