package database

import (
	"context"
	"testing"

	"taskmanager/internal/repository"
)

func TestSeedTasks(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLiteMemory(ctx)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	if err := SeedTasks(ctx, db); err != nil {
		t.Fatalf("seed tasks: %v", err)
	}
	if err := SeedTasks(ctx, db); err != nil {
		t.Fatalf("seed tasks a second time: %v", err)
	}

	items, err := repository.NewSQLiteRepository(db).List(ctx)
	if err != nil {
		t.Fatalf("list seeded tasks: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("seeded task count = %d, want 3", len(items))
	}

	completedCount := 0
	for _, item := range items {
		if item.ID < 1 || item.Title == "" || item.Description == "" {
			t.Errorf("seeded task is missing required model fields: %+v", item)
		}
		if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
			t.Errorf("seeded task is missing timestamps: %+v", item)
		}
		if item.Completed {
			completedCount++
		}
	}
	if completedCount != 1 {
		t.Errorf("completed task count = %d, want 1", completedCount)
	}
}
