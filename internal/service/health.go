package service

import (
	"context"

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
	err := s.healthRepo.HealthCheck(ctx)
	if err != nil {
		logger.LogError(ctx, err, "database_health_check")
		return &dto.HealthStatus{
			Status: StatusUnhealthy,
		}, nil
	}

	return &dto.HealthStatus{
		Status: StatusHealthy,
	}, nil
}
