package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/greenbrown932/fire-pmaas/pkg/api"
	firemiddleware "github.com/greenbrown932/fire-pmaas/pkg/middleware"
)

func main() {
	// --- Database Migration ---
	runMigrations()

	// --- Server Setup ---
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	if err := firemiddleware.InitOIDC(); err != nil {
		log.Fatalf("Failed to initialize OIDC: %v", err)
	}

	api.RegisterRoutes(r)

	log.Println("Starting server on port :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func runMigrations() {
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDb := os.Getenv("POSTGRES_DB")

	if postgresHost == "" || postgresPort == "" || postgresUser == "" || postgresPassword == "" || postgresDb == "" {
		log.Println("Database environment variables not set, skipping migrations.")
		return
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb)

	// Path to migrations is relative to the Docker container's WORKDIR
	migrationsPath := "file://db/migrations"

	var m *migrate.Migrate
	var err error

	// Retry connecting to the database for migrations
	for i := range 10 {
		m, err = migrate.New(migrationsPath, databaseURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database for migration (attempt %d): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not initialize migrate instance: %v", err)
	}

	log.Println("Running database migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while running migrations: %v", err)
	}

	log.Println("Database migrations finished successfully.")
}
