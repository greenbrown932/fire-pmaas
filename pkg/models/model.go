package models

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"strings"
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/lib/pq"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// Property represents a property in our system.
type Property struct {
	ID           int // Unique identifier for the property
	Name         string
	Address      string
	PropertyType string
	CreatedAt    time.Time
	UpdatedAt    time.Time // Time the property was last updated
}

// PropertyUnit represents an individual unit within a property.
type PropertyUnit struct {
	ID          int // Unique identifier for the property unit
	PropertyID  int
	UnitNumber  string
	Bedrooms    int
	Bathrooms   int
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the property unit was last updated
}

// Tenant stores information about individual tenants.
type Tenant struct {
	ID          int // Unique identifier for the tenant
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the tenant was last updated
}

// Lease connects a tenant to a specific unit for a period of time.
type Lease struct {
	ID          int // Unique identifier for the lease
	UnitID      int
	TenantID    int
	StartDate   time.Time
	EndDate     time.Time
	MonthlyRent float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the lease was last updated
}

// PropertyDetail represents a detailed view of a property, including lease and tenant information.
type PropertyDetail struct {
	ID         int    // Unique identifier for the property
	Address    string // Address of the property
	Rent       float64
	Status     string
	Bedrooms   int
	Bathrooms  int
	TenantName string // Name of the tenant currently leasing the property
}

// User represents a user in the system with authentication and profile information
type User struct {
	ID                   int            `json:"id"`
	KeycloakID           sql.NullString `json:"keycloak_id,omitempty"`
	Username             string         `json:"username"`
	Email                string         `json:"email"`
	FirstName            string         `json:"first_name"`
	LastName             string         `json:"last_name"`
	PhoneNumber          sql.NullString `json:"phone_number,omitempty"`
	ProfilePictureURL    sql.NullString `json:"profile_picture_url,omitempty"`
	EmailVerified        bool           `json:"email_verified"`
	MFAEnabled           bool           `json:"mfa_enabled"`
	MFASecret            sql.NullString `json:"-"` // Never expose in JSON
	Status               string         `json:"status"`
	LastLogin            sql.NullTime   `json:"last_login,omitempty"`
	PasswordResetToken   sql.NullString `json:"-"` // Never expose in JSON
	PasswordResetExpires sql.NullTime   `json:"-"` // Never expose in JSON
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	Roles                []Role         `json:"roles,omitempty"`
}

// Role represents a role in the system with associated permissions
type Role struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description sql.NullString `json:"description,omitempty"`
	Permissions StringArray    `json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// UserRole represents the junction table for user-role relationships
type UserRole struct {
	ID         int           `json:"id"`
	UserID     int           `json:"user_id"`
	RoleID     int           `json:"role_id"`
	AssignedAt time.Time     `json:"assigned_at"`
	AssignedBy sql.NullInt32 `json:"assigned_by,omitempty"`
}

// UserSession represents an active user session
type UserSession struct {
	ID           int            `json:"id"`
	UserID       int            `json:"user_id"`
	SessionToken string         `json:"-"` // Never expose in JSON
	IPAddress    sql.NullString `json:"ip_address,omitempty"`
	UserAgent    sql.NullString `json:"user_agent,omitempty"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
}

// StringArray is a custom type for handling PostgreSQL array columns
type StringArray []string

// Value implements the driver.Valuer interface for database storage
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return pq.Array(a).Value()
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	return pq.Array(a).Scan(value)
}

// UserRegistration represents the data needed for user registration
type UserRegistration struct {
	Username        string `json:"username" validate:"required,min=3,max=50"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	FirstName       string `json:"first_name" validate:"required,min=1,max=100"`
	LastName        string `json:"last_name" validate:"required,min=1,max=100"`
	PhoneNumber     string `json:"phone_number,omitempty"`
}

// UserLogin represents the data needed for user login
type UserLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// PasswordReset represents the data needed for password reset
type PasswordReset struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// CreateProperty creates a new property in the database.
func CreateProperty(property *Property) error {
	_, err := db.DB.Exec(`
		INSERT INTO properties (name, address, property_type)
		VALUES ($1, $2, $3)
	`, property.Name, property.Address, property.PropertyType)
	return err
}

// UpdateProperty updates an existing property in the database.
func UpdateProperty(property *Property) error {
	_, err := db.DB.Exec(`
		UPDATE properties
		SET name = $1, address = $2, property_type = $3
		WHERE id = $4
	`, property.Name, property.Address, property.PropertyType, property.ID)
	return err
}

// DeleteProperty deletes a property from the database.
func DeleteProperty(id int) error {
	_, err := db.DB.Exec(`
		DELETE FROM properties
		WHERE id = $1
	`, id)
	return err
}

// GetProperties retrieves a list of properties with details including address, rent, status, and tenant name.
func GetProperties() ([]PropertyDetail, error) {
	// Execute the SQL query to retrieve property details.
	rows, err := db.DB.Query(`
		SELECT
			p.id,                             -- Property ID
			p.address,                        -- Property Address
			COALESCE(l.monthly_rent, 0),      -- Lease Monthly Rent (default to 0 if NULL)
			COALESCE(l.status, 'vacant'),     -- Lease Status (default to 'vacant' if NULL)
			COALESCE(pu.bedrooms, 0),         -- Property Unit Bedrooms (default to 0 if NULL)
			COALESCE(pu.bathrooms, 0),        -- Property Unit Bathrooms (default to 0 if NULL)
			COALESCE(t.first_name || ' ' || t.last_name, '') AS tenant_name -- Tenant Full Name (empty if NULL)
		FROM properties p                                     -- From the properties table
		LEFT JOIN property_units pu ON p.id = pu.property_id   -- Join with property_units table
		LEFT JOIN leases l ON pu.id = l.unit_id AND l.status = 'active' -- Only active leases
		LEFT JOIN tenants t ON l.tenant_id = t.id             -- Join with tenants table
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []PropertyDetail
	for rows.Next() {
		var p PropertyDetail
		if err := rows.Scan(&p.ID, &p.Address, &p.Rent, &p.Status, &p.Bedrooms, &p.Bathrooms, &p.TenantName); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, nil
}

// User management functions

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateResetToken generates a secure random token for password reset
func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateMFASecret generates a new TOTP secret for MFA
func GenerateMFASecret() (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Fire-PMAAS",
		AccountName: "user@example.com", // This should be replaced with actual user email
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// ValidateMFACode validates a TOTP code against the user's secret
func ValidateMFACode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// CreateUser creates a new user in the database
func CreateUser(user *User) error {
	query := `
		INSERT INTO users (keycloak_id, username, email, first_name, last_name, phone_number,
						   profile_picture_url, email_verified, mfa_enabled, mfa_secret, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	return db.DB.QueryRow(query, user.KeycloakID, user.Username, user.Email, user.FirstName,
		user.LastName, user.PhoneNumber, user.ProfilePictureURL, user.EmailVerified,
		user.MFAEnabled, user.MFASecret, user.Status).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT id, keycloak_id, username, email, first_name, last_name, phone_number,
			   profile_picture_url, email_verified, mfa_enabled, mfa_secret, status,
			   last_login, password_reset_token, password_reset_expires, created_at, updated_at
		FROM users WHERE id = $1`

	err := db.DB.QueryRow(query, id).Scan(&user.ID, &user.KeycloakID, &user.Username,
		&user.Email, &user.FirstName, &user.LastName, &user.PhoneNumber,
		&user.ProfilePictureURL, &user.EmailVerified, &user.MFAEnabled, &user.MFASecret,
		&user.Status, &user.LastLogin, &user.PasswordResetToken, &user.PasswordResetExpires,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Load user roles
	user.Roles, err = GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, keycloak_id, username, email, first_name, last_name, phone_number,
			   profile_picture_url, email_verified, mfa_enabled, mfa_secret, status,
			   last_login, password_reset_token, password_reset_expires, created_at, updated_at
		FROM users WHERE email = $1`

	err := db.DB.QueryRow(query, email).Scan(&user.ID, &user.KeycloakID, &user.Username,
		&user.Email, &user.FirstName, &user.LastName, &user.PhoneNumber,
		&user.ProfilePictureURL, &user.EmailVerified, &user.MFAEnabled, &user.MFASecret,
		&user.Status, &user.LastLogin, &user.PasswordResetToken, &user.PasswordResetExpires,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Load user roles
	user.Roles, err = GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, keycloak_id, username, email, first_name, last_name, phone_number,
			   profile_picture_url, email_verified, mfa_enabled, mfa_secret, status,
			   last_login, password_reset_token, password_reset_expires, created_at, updated_at
		FROM users WHERE username = $1`

	err := db.DB.QueryRow(query, username).Scan(&user.ID, &user.KeycloakID, &user.Username,
		&user.Email, &user.FirstName, &user.LastName, &user.PhoneNumber,
		&user.ProfilePictureURL, &user.EmailVerified, &user.MFAEnabled, &user.MFASecret,
		&user.Status, &user.LastLogin, &user.PasswordResetToken, &user.PasswordResetExpires,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Load user roles
	user.Roles, err = GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information in the database
func UpdateUser(user *User) error {
	query := `
		UPDATE users SET username = $1, email = $2, first_name = $3, last_name = $4,
					     phone_number = $5, profile_picture_url = $6, email_verified = $7,
					     mfa_enabled = $8, mfa_secret = $9, status = $10, last_login = $11,
					     password_reset_token = $12, password_reset_expires = $13, updated_at = NOW()
		WHERE id = $14`

	_, err := db.DB.Exec(query, user.Username, user.Email, user.FirstName, user.LastName,
		user.PhoneNumber, user.ProfilePictureURL, user.EmailVerified, user.MFAEnabled,
		user.MFASecret, user.Status, user.LastLogin, user.PasswordResetToken,
		user.PasswordResetExpires, user.ID)

	return err
}

// DeleteUser deletes a user from the database
func DeleteUser(id int) error {
	_, err := db.DB.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// GetUserRoles retrieves all roles for a specific user
func GetUserRoles(userID int) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.permissions, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Permissions, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm == permission {
				return true
			}
			// Check for wildcard permissions
			if strings.HasSuffix(perm, ".*") {
				prefix := strings.TrimSuffix(perm, ".*")
				if strings.HasPrefix(permission, prefix) {
					return true
				}
			}
		}
	}
	return false
}

// HasRole checks if a user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// AssignRole assigns a role to a user
func AssignRole(userID, roleID int, assignedBy *int) error {
	query := `INSERT INTO user_roles (user_id, role_id, assigned_by) VALUES ($1, $2, $3)`
	_, err := db.DB.Exec(query, userID, roleID, assignedBy)
	return err
}

// RemoveRole removes a role from a user
func RemoveRole(userID, roleID int) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := db.DB.Exec(query, userID, roleID)
	return err
}

// GetAllRoles retrieves all available roles
func GetAllRoles() ([]Role, error) {
	query := `SELECT id, name, display_name, description, permissions, created_at, updated_at FROM roles ORDER BY name`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Permissions, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoleByName retrieves a role by its name
func GetRoleByName(name string) (*Role, error) {
	role := &Role{}
	query := `SELECT id, name, display_name, description, permissions, created_at, updated_at FROM roles WHERE name = $1`

	err := db.DB.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.DisplayName,
		&role.Description, &role.Permissions, &role.CreatedAt, &role.UpdatedAt)

	return role, err
}

// CreateUserSession creates a new user session
func CreateUserSession(session *UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return db.DB.QueryRow(query, session.UserID, session.SessionToken, session.IPAddress,
		session.UserAgent, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
}

// GetUserSession retrieves a user session by token
func GetUserSession(token string) (*UserSession, error) {
	session := &UserSession{}
	query := `
		SELECT id, user_id, session_token, ip_address, user_agent, expires_at, created_at
		FROM user_sessions
		WHERE session_token = $1 AND expires_at > NOW()`

	err := db.DB.QueryRow(query, token).Scan(&session.ID, &session.UserID, &session.SessionToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt)

	return session, err
}

// DeleteUserSession deletes a user session
func DeleteUserSession(token string) error {
	_, err := db.DB.Exec("DELETE FROM user_sessions WHERE session_token = $1", token)
	return err
}

// CleanupExpiredSessions removes expired user sessions
func CleanupExpiredSessions() error {
	_, err := db.DB.Exec("DELETE FROM user_sessions WHERE expires_at <= NOW()")
	return err
}

// GetUserByKeycloakID retrieves a user by their Keycloak ID
func GetUserByKeycloakID(keycloakID string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, keycloak_id, username, email, first_name, last_name, phone_number,
			   profile_picture_url, email_verified, mfa_enabled, mfa_secret, status,
			   last_login, password_reset_token, password_reset_expires, created_at, updated_at
		FROM users WHERE keycloak_id = $1`

	err := db.DB.QueryRow(query, keycloakID).Scan(&user.ID, &user.KeycloakID, &user.Username,
		&user.Email, &user.FirstName, &user.LastName, &user.PhoneNumber,
		&user.ProfilePictureURL, &user.EmailVerified, &user.MFAEnabled, &user.MFASecret,
		&user.Status, &user.LastLogin, &user.PasswordResetToken, &user.PasswordResetExpires,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Load user roles
	user.Roles, err = GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// NullString is a helper function to create sql.NullString
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
