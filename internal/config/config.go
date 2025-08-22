package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

type ServerConfig struct {
	GRPCPort     string        `yaml:"grpc_port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	FilePath string `yaml:"file_path"`
	FileName string `yaml:"file_name"`
}

type DatabaseConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Name         string        `yaml:"name"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	QueryTimeout time.Duration `yaml:"query_timeout"`
}

type RedisConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	TTL          time.Duration `yaml:"ttl"`
}

func Load() (*Config, error) {
	cfg, err := loadConfigFile("configs/config.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	overrideFromEnv(cfg)

	return cfg, nil
}

func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func overrideFromEnv(cfg *Config) {
	if grpcPort := os.Getenv("DB_GRPC_PORT"); grpcPort != "" {
		cfg.Server.GRPCPort = grpcPort
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}
	if path := os.Getenv("LOG_FILE_PATH"); path != "" {
		cfg.Logging.FilePath = path
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}

	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Database.Port = port
		}
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}

	if enabledStr := os.Getenv("REDIS_ENABLED"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			cfg.Redis.Enabled = enabled
		}
	}
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Redis.Port = port
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			cfg.Redis.DB = db
		}
	}
}

func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
	)
}

func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) IsRedisEnabled() bool {
	return c.Redis.Enabled
}
