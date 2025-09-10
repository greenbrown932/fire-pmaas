package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB is the database connection instance.
var DB *sql.DB

// InitDB initializes the database connection.
func InitDB() {
	var err error // Variable to hold errors

	// Retrieve database connection details from environment variables.
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDb := os.Getenv("POSTGRES_DB")

	// Check if all required environment variables are set.
	if postgresHost == "" || postgresPort == "" || postgresUser == "" || postgresPassword == "" || postgresDb == "" {
		log.Fatal("One or more PostgreSQL environment variables not set") // Log fatal error and exit if any variable is missing.
	}

	// Construct the database connection URL.
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb) // Format the connection string.

	// Open a database connection.
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Test the database connection.
	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	SeedDatabase()
}

// SeedDatabase seeds the database with initial data.
func SeedDatabase() {
	log.Println("Seeding database...")

	// Example properties
	properties := []struct {
		Name         string
		Address      string
		PropertyType string
	}{
		{"Sunset Apartments", "123 Main St, Anytown", "Apartment Building"},
		{"Oakwood Villas", "456 Elm St, Anytown", "Townhouse Complex"},
		{"Pine Ridge Cottage", "789 Oak St, Anytown", "Single Family Home"},
	}

	for _, p := range properties {
		_, err := DB.Exec(`
			INSERT INTO properties (name, address, property_type)
			VALUES ($1, $2, $3)
		`, p.Name, p.Address, p.PropertyType)
		if err != nil {
			log.Printf("Failed to seed property %s: %v", p.Name, err)
		}
	}

	log.Println("Database seeding complete.")
}
