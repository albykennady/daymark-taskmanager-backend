package repository

import (
	"context"
	"errors"

	"taskmanager/internal/task"
)

var ErrNotFound = errors.New("task not found")

type TaskRepository interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Patch(context.Context, int64, task.PatchInput) (task.Task, error)
	Delete(context.Context, int64) error
}
