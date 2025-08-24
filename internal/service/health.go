package service

import (
	"context"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/repository"
	"github.com/Raisondetr3/checklist-db-service/pkg/dto"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
)

type HealthService interface {
	Health(ctx context.Context) (*dto.HealthStatus, error)
}

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

type healthService struct {
	healthRepo repository.HealthRepository
}

func NewHealthService(healthRepo repository.HealthRepository) HealthService {
	return &healthService{
		healthRepo: healthRepo,
	}
}

func (s *healthService) Health(ctx context.Context) (*dto.HealthStatus, error) {
	start := time.Now()
	operation := "Health"

	err := s.healthRepo.HealthCheck(ctx)
	duration := time.Since(start)

	if err != nil {
		status := StatusUnhealthy
		if repository.IsConnectionError(err) {
			status = StatusUnhealthy
		}

		logger.LogError(ctx, err, operation)

		logger.LogTaskOperation(ctx, operation, "system", duration, err)

		return &dto.HealthStatus{
			Status:    status,
			Timestamp: time.Now(),
			Duration:  duration,
		}, nil
	}

	logger.LogTaskOperation(ctx, operation, "system", duration, nil)

	return &dto.HealthStatus{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}
