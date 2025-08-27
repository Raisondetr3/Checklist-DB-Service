package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"strings"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/model"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisCache interface {
	SetTask(ctx context.Context, task *model.Task, ttl time.Duration) error
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	SetTaskList(ctx context.Context, tasks []*model.Task, ttl time.Duration) error
	GetTaskList(ctx context.Context) ([]*model.Task, error)
	InvalidateTaskList(ctx context.Context) error
	
	Ping(ctx context.Context) error
	Close() error
}

type redisCache struct {
	clients []redis.Cmdable
	enabled bool
}

func NewRedisCache(urls []string, password string, db int, enabled bool) (RedisCache, error) {
	ctx := context.Background()
	
	if !enabled {
		logger.LogCacheStatus(ctx, false, 0, 0)
		return &redisCache{enabled: false}, nil
	}

	if len(urls) == 0 {
		return nil, errors.New("redis URLs cannot be empty when Redis is enabled")
	}

	clients := make([]redis.Cmdable, len(urls))
	
	for i, url := range urls {
		client := redis.NewClient(&redis.Options{
			Addr:     url,
			Password: password,
			DB:       db,
		})
		
		connCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		err := client.Ping(connCtx).Err()
		cancel()
		
		if err != nil {
			logger.LogRedisShardConnection(ctx, i, url, err)
			return nil, err
		}
		
		clients[i] = client
		logger.LogRedisShardConnection(ctx, i, url, nil)
	}

	logger.LogCacheStatus(ctx, true, len(urls), 0) 
	
	return &redisCache{
		clients: clients,
		enabled: true,
	}, nil
}

func (r *redisCache) getShardIndex(key string) int {
	if len(r.clients) == 1 {
		return 0
	}
	
	hash := crc32.ChecksumIEEE([]byte(key))
	return int(hash) % len(r.clients)
}

func (r *redisCache) getClient(key string) redis.Cmdable {
	if !r.enabled || len(r.clients) == 0 {
		return nil
	}
	
	index := r.getShardIndex(key)
	return r.clients[index]
}

func (r *redisCache) SetTask(ctx context.Context, task *model.Task, ttl time.Duration) error {
	if !r.enabled {
		return nil
	}

	key := r.taskKey(task.ID)
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return errors.New("no Redis client available")
	}

	start := time.Now()
	logger.LogRedisShardSelection(ctx, key, shardIndex, "SET")

	data, err := json.Marshal(task)
	if err != nil {
		logger.LogCacheOperation(ctx, "SET", key, shardIndex, time.Since(start), err)
		return err
	}

	err = client.Set(ctx, key, data, ttl).Err()
	duration := time.Since(start)
	
	logger.LogCacheOperation(ctx, "SET", key, shardIndex, duration, err)
	
	return err
}

func (r *redisCache) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	if !r.enabled {
		return nil, errors.New("cache disabled")
	}

	key := r.taskKey(id)
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return nil, errors.New("no Redis client available")
	}

	start := time.Now()
	logger.LogRedisShardSelection(ctx, key, shardIndex, "GET")

	data, err := client.Get(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		if err == redis.Nil {
			logger.LogRedisCacheHit(ctx, key, false, duration)
			return nil, errors.New("task not found in cache")
		}
		logger.LogCacheOperation(ctx, "GET", key, shardIndex, duration, err)
		return nil, err
	}

	logger.LogRedisCacheHit(ctx, key, true, duration)

	var task model.Task
	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		logger.LogCacheOperation(ctx, "GET", key, shardIndex, duration, err)
		return nil, err
	}

	logger.LogCacheOperation(ctx, "GET", key, shardIndex, duration, nil)
	return &task, nil
}

func (r *redisCache) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if !r.enabled {
		return nil
	}

	key := r.taskKey(id)
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return errors.New("no Redis client available")
	}

	start := time.Now()
	logger.LogRedisShardSelection(ctx, key, shardIndex, "DELETE")

	err := client.Del(ctx, key).Err()
	duration := time.Since(start)
	
	logger.LogCacheOperation(ctx, "DELETE", key, shardIndex, duration, err)
	return err
}

func (r *redisCache) SetTaskList(ctx context.Context, tasks []*model.Task, ttl time.Duration) error {
	if !r.enabled {
		return nil
	}

	key := r.taskListKey()
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return errors.New("no Redis client available")
	}

	start := time.Now()
	logger.LogRedisShardSelection(ctx, key, shardIndex, "SET_LIST")

	data, err := json.Marshal(tasks)
	if err != nil {
		logger.LogCacheOperation(ctx, "SET_LIST", key, shardIndex, time.Since(start), err)
		return err
	}

	err = client.Set(ctx, key, data, ttl).Err()
	duration := time.Since(start)
	
	logger.LogCacheOperation(ctx, "SET_LIST", key, shardIndex, duration, err)
	return err
}

func (r *redisCache) GetTaskList(ctx context.Context) ([]*model.Task, error) {
	if !r.enabled {
		return nil, errors.New("cache disabled")
	}

	key := r.taskListKey()
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return nil, errors.New("no Redis client available")
	}

	start := time.Now()
	logger.LogRedisShardSelection(ctx, key, shardIndex, "GET_LIST")

	data, err := client.Get(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		if err == redis.Nil {
			logger.LogRedisCacheHit(ctx, key, false, duration)
			return nil, errors.New("task list not found in cache")
		}
		logger.LogCacheOperation(ctx, "GET_LIST", key, shardIndex, duration, err)
		return nil, err
	}

	logger.LogRedisCacheHit(ctx, key, true, duration)

	var tasks []*model.Task
	err = json.Unmarshal([]byte(data), &tasks)
	if err != nil {
		logger.LogCacheOperation(ctx, "GET_LIST", key, shardIndex, duration, err)
		return nil, err
	}

	logger.LogCacheOperation(ctx, "GET_LIST", key, shardIndex, duration, nil)
	return tasks, nil
}

func (r *redisCache) InvalidateTaskList(ctx context.Context) error {
	if !r.enabled {
		return nil
	}

	key := r.taskListKey()
	shardIndex := r.getShardIndex(key)
	client := r.getClient(key)
	if client == nil {
		return errors.New("no Redis client available")
	}

	start := time.Now()
	err := client.Del(ctx, key).Err()
	duration := time.Since(start)
	
	logger.LogCacheInvalidation(ctx, key, "task_list_changed", err)
	logger.LogCacheOperation(ctx, "DELETE_LIST", key, shardIndex, duration, err)
	
	return err
}

func (r *redisCache) Ping(ctx context.Context) error {
	if !r.enabled {
		return nil
	}

	for i, client := range r.clients {
		start := time.Now()
		err := client.Ping(ctx).Err()
		duration := time.Since(start)
		
		if err != nil {
			logger.LogCacheOperation(ctx, "PING", "health_check", i, duration, err)
			return err
		}
		
		logger.LogCacheOperation(ctx, "PING", "health_check", i, duration, nil)
	}
	return nil
}

func (r *redisCache) Close() error {
	if !r.enabled {
		return nil
	}

	ctx := context.Background()
	var lastErr error
	
	for i, client := range r.clients {
		if redisClient, ok := client.(*redis.Client); ok {
			if err := redisClient.Close(); err != nil {
				logger.LogError(ctx, err, "close_redis_shard", 
					slog.Int("shard_index", i))
				lastErr = err
			}
		}
	}
	return lastErr
}

func (r *redisCache) taskKey(id uuid.UUID) string {
	return fmt.Sprintf("task:%s", id.String())
}

func (r *redisCache) taskListKey() string {
	return "tasks:list"
}

func ParseRedisURLs(urls string) []string {
	if urls == "" {
		return []string{}
	}
	
	urlList := strings.Split(urls, ",")
	result := make([]string, 0, len(urlList))
	
	for _, url := range urlList {
		trimmed := strings.TrimSpace(url)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}