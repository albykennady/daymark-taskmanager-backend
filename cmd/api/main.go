package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"taskmanager/internal/database"
	"taskmanager/internal/handler"
	"taskmanager/internal/repository"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := database.OpenSQLiteMemory(ctx)
	cancel()
	if err != nil {
		log.Fatalf("connect to task database: %v", err)
	}
	defer db.Close()

	if err := database.SeedTasks(context.Background(), db); err != nil {
		log.Fatalf("populate task database with sample data: %v", err)
	}

	taskRepository := repository.NewSQLiteRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepository)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	taskHandler.RegisterRoutes(router)

	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	log.Printf("task API listening on %s", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatal(err)
	}
}
