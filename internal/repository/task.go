package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/model"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) (*model.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) (*model.Task, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*model.Task, error)
}

type taskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	start := time.Now()
	q := `
		INSERT INTO tasks (id, title, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, completed, created_at, updated_at
	`

	var createdTask model.Task
	err := r.db.QueryRow(ctx, q,
		task.ID, task.Title, task.Description, task.Completed,
		task.CreatedAt, task.UpdatedAt,
	).Scan(
		&createdTask.ID, &createdTask.Title, &createdTask.Description,
		&createdTask.Completed, &createdTask.CreatedAt, &createdTask.UpdatedAt,
	)

	duration := time.Since(start)

	if err != nil {
		r.logCriticalDBError(ctx, "create_task", q, duration, err)
		return nil, HandlePgxError("create_task", err)
	}

	r.logSlowQuery(ctx, "create_task", duration)

	return &createdTask, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	start := time.Now()
	q := `SELECT id, title, description, completed, created_at, updated_at FROM tasks WHERE id = $1`

	var task model.Task
	err := r.db.QueryRow(ctx, q, id).Scan(
		&task.ID, &task.Title, &task.Description,
		&task.Completed, &task.CreatedAt, &task.UpdatedAt,
	)

	duration := time.Since(start)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logCriticalDBError(ctx, "get_task_by_id", q, duration, err)
		}
		return nil, HandlePgxError("get_task_by_id", err)
	}

	r.logSlowQuery(ctx, "get_task_by_id", duration)
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *model.Task) (*model.Task, error) {
	start := time.Now()
	q := `
		UPDATE tasks 
		SET title = $2, description = $3, completed = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, completed, created_at, updated_at
	`

	var updatedTask model.Task
	err := r.db.QueryRow(ctx, q, task.ID, task.Title, task.Description, task.Completed).Scan(
		&updatedTask.ID, &updatedTask.Title, &updatedTask.Description,
		&updatedTask.Completed, &updatedTask.CreatedAt, &updatedTask.UpdatedAt,
	)

	duration := time.Since(start)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logCriticalDBError(ctx, "update_task", q, duration, err)
		}
		return nil, HandlePgxError("update_task", err)
	}

	r.logSlowQuery(ctx, "update_task", duration)
	return &updatedTask, nil
}

func (r *taskRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	q := `DELETE FROM tasks WHERE id = $1`

	commandTag, err := r.db.Exec(ctx, q, id)
	duration := time.Since(start)

	if err != nil {
		r.logCriticalDBError(ctx, "delete_task", q, duration, err)
		return HandlePgxError("delete_task", err)
	}

	if commandTag.RowsAffected() == 0 {
		return WrapError("delete_task", ErrTaskNotFound)
	}

	r.logSlowQuery(ctx, "delete_task", duration)
	return nil
}

func (r *taskRepository) List(ctx context.Context) ([]*model.Task, error) {
	start := time.Now()
	q := `SELECT id, title, description, completed, created_at, updated_at FROM tasks ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		duration := time.Since(start)
		r.logCriticalDBError(ctx, "list_tasks", q, duration, err)
		return nil, HandlePgxError("list_tasks", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description,
			&task.Completed, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			duration := time.Since(start)
			r.logCriticalDBError(ctx, "list_tasks_scan", "", duration, err)
			return nil, HandlePgxError("list_tasks_scan", err)
		}
		tasks = append(tasks, &task)
	}

	duration := time.Since(start)
	if err = rows.Err(); err != nil {
		r.logCriticalDBError(ctx, "list_tasks_iteration", "", duration, err)
		return nil, HandlePgxError("list_tasks_iteration", err)
	}

	r.logSlowQuery(ctx, "list_tasks", duration)
	return tasks, nil
}

func (r *taskRepository) logCriticalDBError(ctx context.Context, operation, query string, duration time.Duration, err error) {
	args := []interface{}{}
	logger.LogDatabaseQuery(ctx, query, args, duration, err)

	slog.ErrorContext(ctx, "Critical database error",
		slog.String("operation", operation),
		slog.String("error", err.Error()),
		slog.Duration("duration", duration),
	)
}

func (r *taskRepository) logSlowQuery(ctx context.Context, operation string, duration time.Duration) {
	threshold := 500 * time.Millisecond
	if duration > threshold {
		logger.LogSlowOperation(ctx, operation, duration, threshold)
	}
}
