// pkg/logger/logger.go для DB-Service
// Адаптированная версия shared logger component
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
	Format   string `env:"LOG_FORMAT" envDefault:"json"` // добавили поддержку format
}

// SetupLogger настраивает глобальный логгер для DB-Service
func SetupLogger(cfg Config, serviceName string) error {
	// Создаем директорию для логов
	if err := os.MkdirAll(cfg.FilePath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Если имя файла не указано, используем имя сервиса
	if cfg.FileName == "" {
		cfg.FileName = fmt.Sprintf("%s.log", serviceName)
	}

	fullPath := filepath.Join(cfg.FilePath, cfg.FileName)

	// Открываем файл для логов
	logFile, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Определяем уровень логирования
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

	// Настройки handler'а
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Форматируем время в читаемом виде
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	// Создаем handler в зависимости от формата
	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(logFile, opts)
	} else {
		handler = slog.NewJSONHandler(logFile, opts)
	}

	// Создаем логгер с базовыми полями
	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	// Устанавливаем как глобальный логгер
	slog.SetDefault(logger)

	return nil
}

// ====== HTTP REQUEST LOGGING (для health checks) ======

// LogHTTPRequest логирует HTTP запросы (для health endpoints)
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

// ====== GRPC REQUEST LOGGING ======

// LogGRPCRequest логирует входящие gRPC запросы
func LogGRPCRequest(ctx context.Context, method string, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

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

// LogGRPCCall логирует исходящие gRPC вызовы (если DB-Service вызывает другие сервисы)
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

// ====== DATABASE LOGGING ======

// LogDatabaseQuery логирует запросы к PostgreSQL
func LogDatabaseQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

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
		// Логируем медленные запросы как WARNING
		if duration > 1*time.Second {
			attrs = append(attrs, slog.String("slow_query", "true"))
			slog.LogAttrs(ctx, slog.LevelWarn, "Slow Database Query", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelDebug, "Database Query", attrs...)
		}
	}
}

// LogDatabaseConnection логирует события подключения к БД
func LogDatabaseConnection(ctx context.Context, dsn string, operation string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Маскируем пароль в DSN для логов
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

// ====== REDIS LOGGING ======

// LogRedisOperation логирует операции с Redis
func LogRedisOperation(ctx context.Context, operation, key string, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

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

// LogRedisCacheHit логирует попадания/промахи кэша
func LogRedisCacheHit(ctx context.Context, key string, hit bool, duration time.Duration) {
	if ctx == nil {
		ctx = context.Background()
	}

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

// ====== GENERAL ERROR LOGGING ======

// LogError логирует общие ошибки
func LogError(ctx context.Context, err error, operation string, additionalFields ...slog.Attr) {
	if ctx == nil {
		ctx = context.Background()
	}

	attrs := []slog.Attr{
		slog.String("type", "error"),
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	}
	attrs = append(attrs, additionalFields...)

	slog.LogAttrs(ctx, slog.LevelError, "Operation Error", attrs...)
}

// ====== BUSINESS LOGIC LOGGING ======

// LogTaskOperation логирует операции с задачами
func LogTaskOperation(ctx context.Context, operation, taskID string, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

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

// ====== HELPER FUNCTIONS ======

// WithTaskID добавляет task ID к логгеру
func WithTaskID(taskID string) *slog.Logger {
	return slog.With(slog.String("task_id", taskID))
}

// WithRequestID добавляет request ID к логгеру
func WithRequestID(requestID string) *slog.Logger {
	return slog.With(slog.String("request_id", requestID))
}

// WithOperation добавляет операцию к логгеру
func WithOperation(operation string) *slog.Logger {
	return slog.With(slog.String("operation", operation))
}

// ====== UTILITY FUNCTIONS ======

// maskPassword маскирует пароль в строке подключения к БД
func maskPassword(dsn string) string {
	// Простая замена password=xxx на password=***
	// В production можно использовать более сложную логику
	if dsn == "" {
		return dsn
	}

	// Ищем паттерн password=...
	start := strings.Index(dsn, "password=")
	if start == -1 {
		return dsn
	}

	start += len("password=")
	end := start

	// Ищем конец пароля (пробел или конец строки)
	for end < len(dsn) && dsn[end] != ' ' && dsn[end] != '&' {
		end++
	}

	// Заменяем пароль на ***
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

func LogServiceStart(serviceName, version string, config map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("type", "service_lifecycle"),
		slog.String("event", "start"),
		slog.String("service", serviceName),
		slog.String("version", version),
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
