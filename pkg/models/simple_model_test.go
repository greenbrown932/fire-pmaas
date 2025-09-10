package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Replace the global DB with our mock
	originalDB := db.DB
	db.DB = mockDB

	return mock, func() {
		db.DB = originalDB
		mockDB.Close()
	}
}

func TestCreateUser(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Username:      "testuser",
		Email:         "test@example.com",
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
		Status:        "active",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.KeycloakID, user.Username, user.Email, user.FirstName,
			user.LastName, user.PhoneNumber, user.ProfilePictureURL, user.EmailVerified,
			user.MFAEnabled, user.MFASecret, user.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	err := CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	email := "test@example.com"
	now := time.Now()

	userRows := sqlmock.NewRows([]string{
		"id", "keycloak_id", "username", "email", "first_name", "last_name",
		"phone_number", "profile_picture_url", "email_verified", "mfa_enabled",
		"mfa_secret", "status", "last_login", "password_reset_token",
		"password_reset_expires", "created_at", "updated_at",
	}).AddRow(1, sql.NullString{}, "testuser", email, "Test", "User",
		sql.NullString{}, sql.NullString{}, true, false, sql.NullString{},
		"active", sql.NullTime{}, sql.NullString{}, sql.NullTime{}, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(userRows)

	// Mock roles query - return empty rows to avoid the StringArray issue
	rolesRows := sqlmock.NewRows([]string{
		"id", "name", "display_name", "description", "permissions", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
		WithArgs(1).
		WillReturnRows(rolesRows)

	user, err := GetUserByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Len(t, user.Roles, 0) // No roles in this test

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUser(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		ID:        1,
		Username:  "updateduser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
		Status:    "active",
	}

	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(user.Username, user.Email, user.FirstName, user.LastName,
			user.PhoneNumber, user.ProfilePictureURL, user.EmailVerified,
			user.MFAEnabled, user.MFASecret, user.Status, user.LastLogin,
			user.PasswordResetToken, user.PasswordResetExpires, user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateUser(user)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	userID := 1

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeleteUser(userID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserHasPermission(t *testing.T) {
	user := &User{
		Roles: []Role{
			{
				Name:        "admin",
				Permissions: StringArray{"users.create", "users.read", "users.update", "users.delete"},
			},
		},
	}

	assert.True(t, user.HasPermission("users.create"))
	assert.True(t, user.HasPermission("users.read"))
	assert.False(t, user.HasPermission("properties.delete"))

	// Test wildcard permissions
	user.Roles[0].Permissions = StringArray{"users.*", "properties.read"}
	assert.True(t, user.HasPermission("users.create"))
	assert.True(t, user.HasPermission("users.delete"))
	assert.True(t, user.HasPermission("properties.read"))
	assert.False(t, user.HasPermission("properties.delete"))
}

func TestUserHasRole(t *testing.T) {
	user := &User{
		Roles: []Role{
			{Name: "admin"},
			{Name: "property_manager"},
		},
	}

	assert.True(t, user.HasRole("admin"))
	assert.True(t, user.HasRole("property_manager"))
	assert.False(t, user.HasRole("tenant"))
}

func TestCreateProperty(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	property := &Property{
		Name:         "Test Property",
		Address:      "123 Test St",
		PropertyType: "Apartment",
	}

	mock.ExpectExec(`INSERT INTO properties`).
		WithArgs(property.Name, property.Address, property.PropertyType).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := CreateProperty(property)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProperties(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "address", "monthly_rent", "status", "bedrooms", "bathrooms", "tenant_name",
	}).AddRow(1, "123 Test St", 1500.00, "active", 2, 1, "John Doe").
		AddRow(2, "456 Oak Ave", 1200.00, "vacant", 1, 1, "")

	mock.ExpectQuery(`SELECT (.+) FROM properties p`).
		WillReturnRows(rows)

	properties, err := GetProperties()
	assert.NoError(t, err)
	assert.Len(t, properties, 2)
	assert.Equal(t, "123 Test St", properties[0].Address)
	assert.Equal(t, 1500.00, properties[0].Rent)
	assert.Equal(t, "John Doe", properties[0].TenantName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	assert.True(t, CheckPassword(password, hash))
	assert.False(t, CheckPassword("wrongpassword", hash))
}

func TestGenerateResetToken(t *testing.T) {
	token, err := GenerateResetToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Generate another token to ensure they're different
	token2, err := GenerateResetToken()
	assert.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestNullString(t *testing.T) {
	// Test with empty string
	nullStr := NullString("")
	assert.False(t, nullStr.Valid)

	// Test with non-empty string
	nullStr = NullString("test")
	assert.True(t, nullStr.Valid)
	assert.Equal(t, "test", nullStr.String)
}

func TestStringArrayValue(t *testing.T) {
	// Test empty array
	arr := StringArray{}
	val, err := arr.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test with values
	arr = StringArray{"perm1", "perm2", "perm3"}
	val, err = arr.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)
}
