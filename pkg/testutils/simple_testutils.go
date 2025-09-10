package testutils

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
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

// MakeTestRequest creates a test HTTP request
func MakeTestRequest(t *testing.T, method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	return req
}

// ExecuteRequest executes a test request and returns the response
func ExecuteRequest(req *http.Request, router http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
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

// MockRoleRows creates mock SQL rows for role queries
func MockRoleRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "display_name", "description", "permissions", "created_at", "updated_at",
	})
}

// TestTime returns a consistent time for testing
func TestTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}
