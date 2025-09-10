package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"github.com/greenbrown932/fire-pmaas/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		permission     string
		userRoles      []models.Role
		expectedStatus int
	}{
		{
			name:       "user has required permission",
			permission: "users.read",
			userRoles: []models.Role{
				{
					Name:        "admin",
					Permissions: models.StringArray{"users.read", "users.create"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "user has wildcard permission",
			permission: "users.delete",
			userRoles: []models.Role{
				{
					Name:        "admin",
					Permissions: models.StringArray{"users.*"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "user lacks required permission",
			permission: "properties.delete",
			userRoles: []models.Role{
				{
					Name:        "viewer",
					Permissions: models.StringArray{"properties.read"},
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no user in context",
			permission:     "users.read",
			userRoles:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that will be called if middleware passes
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Wrap with permission middleware
			handler := RequirePermission(tt.permission)(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Add user to context if userRoles is provided
			if tt.userRoles != nil {
				user := &models.User{
					ID:    1,
					Roles: tt.userRoles,
				}
				ctx := context.WithValue(req.Context(), UserContextKey, user)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		requiredRole   string
		userRoles      []models.Role
		expectedStatus int
	}{
		{
			name:         "user has required role",
			requiredRole: "admin",
			userRoles: []models.Role{
				{Name: "admin"},
				{Name: "user"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user lacks required role",
			requiredRole: "admin",
			userRoles: []models.Role{
				{Name: "user"},
				{Name: "viewer"},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no user in context",
			requiredRole:   "admin",
			userRoles:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			handler := RequireRole(tt.requiredRole)(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			if tt.userRoles != nil {
				user := &models.User{
					ID:    1,
					Roles: tt.userRoles,
				}
				ctx := context.WithValue(req.Context(), UserContextKey, user)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	tests := []struct {
		name           string
		requiredRoles  []string
		userRoles      []models.Role
		expectedStatus int
	}{
		{
			name:          "user has one of required roles",
			requiredRoles: []string{"admin", "property_manager"},
			userRoles: []models.Role{
				{Name: "property_manager"},
				{Name: "user"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user has multiple required roles",
			requiredRoles: []string{"admin", "property_manager"},
			userRoles: []models.Role{
				{Name: "admin"},
				{Name: "property_manager"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user lacks all required roles",
			requiredRoles: []string{"admin", "property_manager"},
			userRoles: []models.Role{
				{Name: "tenant"},
				{Name: "viewer"},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no user in context",
			requiredRoles:  []string{"admin"},
			userRoles:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			handler := RequireAnyRole(tt.requiredRoles...)(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			if tt.userRoles != nil {
				user := &models.User{
					ID:    1,
					Roles: tt.userRoles,
				}
				ctx := context.WithValue(req.Context(), UserContextKey, user)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test with user in context
	ctx := context.WithValue(context.Background(), UserContextKey, user)
	retrievedUser, ok := GetUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, user, retrievedUser)

	// Test without user in context
	emptyCtx := context.Background()
	retrievedUser, ok = GetUserFromContext(emptyCtx)
	assert.False(t, ok)
	assert.Nil(t, retrievedUser)
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Token should be URL-safe base64 encoded
	assert.Regexp(t, `^[A-Za-z0-9_-]+$`, token1)
}

func TestCreateUserSession(t *testing.T) {
	testutils.SetTestEnv(t)
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userID := 1
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-browser")
	req.RemoteAddr = "192.168.1.1:12345"

	// Mock session creation
	testDB.Mock.ExpectQuery(`INSERT INTO user_sessions`).
		WillReturnRows(testDB.Mock.NewRows([]string{"id", "created_at"}).
			AddRow(1, time.Now()))

	session, err := CreateUserSession(userID, req)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, userID, session.UserID)
	assert.NotEmpty(t, session.SessionToken)
	assert.True(t, session.IPAddress.Valid)
	assert.True(t, session.UserAgent.Valid)

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestSessionAuth(t *testing.T) {
	testutils.SetTestEnv(t)
	testDB := testutils.SetupTestDB(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name          string
		sessionToken  string
		setupMock     func()
		expectUserSet bool
	}{
		{
			name:         "valid session token",
			sessionToken: "valid_token_123",
			setupMock: func() {
				// Mock session lookup
				sessionRows := testDB.Mock.NewRows([]string{
					"id", "user_id", "session_token", "ip_address", "user_agent", "expires_at", "created_at",
				}).AddRow(1, 1, "valid_token_123", "127.0.0.1", "test-agent", time.Now().Add(time.Hour), time.Now())

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM user_sessions WHERE session_token = \$1`).
					WithArgs("valid_token_123").
					WillReturnRows(sessionRows)

				// Mock user lookup
				userRows := testutils.MockUserRows().AddRow(
					1, nil, "testuser", "test@example.com", "Test", "User",
					nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
				)

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(userRows)

				// Mock roles query
				rolesRows := testDB.Mock.NewRows([]string{
					"id", "name", "display_name", "description", "permissions", "created_at", "updated_at",
				})

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
					WithArgs(1).
					WillReturnRows(rolesRows)
			},
			expectUserSet: true,
		},
		{
			name:         "invalid session token",
			sessionToken: "invalid_token",
			setupMock: func() {
				testDB.Mock.ExpectQuery(`SELECT (.+) FROM user_sessions WHERE session_token = \$1`).
					WithArgs("invalid_token").
					WillReturnError(sql.ErrNoRows)
			},
			expectUserSet: false,
		},
		{
			name:          "no session token",
			sessionToken:  "",
			setupMock:     func() {},
			expectUserSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, ok := GetUserFromContext(r.Context())
				if tt.expectUserSet {
					assert.True(t, ok)
					assert.NotNil(t, user)
				} else {
					assert.False(t, ok)
					assert.Nil(t, user)
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := SessionAuth(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.sessionToken})
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			if len(testDB.Mock.ExpectedQuery) > 0 {
				require.NoError(t, testDB.Mock.ExpectationsWereMet())
			}
		})
	}
}

func TestLoadUserFromToken(t *testing.T) {
	// This test would require setting up a proper OIDC provider mock
	// For now, we'll test the basic structure without actual OIDC verification

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler would check if user was loaded from token
		w.WriteHeader(http.StatusOK)
	})

	handler := LoadUserFromToken(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	// Add a mock ID token cookie (in real tests, this would be a valid JWT)
	req.AddCookie(&http.Cookie{Name: "id_token", Value: "mock_jwt_token"})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Without proper OIDC setup, this should just pass through
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For header present",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name: "X-Real-IP header present",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.195",
			},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "no special headers",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.100:12345",
			expected:   "192.168.1.100:12345",
		},
		{
			name: "X-Forwarded-For takes precedence over X-Real-IP",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "192.168.1.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getClientIP(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}
