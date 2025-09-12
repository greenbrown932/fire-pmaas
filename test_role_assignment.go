package main

import (
	"fmt"
	"log"
	"os"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

// Mock the assignRolesFromKeycloak function for testing
func assignRolesFromKeycloak(userID int, keycloakRoles []string) {
	fmt.Printf("DEBUG: assignRolesFromKeycloak called with userID=%d, roles=%v\n", userID, keycloakRoles)

	// Role mapping from Keycloak realm roles to application roles
	roleMapping := map[string]string{
		"admin":            "admin",
		"property_manager": "property_manager",
		"tenant":           "tenant",
		"viewer":           "viewer",
	}

	// First, remove all existing roles for this user to ensure clean sync
	for _, appRole := range roleMapping {
		appRoleRecord, err := models.GetRoleByName(appRole)
		if err == nil {
			err = models.RemoveRole(userID, appRoleRecord.ID)
			if err != nil {
				fmt.Printf("DEBUG: Failed to remove role %s from user %d: %v\n", appRole, userID, err)
			}
		} else {
			fmt.Printf("DEBUG: Failed to get role %s: %v\n", appRole, err)
		}
	}

	// Assign roles based on Keycloak roles
	assignedCount := 0
	for _, keycloakRole := range keycloakRoles {
		if appRole, exists := roleMapping[keycloakRole]; exists {
			fmt.Printf("DEBUG: Mapping Keycloak role '%s' to app role '%s'\n", keycloakRole, appRole)
			appRoleRecord, err := models.GetRoleByName(appRole)
			if err == nil {
				err = models.AssignRole(userID, appRoleRecord.ID, nil)
				if err != nil {
					fmt.Printf("DEBUG: Failed to assign role %s to user %d: %v\n", appRole, userID, err)
				} else {
					fmt.Printf("DEBUG: Successfully assigned role %s to user %d\n", appRole, userID)
					assignedCount++
				}
			} else {
				fmt.Printf("DEBUG: Failed to get role %s: %v\n", appRole, err)
			}
		} else {
			fmt.Printf("DEBUG: Keycloak role '%s' not mapped to any app role\n", keycloakRole)
		}
	}

	// If no mapped roles were found, assign default tenant role
	if assignedCount == 0 {
		fmt.Printf("DEBUG: No roles assigned, assigning default 'tenant' role\n")
		defaultRole, err := models.GetRoleByName("tenant")
		if err == nil {
			err = models.AssignRole(userID, defaultRole.ID, nil)
			if err != nil {
				fmt.Printf("DEBUG: Failed to assign default tenant role: %v\n", err)
			} else {
				fmt.Printf("DEBUG: Successfully assigned default tenant role\n")
			}
		} else {
			fmt.Printf("DEBUG: Failed to get default tenant role: %v\n", err)
		}
	}

	fmt.Printf("DEBUG: assignRolesFromKeycloak completed, assigned %d roles\n", assignedCount)
}

func main() {
	// Set environment variables
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "pmaas_user")
	os.Setenv("POSTGRES_PASSWORD", "pmaas_pass")
	os.Setenv("POSTGRES_DB", "pmaas_dev")

	// Initialize database connection
	db.InitDB()

	// Create a test user
	testUser := &models.User{
		KeycloakID:    models.NullString("test-admin-keycloak-id"),
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

	// Test role assignment with admin role
	fmt.Println("\n=== Testing role assignment with admin role ===")
	assignRolesFromKeycloak(testUser.ID, []string{"admin"})

	// Check if roles were assigned
	userWithRoles, err := models.GetUserByID(testUser.ID)
	if err != nil {
		log.Fatalf("Failed to reload user: %v", err)
	}

	fmt.Printf("\nUser now has %d roles:\n", len(userWithRoles.Roles))
	for _, role := range userWithRoles.Roles {
		fmt.Printf("  - %s\n", role.Name)
	}

	// Test role checking
	fmt.Printf("\nRole checking results:\n")
	fmt.Printf("User HasRole('admin'): %v\n", userWithRoles.HasRole("admin"))
	fmt.Printf("User HasRole('tenant'): %v\n", userWithRoles.HasRole("tenant"))
	fmt.Printf("User HasAnyRole('admin', 'property_manager'): %v\n", userWithRoles.HasAnyRole("admin", "property_manager"))
	fmt.Printf("User HasAnyRole('viewer', 'tenant'): %v\n", userWithRoles.HasAnyRole("viewer", "tenant"))
}
