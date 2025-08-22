package model

import (
	"time"

	"github.com/google/uuid"
)

// Task доменная модель (НЕ protobuf!)
type Task struct {
	ID          uuid.UUID
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Бизнес-методы
func NewTask(title, description string) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *Task) Update(title, description *string, completed *bool) {
	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if completed != nil {
		t.Completed = *completed
	}
	t.UpdatedAt = time.Now()
}
