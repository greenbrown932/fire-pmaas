package main

import (
	"fmt"
	"log"      // For logging errors and other information
	"net/http" // For creating HTTP servers
	"os"       // For accessing environment variables
	"time"     // For time-related operations, like sleeping

	// Third-party libraries
	"github.com/go-chi/chi"                                             // Lightweight HTTP router
	chimiddleware "github.com/go-chi/chi/middleware"                    // Useful middleware for Chi
	"github.com/golang-migrate/migrate/v4"                              // Database migration tool
	_ "github.com/golang-migrate/migrate/v4/database/postgres"          // PostgreSQL driver for migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"                // File source driver for migrate
	"github.com/greenbrown932/fire-pmaas/pkg/api"                       // API route definitions
	"github.com/greenbrown932/fire-pmaas/pkg/db"                        // Database initialization and connection
	firemiddleware "github.com/greenbrown932/fire-pmaas/pkg/middleware" // Custom middleware
)

func main() {
	runMigrations()
	db.InitDB()

	// Initialize OIDC provider
	if err := firemiddleware.InitOIDC(); err != nil {
		log.Fatalf("Failed to initialize OIDC: %v", err)
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)    // Log API requests
	r.Use(chimiddleware.Recoverer) // Recover from panics

	api.RegisterRoutes(r)

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
	for i := 0; i < 10; i++ {
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
