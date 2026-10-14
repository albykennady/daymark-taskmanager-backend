package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"taskmanager/internal/task"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	now := time.Now().UTC()
	item := task.Task{
		Title:       input.Title,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, title, description, completed, created_at, updated_at)
		VALUES (NULL, ?, ?, ?, ?, ?)`,
		item.Title, item.Description, item.Completed, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return task.Task{}, err
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return task.Task{}, err
	}
	return item, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]task.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]task.Task, 0)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SQLiteRepository) Get(ctx context.Context, id int64) (task.Task, error) {
	item, err := scanTask(r.db.QueryRowContext(ctx, `
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, ErrNotFound
	}
	return item, err
}

func (r *SQLiteRepository) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = ?, description = ?, completed = ?, updated_at = ?
		WHERE id = ?`,
		input.Title, input.Description, input.Completed, time.Now().UTC(), id,
	)
	if err != nil {
		return task.Task{}, err
	}
	if err := requireAffectedRow(result); err != nil {
		return task.Task{}, err
	}
	return r.Get(ctx, id)
}

func (r *SQLiteRepository) Patch(ctx context.Context, id int64, input task.PatchInput) (task.Task, error) {
	var assignments []string
	var args []any
	if input.Title != nil {
		assignments = append(assignments, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Description != nil {
		assignments = append(assignments, "description = ?")
		args = append(args, *input.Description)
	}
	if input.Completed != nil {
		assignments = append(assignments, "completed = ?")
		args = append(args, *input.Completed)
	}
	assignments = append(assignments, "updated_at = ?")
	args = append(args, time.Now().UTC(), id)

	query := "UPDATE tasks SET " + strings.Join(assignments, ", ") + " WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return task.Task{}, err
	}
	if err := requireAffectedRow(result); err != nil {
		return task.Task{}, err
	}
	return r.Get(ctx, id)
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	return requireAffectedRow(result)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (task.Task, error) {
	var item task.Task
	err := row.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Completed,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func requireAffectedRow(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
