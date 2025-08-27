package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/cache"
	"github.com/Raisondetr3/checklist-db-service/internal/model"
	"github.com/google/uuid"
)

type cachedTaskRepository struct {
	repo  TaskRepository
	cache cache.RedisCache
	ttl   time.Duration
}

func NewCachedTaskRepository(repo TaskRepository, cache cache.RedisCache, ttl time.Duration) TaskRepository {
	return &cachedTaskRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
	}
}

func (r *cachedTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	createdTask, err := r.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	if err := r.cache.SetTask(ctx, createdTask, r.ttl); err != nil {
		slog.Warn("Failed to cache created task", 
			slog.String("task_id", createdTask.ID.String()),
			slog.String("error", err.Error()))
	}

	if err := r.cache.InvalidateTaskList(ctx); err != nil {
		slog.Warn("Failed to invalidate task list cache", 
			slog.String("error", err.Error()))
	}

	return createdTask, nil
}

func (r *cachedTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task, err := r.cache.GetTask(ctx, id)
	if err == nil {
		slog.Debug("Task found in cache", slog.String("task_id", id.String()))
		return task, nil
	}

	slog.Debug("Task not in cache, fetching from database", 
		slog.String("task_id", id.String()))
	
	task, err = r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cache.SetTask(ctx, task, r.ttl); err != nil {
		slog.Warn("Failed to cache retrieved task", 
			slog.String("task_id", task.ID.String()),
			slog.String("error", err.Error()))
	}

	return task, nil
}

func (r *cachedTaskRepository) Update(ctx context.Context, task *model.Task) (*model.Task, error) {
	updatedTask, err := r.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	if err := r.cache.SetTask(ctx, updatedTask, r.ttl); err != nil {
		slog.Warn("Failed to cache updated task", 
			slog.String("task_id", updatedTask.ID.String()),
			slog.String("error", err.Error()))
	}

	if err := r.cache.InvalidateTaskList(ctx); err != nil {
		slog.Warn("Failed to invalidate task list cache", 
			slog.String("error", err.Error()))
	}

	return updatedTask, nil
}

func (r *cachedTaskRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	err := r.repo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.cache.DeleteTask(ctx, id); err != nil {
		slog.Warn("Failed to delete task from cache", 
			slog.String("task_id", id.String()),
			slog.String("error", err.Error()))
	}

	if err := r.cache.InvalidateTaskList(ctx); err != nil {
		slog.Warn("Failed to invalidate task list cache", 
			slog.String("error", err.Error()))
	}

	return nil
}

func (r *cachedTaskRepository) List(ctx context.Context) ([]*model.Task, error) {
	tasks, err := r.cache.GetTaskList(ctx)
	if err == nil {
		slog.Debug("Task list found in cache", slog.Int("count", len(tasks)))
		return tasks, nil
	}

	slog.Debug("Task list not in cache, fetching from database")
	
	tasks, err = r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	listTTL := 60 * time.Second 
	if err := r.cache.SetTaskList(ctx, tasks, listTTL); err != nil {
		slog.Warn("Failed to cache task list", 
			slog.Int("count", len(tasks)),
			slog.String("error", err.Error()))
	}

	return tasks, nil
}