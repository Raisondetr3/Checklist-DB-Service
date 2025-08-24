package http

import (
	"github.com/Raisondetr3/checklist-db-service/internal/config"
	"github.com/Raisondetr3/checklist-db-service/internal/service"
	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	config  *config.Config
	service service.HealthService
}

func NewHTTPHandlers(cfg *config.Config, healthService service.HealthService) *HTTPHandlers {
	return &HTTPHandlers{
		config: cfg,
		service: healthService,
	}
}

func (h *HTTPHandlers) SetupRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.HandleHealthCheck).Methods("GET")
}
