package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"taskmanager/internal/task"
)

func SeedTasks(ctx context.Context, db *sql.DB) error {
	now := time.Now().UTC()
	tasks := []task.Task{
		{
			ID:          1,
			Title:       "Plan weekly sprint",
			Description: "Review priorities and prepare the sprint backlog.",
			Completed:   false,
			CreatedAt:   now.Add(-72 * time.Hour),
			UpdatedAt:   now.Add(-24 * time.Hour),
		},
		{
			ID:          2,
			Title:       "Update API documentation",
			Description: "Document the task CRUD endpoints and request examples.",
			Completed:   true,
			CreatedAt:   now.Add(-48 * time.Hour),
			UpdatedAt:   now.Add(-12 * time.Hour),
		},
		{
			ID:          3,
			Title:       "Add repository tests",
			Description: "Cover task creation, updates, and not-found behavior.",
			Completed:   false,
			CreatedAt:   now.Add(-6 * time.Hour),
			UpdatedAt:   now.Add(-6 * time.Hour),
		},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin task seed transaction: %w", err)
	}
	defer tx.Rollback()

	for _, item := range tasks {
		_, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO tasks
				(id, title, description, completed, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			item.ID,
			item.Title,
			item.Description,
			item.Completed,
			item.CreatedAt,
			item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("seed task %q: %w", item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit task seed transaction: %w", err)
	}
	return nil
}
