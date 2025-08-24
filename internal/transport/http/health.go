package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Raisondetr3/checklist-db-service/pkg/dto"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
)

func (h *HTTPHandlers) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	health, err := h.service.Health(ctx)

	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusInternalServerError
		duration := time.Since(start)
		logger.LogHTTPRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), getRequestID(ctx), duration, statusCode)

		errDTO := dto.NewErr(err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(errDTO.ToString()))

		return
	}

	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		logger.LogError(ctx, err, "encode_health_response")
		
		return
	}

	duration := time.Since(start)
	logger.LogHTTPRequest(ctx, r.Method, r.URL.Path, r.UserAgent(), getRequestID(ctx), duration, statusCode)
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}