package task

import "time"

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Title       string
	Description string
}

type UpdateInput struct {
	Title       string
	Description string
	Completed   bool
}

type PatchInput struct {
	Title       *string
	Description *string
	Completed   *bool
}
