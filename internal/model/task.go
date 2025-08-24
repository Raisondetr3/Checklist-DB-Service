package model

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(title, description string) *Task {
	return &Task{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
