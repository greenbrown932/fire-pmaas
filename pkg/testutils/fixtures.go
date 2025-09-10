package testutils

import (
	"database/sql"
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

// UserFixtures contains pre-defined test users for different scenarios
type UserFixtures struct {
	AdminUser       *models.User
	PropertyManager *models.User
	TenantUser      *models.User
	ViewerUser      *models.User
	InactiveUser    *models.User
	UnverifiedUser  *models.User
	MFAEnabledUser  *models.User
}

// PropertyFixtures contains pre-defined test properties
type PropertyFixtures struct {
	ApartmentBuilding *models.Property
	SingleFamilyHome  *models.Property
	TownhouseComplex  *models.Property
}

// PropertyUnitFixtures contains pre-defined test property units
type PropertyUnitFixtures struct {
	ApartmentUnit101 *models.PropertyUnit
	ApartmentUnit102 *models.PropertyUnit
	HouseMainUnit    *models.PropertyUnit
	TownhouseUnitA   *models.PropertyUnit
}

// TenantFixtures contains pre-defined test tenants
type TenantFixtures struct {
	ActiveTenant   *models.Tenant
	PendingTenant  *models.Tenant
	ArchivedTenant *models.Tenant
}

// LeaseFixtures contains pre-defined test leases
type LeaseFixtures struct {
	ActiveLease  *models.Lease
	EndedLease   *models.Lease
	PendingLease *models.Lease
}

// GetUserFixtures returns a set of test users for various scenarios
func GetUserFixtures() *UserFixtures {
	baseTime := time.Now()

	return &UserFixtures{
		AdminUser: &models.User{
			ID:            1,
			Username:      "admin",
			Email:         "admin@fire-pmaas.com",
			FirstName:     "System",
			LastName:      "Administrator",
			PhoneNumber:   models.NullString("+1-555-0001"),
			EmailVerified: true,
			MFAEnabled:    true,
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, -6, 0),
			UpdatedAt:     baseTime,
			Roles: []models.Role{
				{
					ID:          1,
					Name:        "admin",
					DisplayName: "Administrator",
					Description: models.NullString("Full system access"),
					Permissions: GetRolePermissions("admin"),
				},
			},
		},

		PropertyManager: &models.User{
			ID:            2,
			Username:      "propmanager",
			Email:         "manager@fire-pmaas.com",
			FirstName:     "Property",
			LastName:      "Manager",
			PhoneNumber:   models.NullString("+1-555-0002"),
			EmailVerified: true,
			MFAEnabled:    false,
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, -3, 0),
			UpdatedAt:     baseTime.AddDate(0, 0, -7),
			Roles: []models.Role{
				{
					ID:          2,
					Name:        "property_manager",
					DisplayName: "Property Manager",
					Description: models.NullString("Manages properties and tenants"),
					Permissions: GetRolePermissions("property_manager"),
				},
			},
		},

		TenantUser: &models.User{
			ID:            3,
			Username:      "johndoe",
			Email:         "john.doe@email.com",
			FirstName:     "John",
			LastName:      "Doe",
			PhoneNumber:   models.NullString("+1-555-0123"),
			EmailVerified: true,
			MFAEnabled:    false,
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, -1, 0),
			UpdatedAt:     baseTime.AddDate(0, 0, -3),
			Roles: []models.Role{
				{
					ID:          3,
					Name:        "tenant",
					DisplayName: "Tenant",
					Description: models.NullString("Property tenant"),
					Permissions: GetRolePermissions("tenant"),
				},
			},
		},

		ViewerUser: &models.User{
			ID:            4,
			Username:      "viewer",
			Email:         "viewer@fire-pmaas.com",
			FirstName:     "Read",
			LastName:      "Only",
			PhoneNumber:   models.NullString("+1-555-0004"),
			EmailVerified: true,
			MFAEnabled:    false,
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, -2, 0),
			UpdatedAt:     baseTime.AddDate(0, 0, -14),
			Roles: []models.Role{
				{
					ID:          4,
					Name:        "viewer",
					DisplayName: "Viewer",
					Description: models.NullString("Read-only access"),
					Permissions: GetRolePermissions("viewer"),
				},
			},
		},

		InactiveUser: &models.User{
			ID:            5,
			Username:      "inactive",
			Email:         "inactive@email.com",
			FirstName:     "Inactive",
			LastName:      "User",
			EmailVerified: true,
			MFAEnabled:    false,
			Status:        "inactive",
			CreatedAt:     baseTime.AddDate(0, -4, 0),
			UpdatedAt:     baseTime.AddDate(0, -1, 0),
			Roles:         []models.Role{},
		},

		UnverifiedUser: &models.User{
			ID:            6,
			Username:      "unverified",
			Email:         "unverified@email.com",
			FirstName:     "Unverified",
			LastName:      "User",
			EmailVerified: false,
			MFAEnabled:    false,
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, 0, -1),
			UpdatedAt:     baseTime.AddDate(0, 0, -1),
			Roles: []models.Role{
				{
					ID:          3,
					Name:        "tenant",
					DisplayName: "Tenant",
					Permissions: GetRolePermissions("tenant"),
				},
			},
		},

		MFAEnabledUser: &models.User{
			ID:            7,
			Username:      "mfauser",
			Email:         "mfa@email.com",
			FirstName:     "MFA",
			LastName:      "User",
			PhoneNumber:   models.NullString("+1-555-0007"),
			EmailVerified: true,
			MFAEnabled:    true,
			MFASecret:     models.NullString("JBSWY3DPEHPK3PXP"),
			Status:        "active",
			CreatedAt:     baseTime.AddDate(0, 0, -5),
			UpdatedAt:     baseTime.AddDate(0, 0, -1),
			Roles: []models.Role{
				{
					ID:          3,
					Name:        "tenant",
					DisplayName: "Tenant",
					Permissions: GetRolePermissions("tenant"),
				},
			},
		},
	}
}

// GetPropertyFixtures returns a set of test properties
func GetPropertyFixtures() *PropertyFixtures {
	baseTime := time.Now()

	return &PropertyFixtures{
		ApartmentBuilding: &models.Property{
			ID:           1,
			Name:         "Sunset Apartments",
			Address:      "123 Main Street, Anytown, ST 12345",
			PropertyType: "Apartment Building",
			CreatedAt:    baseTime.AddDate(0, -6, 0),
			UpdatedAt:    baseTime.AddDate(0, -1, 0),
		},

		SingleFamilyHome: &models.Property{
			ID:           2,
			Name:         "Pine Ridge Cottage",
			Address:      "789 Oak Street, Anytown, ST 12345",
			PropertyType: "Single Family Home",
			CreatedAt:    baseTime.AddDate(0, -4, 0),
			UpdatedAt:    baseTime.AddDate(0, 0, -15),
		},

		TownhouseComplex: &models.Property{
			ID:           3,
			Name:         "Oakwood Villas",
			Address:      "456 Elm Street, Anytown, ST 12345",
			PropertyType: "Townhouse Complex",
			CreatedAt:    baseTime.AddDate(0, -3, 0),
			UpdatedAt:    baseTime.AddDate(0, 0, -7),
		},
	}
}

// GetPropertyUnitFixtures returns a set of test property units
func GetPropertyUnitFixtures() *PropertyUnitFixtures {
	baseTime := time.Now()

	return &PropertyUnitFixtures{
		ApartmentUnit101: &models.PropertyUnit{
			ID:          1,
			PropertyID:  1, // Sunset Apartments
			UnitNumber:  "101",
			Bedrooms:    2,
			Bathrooms:   1,
			Description: "Ground floor unit with garden view",
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, -1, 0),
		},

		ApartmentUnit102: &models.PropertyUnit{
			ID:          2,
			PropertyID:  1, // Sunset Apartments
			UnitNumber:  "102",
			Bedrooms:    1,
			Bathrooms:   1,
			Description: "Cozy studio apartment",
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, 0, -30),
		},

		HouseMainUnit: &models.PropertyUnit{
			ID:          3,
			PropertyID:  2, // Pine Ridge Cottage
			UnitNumber:  "Main",
			Bedrooms:    3,
			Bathrooms:   2,
			Description: "Charming single family home with backyard",
			CreatedAt:   baseTime.AddDate(0, -4, 0),
			UpdatedAt:   baseTime.AddDate(0, 0, -15),
		},

		TownhouseUnitA: &models.PropertyUnit{
			ID:          4,
			PropertyID:  3, // Oakwood Villas
			UnitNumber:  "A",
			Bedrooms:    2,
			Bathrooms:   2,
			Description: "Modern townhouse with attached garage",
			CreatedAt:   baseTime.AddDate(0, -3, 0),
			UpdatedAt:   baseTime.AddDate(0, 0, -7),
		},
	}
}

// GetTenantFixtures returns a set of test tenants
func GetTenantFixtures() *TenantFixtures {
	baseTime := time.Now()

	return &TenantFixtures{
		ActiveTenant: &models.Tenant{
			ID:          1,
			FirstName:   "John",
			LastName:    "Doe",
			Email:       "john.doe@email.com",
			PhoneNumber: "+1-555-0123",
			Status:      "active",
			CreatedAt:   baseTime.AddDate(0, -1, 0),
			UpdatedAt:   baseTime.AddDate(0, 0, -3),
		},

		PendingTenant: &models.Tenant{
			ID:          2,
			FirstName:   "Jane",
			LastName:    "Smith",
			Email:       "jane.smith@email.com",
			PhoneNumber: "+1-555-0456",
			Status:      "pending",
			CreatedAt:   baseTime.AddDate(0, 0, -7),
			UpdatedAt:   baseTime.AddDate(0, 0, -7),
		},

		ArchivedTenant: &models.Tenant{
			ID:          3,
			FirstName:   "Bob",
			LastName:    "Johnson",
			Email:       "bob.johnson@email.com",
			PhoneNumber: "+1-555-0789",
			Status:      "archived",
			CreatedAt:   baseTime.AddDate(-1, 0, 0),
			UpdatedAt:   baseTime.AddDate(0, -3, 0),
		},
	}
}

// GetLeaseFixtures returns a set of test leases
func GetLeaseFixtures() *LeaseFixtures {
	baseTime := time.Now()

	return &LeaseFixtures{
		ActiveLease: &models.Lease{
			ID:          1,
			UnitID:      1, // Apartment Unit 101
			TenantID:    1, // John Doe
			StartDate:   baseTime.AddDate(0, -1, 0),
			EndDate:     baseTime.AddDate(1, -1, 0),
			MonthlyRent: 1500.00,
			Status:      "active",
			CreatedAt:   baseTime.AddDate(0, -1, -5),
			UpdatedAt:   baseTime.AddDate(0, -1, 0),
		},

		EndedLease: &models.Lease{
			ID:          2,
			UnitID:      2, // Apartment Unit 102
			TenantID:    3, // Bob Johnson
			StartDate:   baseTime.AddDate(-1, 0, 0),
			EndDate:     baseTime.AddDate(0, -3, 0),
			MonthlyRent: 1200.00,
			Status:      "ended",
			CreatedAt:   baseTime.AddDate(-1, 0, -5),
			UpdatedAt:   baseTime.AddDate(0, -3, 0),
		},

		PendingLease: &models.Lease{
			ID:          3,
			UnitID:      4, // Townhouse Unit A
			TenantID:    2, // Jane Smith
			StartDate:   baseTime.AddDate(0, 1, 0),
			EndDate:     baseTime.AddDate(1, 1, 0),
			MonthlyRent: 1800.00,
			Status:      "pending",
			CreatedAt:   baseTime.AddDate(0, 0, -7),
			UpdatedAt:   baseTime.AddDate(0, 0, -7),
		},
	}
}

// GetMaintenanceRequestFixtures returns test maintenance requests
func GetMaintenanceRequestFixtures() []map[string]interface{} {
	baseTime := time.Now()

	return []map[string]interface{}{
		{
			"id":                    1,
			"property_id":           1,
			"reported_by_tenant_id": 1,
			"description":           "Leaky faucet in kitchen sink",
			"status":                "reported",
			"priority":              "medium",
			"reported_date":         baseTime.AddDate(0, 0, -3),
			"completed_date":        nil,
			"created_at":            baseTime.AddDate(0, 0, -3),
			"updated_at":            baseTime.AddDate(0, 0, -3),
		},
		{
			"id":                    2,
			"property_id":           2,
			"reported_by_tenant_id": nil,
			"description":           "HVAC system maintenance check",
			"status":                "in_progress",
			"priority":              "low",
			"reported_date":         baseTime.AddDate(0, 0, -7),
			"completed_date":        nil,
			"created_at":            baseTime.AddDate(0, 0, -7),
			"updated_at":            baseTime.AddDate(0, 0, -1),
		},
		{
			"id":                    3,
			"property_id":           1,
			"reported_by_tenant_id": 1,
			"description":           "Broken window in living room",
			"status":                "completed",
			"priority":              "high",
			"reported_date":         baseTime.AddDate(0, 0, -14),
			"completed_date":        baseTime.AddDate(0, 0, -7),
			"created_at":            baseTime.AddDate(0, 0, -14),
			"updated_at":            baseTime.AddDate(0, 0, -7),
		},
	}
}

// GetPaymentFixtures returns test payment records
func GetPaymentFixtures() []map[string]interface{} {
	baseTime := time.Now()

	return []map[string]interface{}{
		{
			"id":             1,
			"lease_id":       1,
			"amount":         1500.00,
			"payment_date":   baseTime.AddDate(0, 0, -5),
			"payment_method": "Credit Card",
			"status":         "completed",
			"created_at":     baseTime.AddDate(0, 0, -5),
		},
		{
			"id":             2,
			"lease_id":       1,
			"amount":         1500.00,
			"payment_date":   baseTime.AddDate(0, -1, -5),
			"payment_method": "Bank Transfer",
			"status":         "completed",
			"created_at":     baseTime.AddDate(0, -1, -5),
		},
		{
			"id":             3,
			"lease_id":       3,
			"amount":         1800.00,
			"payment_date":   baseTime.AddDate(0, 0, -1),
			"payment_method": "Credit Card",
			"status":         "pending",
			"created_at":     baseTime.AddDate(0, 0, -1),
		},
	}
}

// GetRoleFixtures returns test roles
func GetRoleFixtures() []models.Role {
	baseTime := time.Now()

	return []models.Role{
		{
			ID:          1,
			Name:        "admin",
			DisplayName: "Administrator",
			Description: models.NullString("Full system access with all permissions"),
			Permissions: GetRolePermissions("admin"),
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, -6, 0),
		},
		{
			ID:          2,
			Name:        "property_manager",
			DisplayName: "Property Manager",
			Description: models.NullString("Manage properties, tenants, and maintenance"),
			Permissions: GetRolePermissions("property_manager"),
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, -6, 0),
		},
		{
			ID:          3,
			Name:        "tenant",
			DisplayName: "Tenant",
			Description: models.NullString("View own information and submit maintenance requests"),
			Permissions: GetRolePermissions("tenant"),
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, -6, 0),
		},
		{
			ID:          4,
			Name:        "viewer",
			DisplayName: "Viewer",
			Description: models.NullString("Read-only access to basic information"),
			Permissions: GetRolePermissions("viewer"),
			CreatedAt:   baseTime.AddDate(0, -6, 0),
			UpdatedAt:   baseTime.AddDate(0, -6, 0),
		},
	}
}

// CreateTestDatabase creates and seeds a test database with fixtures
func CreateTestDatabase() (*sql.DB, error) {
	// This would create a temporary test database and populate it with fixtures
	// Implementation would depend on your testing strategy (in-memory DB, Docker, etc.)
	return nil, nil
}

// SeedTestData seeds the database with test fixtures
func SeedTestData(db *sql.DB) error {
	// This would populate the database with all the fixture data
	// Implementation would involve running INSERT statements for all fixtures
	return nil
}

// CleanTestData removes all test data from the database
func CleanTestData(db *sql.DB) error {
	// This would clean up all test data
	// Implementation would involve running DELETE/TRUNCATE statements
	return nil
}
