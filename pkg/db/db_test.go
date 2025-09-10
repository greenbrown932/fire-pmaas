package db

import (
	"database/sql"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/greenbrown932/fire-pmaas/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB(t *testing.T) {
	// This test would require a real database connection or more complex mocking
	// For now, we'll test the environment variable validation logic

	testutils.SetTestEnv(t)

	// Test that environment variables are set correctly
	testEnvVars := map[string]string{
		"POSTGRES_HOST":     "localhost",
		"POSTGRES_PORT":     "5432",
		"POSTGRES_USER":     "test_user",
		"POSTGRES_PASSWORD": "test_pass",
		"POSTGRES_DB":       "test_db",
	}

	for key, expectedValue := range testEnvVars {
		actualValue := os.Getenv(key)
		assert.Equal(t, expectedValue, actualValue, "Environment variable %s should be set correctly", key)
	}
}

func TestSeedDatabase(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Store original DB and replace with mock
	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Expected properties to be seeded
	expectedProperties := []struct {
		Name         string
		Address      string
		PropertyType string
	}{
		{"Sunset Apartments", "123 Main St, Anytown", "Apartment Building"},
		{"Oakwood Villas", "456 Elm St, Anytown", "Townhouse Complex"},
		{"Pine Ridge Cottage", "789 Oak St, Anytown", "Single Family Home"},
	}

	// Mock the INSERT statements for each property
	for _, prop := range expectedProperties {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO properties (name, address, property_type) VALUES ($1, $2, $3)`)).
			WithArgs(prop.Name, prop.Address, prop.PropertyType).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Run the seed function
	SeedDatabase()

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseConnectionString(t *testing.T) {
	testutils.SetTestEnv(t)

	expectedURL := "postgres://test_user:test_pass@localhost:5432/test_db?sslmode=disable"
	actualURL := testutils.TestDBURL()

	assert.Equal(t, expectedURL, actualURL)
}

func TestDatabaseTransaction(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Test a transaction scenario
	mock.ExpectBegin()

	// Mock a property insertion within transaction
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO properties`)).
		WithArgs("Test Property", "123 Test St", "Test Type").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock a property unit insertion within same transaction
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO property_units`)).
		WithArgs(1, "101", 2, 1, "Test unit").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// Simulate transaction
	tx, err := mockDB.Begin()
	require.NoError(t, err)

	// Insert property
	_, err = tx.Exec(`INSERT INTO properties (name, address, property_type) VALUES ($1, $2, $3)`,
		"Test Property", "123 Test St", "Test Type")
	assert.NoError(t, err)

	// Insert property unit
	_, err = tx.Exec(`INSERT INTO property_units (property_id, unit_number, bedrooms, bathrooms, description) VALUES ($1, $2, $3, $4, $5)`,
		1, "101", 2, 1, "Test unit")
	assert.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	assert.NoError(t, err)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseRollback(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Test a transaction rollback scenario
	mock.ExpectBegin()

	// Mock a successful insertion
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO properties`)).
		WithArgs("Test Property", "123 Test St", "Test Type").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock a failed insertion
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO property_units`)).
		WithArgs(1, "101", 2, 1, "Test unit").
		WillReturnError(sql.ErrConnDone)

	mock.ExpectRollback()

	// Simulate transaction with rollback
	tx, err := mockDB.Begin()
	require.NoError(t, err)

	// Insert property (successful)
	_, err = tx.Exec(`INSERT INTO properties (name, address, property_type) VALUES ($1, $2, $3)`,
		"Test Property", "123 Test St", "Test Type")
	assert.NoError(t, err)

	// Insert property unit (fails)
	_, err = tx.Exec(`INSERT INTO property_units (property_id, unit_number, bedrooms, bathrooms, description) VALUES ($1, $2, $3, $4, $5)`,
		1, "101", 2, 1, "Test unit")
	assert.Error(t, err)

	// Rollback transaction
	err = tx.Rollback()
	assert.NoError(t, err)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseQuery(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Mock a SELECT query
	rows := sqlmock.NewRows([]string{"id", "name", "address", "property_type"}).
		AddRow(1, "Test Property", "123 Test St", "Apartment").
		AddRow(2, "Another Property", "456 Oak Ave", "House")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, address, property_type FROM properties WHERE property_type = $1`)).
		WithArgs("Apartment").
		WillReturnRows(rows)

	// Execute query
	queryRows, err := mockDB.Query(`SELECT id, name, address, property_type FROM properties WHERE property_type = $1`, "Apartment")
	require.NoError(t, err)
	defer queryRows.Close()

	// Process results
	var properties []map[string]interface{}
	for queryRows.Next() {
		var id int
		var name, address, propertyType string
		err := queryRows.Scan(&id, &name, &address, &propertyType)
		require.NoError(t, err)

		properties = append(properties, map[string]interface{}{
			"id":            id,
			"name":          name,
			"address":       address,
			"property_type": propertyType,
		})
	}

	// Verify results
	assert.Len(t, properties, 2)
	assert.Equal(t, "Test Property", properties[0]["name"])
	assert.Equal(t, "Another Property", properties[1]["name"])

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseUpdate(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Mock an UPDATE query
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE properties SET name = $1, address = $2 WHERE id = $3`)).
		WithArgs("Updated Property", "999 New St", 1).
		WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected

	// Execute update
	result, err := mockDB.Exec(`UPDATE properties SET name = $1, address = $2 WHERE id = $3`,
		"Updated Property", "999 New St", 1)
	require.NoError(t, err)

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseDelete(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Mock a DELETE query
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM properties WHERE id = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected

	// Execute delete
	result, err := mockDB.Exec(`DELETE FROM properties WHERE id = $1`, 1)
	require.NoError(t, err)

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseConstraintViolation(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Mock a constraint violation (e.g., duplicate email)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (username, email) VALUES ($1, $2)`)).
		WithArgs("testuser", "duplicate@test.com").
		WillReturnError(sql.ErrConnDone) // Simulate constraint error

	// Execute insert that violates constraint
	_, err = mockDB.Exec(`INSERT INTO users (username, email) VALUES ($1, $2)`,
		"testuser", "duplicate@test.com")
	assert.Error(t, err)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseConnectionPool(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Test connection pool settings
	mockDB.SetMaxOpenConns(25)
	mockDB.SetMaxIdleConns(5)

	// Mock a simple query to test connection
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1`)).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(1))

	// Execute query
	var result int
	err = mockDB.QueryRow(`SELECT 1`).Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabasePreparedStatement(t *testing.T) {
	// Setup mock database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Mock prepared statement
	mock.ExpectPrepare(regexp.QuoteMeta(`SELECT name FROM properties WHERE id = $1`))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT name FROM properties WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Test Property"))

	// Prepare statement
	stmt, err := mockDB.Prepare(`SELECT name FROM properties WHERE id = $1`)
	require.NoError(t, err)
	defer stmt.Close()

	// Execute prepared statement
	var name string
	err = stmt.QueryRow(1).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Test Property", name)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}
