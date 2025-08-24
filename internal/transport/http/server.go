package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Raisondetr3/checklist-db-service/internal/config"
	"github.com/Raisondetr3/checklist-db-service/internal/transport/http/middleware"
	"github.com/gorilla/mux"
)

type HTTPServer struct {
	server   *http.Server
	handlers *HTTPHandlers
	config   *config.Config
}

func NewHTTPServer(cfg *config.Config, handlers *HTTPHandlers) *HTTPServer {
	router := mux.NewRouter()

	router.Use(middleware.PanicRecoveryMiddleware)
	router.Use(middleware.LoggingMiddleware)

	handlers.SetupRoutes(router)

	return &HTTPServer{
		handlers: handlers,
		config:   cfg,
		server: &http.Server{
			Addr:         ":" + cfg.Server.HTTPPort,
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}
}

func (s *HTTPServer) StartServer() error {
	slog.Info("Starting HTTP server",
		slog.String("address", s.server.Addr),
	)

	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("HTTP server stopped")
			return nil
		}
		slog.Error("HTTP server error", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	slog.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}
