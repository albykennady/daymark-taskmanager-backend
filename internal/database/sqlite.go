package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSQLiteMemory(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open in-memory SQLite database: %w", err)
	}

	// A plain :memory: database is per connection, so keep this pool on one connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping in-memory SQLite database: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize SQLite schema: %w", err)
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`
	_, err := db.ExecContext(ctx, schema)
	return err
}
