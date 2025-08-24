package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.written += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.New().String()

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		slog.Info("HTTP Request started",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		logger.LogHTTPRequest(
			r.Context(),
			r.Method,
			r.URL.Path,
			r.UserAgent(),
			requestID,
			duration,
			wrapped.statusCode,
		)
	})
}

func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestIDFromContext(r.Context())
				
				slog.Error("Panic recovered in HTTP handler",
					slog.String("request_id", requestID),
					slog.Any("panic", err),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}