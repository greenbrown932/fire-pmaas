package testutils

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestDB represents a test database connection
type TestDB struct {
	DB   *sql.DB
	Mock sqlmock.Sqlmock
}

// SetupTestDB creates a mock database for testing
func SetupTestDB(t *testing.T) *TestDB {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return &TestDB{
		DB:   db,
		Mock: mock,
	}
}

// TeardownTestDB closes the test database
func (tdb *TestDB) TeardownTestDB() {
	tdb.DB.Close()
}

// CreateTestUser creates a test user with given parameters
func CreateTestUser(t *testing.T, username, email, role string) *models.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &models.User{
		ID:            1,
		Username:      username,
		Email:         email,
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Add role if specified
	if role != "" {
		user.Roles = []models.Role{
			{
				ID:          1,
				Name:        role,
				DisplayName: role,
				Permissions: GetRolePermissions(role),
			},
		}
	}

	return user
}

// GetRolePermissions returns permissions for a given role
func GetRolePermissions(role string) models.StringArray {
	switch role {
	case "admin":
		return models.StringArray{
			"users.create", "users.read", "users.update", "users.delete",
			"properties.create", "properties.read", "properties.update", "properties.delete",
			"tenants.create", "tenants.read", "tenants.update", "tenants.delete",
			"leases.create", "leases.read", "leases.update", "leases.delete",
			"payments.create", "payments.read", "payments.update", "payments.delete",
			"maintenance.create", "maintenance.read", "maintenance.update", "maintenance.delete",
			"roles.manage", "system.settings",
		}
	case "property_manager":
		return models.StringArray{
			"properties.create", "properties.read", "properties.update", "properties.delete",
			"tenants.create", "tenants.read", "tenants.update", "tenants.delete",
			"leases.create", "leases.read", "leases.update", "leases.delete",
			"payments.read", "payments.update",
			"maintenance.create", "maintenance.read", "maintenance.update", "maintenance.delete",
		}
	case "tenant":
		return models.StringArray{
			"profile.read", "profile.update",
			"lease.read.own", "payments.read.own",
			"maintenance.create.own", "maintenance.read.own",
		}
	case "viewer":
		return models.StringArray{
			"properties.read", "tenants.read", "maintenance.read",
		}
	default:
		return models.StringArray{}
	}
}

// CreateTestProperty creates a test property
func CreateTestProperty() *models.Property {
	return &models.Property{
		ID:           1,
		Name:         "Test Property",
		Address:      "123 Test St, Test City, TS 12345",
		PropertyType: "Apartment Building",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CreateTestPropertyUnit creates a test property unit
func CreateTestPropertyUnit(propertyID int) *models.PropertyUnit {
	return &models.PropertyUnit{
		ID:          1,
		PropertyID:  propertyID,
		UnitNumber:  "101",
		Bedrooms:    2,
		Bathrooms:   1,
		Description: "Test unit description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestTenant creates a test tenant
func CreateTestTenant() *models.Tenant {
	return &models.Tenant{
		ID:          1,
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@test.com",
		PhoneNumber: "555-0123",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestLease creates a test lease
func CreateTestLease(unitID, tenantID int) *models.Lease {
	return &models.Lease{
		ID:          1,
		UnitID:      unitID,
		TenantID:    tenantID,
		StartDate:   time.Now().AddDate(0, -1, 0), // Started 1 month ago
		EndDate:     time.Now().AddDate(1, 0, 0),  // Ends in 1 year
		MonthlyRent: 1500.00,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// SetupTestRouter creates a test router with basic middleware
func SetupTestRouter() *chi.Mux {
	r := chi.NewRouter()
	return r
}

// MakeTestRequest creates a test HTTP request
func MakeTestRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	require.NoError(t, err)
	return req
}

// ExecuteRequest executes a test request and returns the response
func ExecuteRequest(req *http.Request, router http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// SetTestEnv sets up test environment variables
func SetTestEnv(t *testing.T) {
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "test_user")
	os.Setenv("POSTGRES_PASSWORD", "test_pass")
	os.Setenv("POSTGRES_DB", "test_db")
	os.Setenv("KEYCLOAK_ISSUER", "http://localhost:8080/realms/test")

	t.Cleanup(func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DB")
		os.Unsetenv("KEYCLOAK_ISSUER")
	})
}

// AssertJSONResponse checks if response contains expected JSON
func AssertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) {
	require.Equal(t, expectedStatus, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

// CreateTestSession creates a test user session
func CreateTestSession(userID int) *models.UserSession {
	return &models.UserSession{
		ID:           1,
		UserID:       userID,
		SessionToken: "test_session_token_123",
		IPAddress:    models.NullString("127.0.0.1"),
		UserAgent:    models.NullString("test-agent"),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}
}

// MockUserRows creates mock SQL rows for user queries
func MockUserRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "keycloak_id", "username", "email", "first_name", "last_name",
		"phone_number", "profile_picture_url", "email_verified", "mfa_enabled",
		"mfa_secret", "status", "last_login", "password_reset_token",
		"password_reset_expires", "created_at", "updated_at",
	})
}

// MockPropertyRows creates mock SQL rows for property queries
func MockPropertyRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "address", "property_type", "created_at", "updated_at",
	})
}

// MockPropertyDetailRows creates mock SQL rows for property detail queries
func MockPropertyDetailRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "address", "monthly_rent", "status", "bedrooms", "bathrooms", "tenant_name",
	})
}

// TestDBURL returns a test database URL
func TestDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"test_user", "test_pass", "localhost", "5432", "test_db")
}
