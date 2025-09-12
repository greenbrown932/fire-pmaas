package main

import (
	"fmt"
	"log"
	"os"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
)

func main() {
	// Set environment variables
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "pmaas_user")
	os.Setenv("POSTGRES_PASSWORD", "pmaas_pass")
	os.Setenv("POSTGRES_DB", "pmaas_dev")

	// Initialize database connection
	db.InitDB()

	// Simple query to check roles
	rows, err := db.DB.Query("SELECT id, name, display_name FROM roles ORDER BY name")
	if err != nil {
		log.Fatalf("Failed to query roles: %v", err)
	}
	defer rows.Close()

	fmt.Println("Roles in database:")
	for rows.Next() {
		var id int
		var name, displayName string
		if err := rows.Scan(&id, &name, &displayName); err != nil {
			log.Fatalf("Failed to scan role: %v", err)
		}
		fmt.Printf("  %d: %s (%s)\n", id, name, displayName)
	}

	// Check if any users exist
	var userCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	fmt.Printf("\nTotal users in database: %d\n", userCount)

	// Check user_roles table
	var userRoleCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM user_roles").Scan(&userRoleCount)
	if err != nil {
		log.Fatalf("Failed to count user_roles: %v", err)
	}
	fmt.Printf("Total user-role assignments: %d\n", userRoleCount)
}
