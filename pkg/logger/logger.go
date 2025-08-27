package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Level    string `env:"LOG_LEVEL" envDefault:"info"`
	FilePath string `env:"LOG_FILE_PATH" envDefault:"logs"`
	FileName string `env:"LOG_FILE_NAME"`
}

func SetupLogger(cfg Config, serviceName string) error {
	if err := os.MkdirAll(cfg.FilePath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	if cfg.FileName == "" {
		cfg.FileName = fmt.Sprintf("%s.log", serviceName)
	}

	fullPath := filepath.Join(cfg.FilePath, cfg.FileName)

	logFile, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(logFile, opts)

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	slog.SetDefault(logger)

	return nil
}

func LogHTTPRequest(ctx context.Context, method, path, userAgent, requestID string, duration time.Duration, statusCode int) {
	if ctx == nil {
		ctx = context.Background()
	}

	attrs := []slog.Attr{
		slog.String("type", "http_request"),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("user_agent", userAgent),
		slog.String("request_id", requestID),
		slog.Duration("duration", duration),
		slog.Int("status_code", statusCode),
	}

	if statusCode >= 500 {
		slog.LogAttrs(ctx, slog.LevelError, "HTTP Request", attrs...)
	} else if statusCode >= 400 {
		slog.LogAttrs(ctx, slog.LevelWarn, "HTTP Request", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "HTTP Request", attrs...)
	}
}

func LogGRPCRequest(ctx context.Context, method string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("type", "grpc_request"),
		slog.String("method", method),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "gRPC Request Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "gRPC Request", attrs...)
	}
}

func LogGRPCCall(ctx context.Context, service, method string, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	attrs := []slog.Attr{
		slog.String("type", "grpc_call"),
		slog.String("service", service),
		slog.String("method", method),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "gRPC Call Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, "gRPC Call", attrs...)
	}
}

func LogDatabaseQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("type", "database_query"),
		slog.String("query", query),
		slog.Any("args", args),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Database Query Failed", attrs...)
	} else {
		if duration > 1*time.Second {
			attrs = append(attrs, slog.String("slow_query", "true"))
			slog.LogAttrs(ctx, slog.LevelWarn, "Slow Database Query", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelDebug, "Database Query", attrs...)
		}
	}
}

func LogDatabaseConnection(ctx context.Context, dsn string, operation string, err error) {
	maskedDSN := maskPassword(dsn)

	attrs := []slog.Attr{
		slog.String("type", "database_connection"),
		slog.String("dsn", maskedDSN),
		slog.String("operation", operation),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Database Connection Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "Database Connection", attrs...)
	}
}

func LogRedisOperation(ctx context.Context, operation, key string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("type", "redis_operation"),
		slog.String("operation", operation),
		slog.String("key", key),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Redis Operation Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, "Redis Operation", attrs...)
	}
}

func LogRedisCacheHit(ctx context.Context, key string, hit bool, duration time.Duration) {
	attrs := []slog.Attr{
		slog.String("type", "cache_event"),
		slog.String("key", key),
		slog.Bool("cache_hit", hit),
		slog.Duration("duration", duration),
	}

	if hit {
		slog.LogAttrs(ctx, slog.LevelDebug, "Cache Hit", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, "Cache Miss", attrs...)
	}
}

func LogError(ctx context.Context, err error, operation string, additionalFields ...slog.Attr) {
	attrs := []slog.Attr{
		slog.String("type", "error"),
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	}
	attrs = append(attrs, additionalFields...)

	slog.LogAttrs(ctx, slog.LevelError, "Operation Error", attrs...)
}

func LogTaskOperation(ctx context.Context, operation, taskID string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("type", "task_operation"),
		slog.String("operation", operation),
		slog.String("task_id", taskID),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Task Operation Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "Task Operation", attrs...)
	}
}

func WithTaskID(taskID string) *slog.Logger {
	return slog.With(slog.String("task_id", taskID))
}

func WithRequestID(requestID string) *slog.Logger {
	return slog.With(slog.String("request_id", requestID))
}

func WithOperation(operation string) *slog.Logger {
	return slog.With(slog.String("operation", operation))
}

func maskPassword(dsn string) string {
	if dsn == "" {
		return dsn
	}

	start := strings.Index(dsn, "password=")
	if start == -1 {
		return dsn
	}

	start += len("password=")
	end := start

	for end < len(dsn) && dsn[end] != ' ' && dsn[end] != '&' {
		end++
	}

	masked := dsn[:start] + "***"
	if end < len(dsn) {
		masked += dsn[end:]
	}

	return masked
}

func LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration) {
	if duration <= threshold {
		return
	}

	attrs := []slog.Attr{
		slog.String("type", "slow_operation"),
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.Duration("threshold", threshold),
	}

	slog.LogAttrs(ctx, slog.LevelWarn, "Slow Operation Detected", attrs...)
}

func LogServiceStart(serviceName string, config map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("type", "service_lifecycle"),
		slog.String("event", "start"),
		slog.String("service", serviceName),
		slog.Any("config", config),
	}

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Service Starting", attrs...)
}

func LogServiceStop(serviceName string, reason string) {
	attrs := []slog.Attr{
		slog.String("type", "service_lifecycle"),
		slog.String("event", "stop"),
		slog.String("service", serviceName),
		slog.String("reason", reason),
	}

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Service Stopping", attrs...)
}

func LogRedisShardConnection(ctx context.Context, shardIndex int, addr string, err error) {
	attrs := []slog.Attr{
		slog.String("type", "redis_shard_connection"),
		slog.Int("shard_index", shardIndex),
		slog.String("address", addr),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Redis Shard Connection Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "Redis Shard Connected", attrs...)
	}
}

func LogRedisShardSelection(ctx context.Context, key string, shardIndex int, operation string) {
	attrs := []slog.Attr{
		slog.String("type", "redis_shard_selection"),
		slog.String("key", key),
		slog.Int("shard_index", shardIndex),
		slog.String("operation", operation),
	}

	slog.LogAttrs(ctx, slog.LevelDebug, "Redis Shard Selected", attrs...)
}

func LogCacheOperation(ctx context.Context, operation, key string, shardIndex int, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("type", "cache_operation"),
		slog.String("operation", operation),
		slog.String("key", key),
		slog.Int("shard_index", shardIndex),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelError, "Cache Operation Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, "Cache Operation Success", attrs...)
	}
}

func LogCacheInvalidation(ctx context.Context, key string, reason string, err error) {
	attrs := []slog.Attr{
		slog.String("type", "cache_invalidation"),
		slog.String("key", key),
		slog.String("reason", reason),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		slog.LogAttrs(ctx, slog.LevelWarn, "Cache Invalidation Failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, "Cache Invalidated", attrs...)
	}
}

func LogCacheStatus(ctx context.Context, enabled bool, shardCount int, ttl time.Duration) {
	attrs := []slog.Attr{
		slog.String("type", "cache_status"),
		slog.Bool("enabled", enabled),
		slog.Int("shard_count", shardCount),
		slog.Duration("default_ttl", ttl),
	}

	if enabled {
		slog.LogAttrs(ctx, slog.LevelInfo, "Cache Initialized", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "Cache Disabled", attrs...)
	}
}