package main

import (
	"fmt"
	"log"
	"os"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
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

	// Test creating a user with admin role
	testUser := &models.User{
		KeycloakID:    models.NullString("test-keycloak-id"),
		Username:      "testadmin",
		Email:         "testadmin@test.com",
		FirstName:     "Test",
		LastName:      "Admin",
		EmailVerified: true,
		Status:        "active",
	}

	// Create user
	if err := models.CreateUser(testUser); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user with ID: %d\n", testUser.ID)

	// Assign admin role
	adminRole, err := models.GetRoleByName("admin")
	if err != nil {
		log.Fatalf("Failed to get admin role: %v", err)
	}

	fmt.Printf("Found admin role: %s (ID: %d)\n", adminRole.Name, adminRole.ID)

	// Assign role to user
	if err := models.AssignRole(testUser.ID, adminRole.ID, nil); err != nil {
		log.Fatalf("Failed to assign role: %v", err)
	}

	fmt.Printf("Assigned admin role to user\n")

	// Reload user with roles
	userWithRoles, err := models.GetUserByID(testUser.ID)
	if err != nil {
		log.Fatalf("Failed to reload user: %v", err)
	}

	fmt.Printf("User has %d roles:\n", len(userWithRoles.Roles))
	for _, role := range userWithRoles.Roles {
		fmt.Printf("  - %s\n", role.Name)
	}

	// Test role checking
	fmt.Printf("User HasRole('admin'): %v\n", userWithRoles.HasRole("admin"))
	fmt.Printf("User HasRole('tenant'): %v\n", userWithRoles.HasRole("tenant"))
	fmt.Printf("User HasAnyRole('admin', 'property_manager'): %v\n", userWithRoles.HasAnyRole("admin", "property_manager"))
}
