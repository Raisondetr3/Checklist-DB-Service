package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Logging  LoggingConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	HTTPPort     string
	GRPCPort     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type LoggingConfig struct {
	Level    string
	FilePath string
	FileName string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Password string
	DB       int
}

func Load() (*Config, error) {
	cfg := &Config{
		// Дефолтные значения
		Server: ServerConfig{
			HTTPPort:     getEnv("HTTP_PORT", "8081"),
			GRPCPort:     getEnv("GRPC_PORT", "9090"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Logging: LoggingConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			FilePath: getEnv("LOG_FILE_PATH", "logs"),
			FileName: getEnv("LOG_FILE_NAME", "db-service.log"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "checklist_db"),
			User:     getEnv("DB_USER", "checklist_user"),
			Password: getEnv("DB_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Enabled:  getEnvBool("REDIS_ENABLED", false),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}

	return cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}