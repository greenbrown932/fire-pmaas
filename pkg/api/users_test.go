package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"github.com/greenbrown932/fire-pmaas/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserAPITest(t *testing.T) (*chi.Mux, *testutils.TestDB) {
	testutils.SetTestEnv(t)
	testDB := testutils.SetupTestDB(t)

	router := chi.NewRouter()
	RegisterUserRoutes(router)

	return router, testDB
}

func TestUserRegistration(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name           string
		payload        models.UserRegistration
		expectedStatus int
		setupMock      func()
	}{
		{
			name: "successful registration",
			payload: models.UserRegistration{
				Username:        "newuser",
				Email:           "newuser@test.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				FirstName:       "New",
				LastName:        "User",
				PhoneNumber:     "555-0123",
			},
			expectedStatus: http.StatusOK,
			setupMock: func() {
				// Mock user existence check (should return error = not found)
				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
					WithArgs("newuser@test.com").
					WillReturnError(sql.ErrNoRows)

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE username = \$1`).
					WithArgs("newuser").
					WillReturnError(sql.ErrNoRows)

				// Mock user creation
				testDB.Mock.ExpectQuery(`INSERT INTO users`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()))

				// Mock role assignment
				testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles WHERE name = \$1`).
					WithArgs("tenant").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"}).
						AddRow(1, "tenant", "Tenant", nil, testutils.GetRolePermissions("tenant"), time.Now(), time.Now()))

				testDB.Mock.ExpectExec(`INSERT INTO user_roles`).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "invalid email format",
			payload: models.UserRegistration{
				Username:        "newuser",
				Email:           "invalid-email",
				Password:        "password123",
				ConfirmPassword: "password123",
				FirstName:       "New",
				LastName:        "User",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      func() {},
		},
		{
			name: "password mismatch",
			payload: models.UserRegistration{
				Username:        "newuser",
				Email:           "newuser@test.com",
				Password:        "password123",
				ConfirmPassword: "differentpassword",
				FirstName:       "New",
				LastName:        "User",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      func() {},
		},
		{
			name: "user already exists",
			payload: models.UserRegistration{
				Username:        "existinguser",
				Email:           "existing@test.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				FirstName:       "Existing",
				LastName:        "User",
			},
			expectedStatus: http.StatusConflict,
			setupMock: func() {
				// Mock user existence check (user exists)
				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
					WithArgs("existing@test.com").
					WillReturnRows(testutils.MockUserRows().AddRow(
						1, nil, "existinguser", "existing@test.com", "Existing", "User",
						nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload.Username, response["username"])
				assert.Equal(t, tt.payload.Email, response["email"])
				assert.NotContains(t, response, "password") // Ensure password is not returned
			}

			require.NoError(t, testDB.Mock.ExpectationsWereMet())
		})
	}
}

func TestGetProfile(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	testUser := testutils.CreateTestUser(t, "testuser", "test@example.com", "tenant")

	req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)

	// Add user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testUser.Username, response.Username)
	assert.Equal(t, testUser.Email, response.Email)
}

func TestUpdateProfile(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	testUser := testutils.CreateTestUser(t, "testuser", "test@example.com", "tenant")

	updateData := map[string]interface{}{
		"first_name":   "Updated",
		"last_name":    "Name",
		"phone_number": "555-9999",
	}

	// Mock update user
	testDB.Mock.ExpectExec(`UPDATE users SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payloadBytes, err := json.Marshal(updateData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.User
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", response.FirstName)
	assert.Equal(t, "Name", response.LastName)

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestListUsers_AdminOnly(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name           string
		userRole       string
		expectedStatus int
		setupMock      func()
	}{
		{
			name:           "admin can list users",
			userRole:       "admin",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				rows := testutils.MockUserRows().
					AddRow(1, nil, "user1", "user1@test.com", "User", "One", nil, true, false, "active", nil, time.Now()).
					AddRow(2, nil, "user2", "user2@test.com", "User", "Two", nil, true, false, "active", nil, time.Now())

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users u ORDER BY`).
					WillReturnRows(rows)
			},
		},
		{
			name:           "property manager can list users",
			userRole:       "property_manager",
			expectedStatus: http.StatusOK,
			setupMock: func() {
				rows := testutils.MockUserRows().
					AddRow(1, nil, "user1", "user1@test.com", "User", "One", nil, true, false, "active", nil, time.Now())

				testDB.Mock.ExpectQuery(`SELECT (.+) FROM users u ORDER BY`).
					WillReturnRows(rows)
			},
		},
		{
			name:           "tenant cannot list users",
			userRole:       "tenant",
			expectedStatus: http.StatusForbidden,
			setupMock:      func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			testUser := testutils.CreateTestUser(t, "testuser", "test@example.com", tt.userRole)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)

			// Add user to context
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Greater(t, len(response), 0)
			}

			if len(testDB.Mock.ExpectedQuery) > 0 {
				require.NoError(t, testDB.Mock.ExpectationsWereMet())
			}
		})
	}
}

func TestGetUser_AdminOnly(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	adminUser := testutils.CreateTestUser(t, "admin", "admin@test.com", "admin")
	targetUserID := 2

	// Mock getting user by ID
	userRows := testutils.MockUserRows().AddRow(
		targetUserID, nil, "targetuser", "target@test.com", "Target", "User",
		nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
	)

	testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
		WithArgs(targetUserID).
		WillReturnRows(userRows)

	// Mock roles query
	rolesRows := sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"})
	testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
		WithArgs(targetUserID).
		WillReturnRows(rolesRows)

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+strconv.Itoa(targetUserID), nil)

	// Set up chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(targetUserID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "targetuser", response.Username)

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestUpdateUser_AdminOnly(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	adminUser := testutils.CreateTestUser(t, "admin", "admin@test.com", "admin")
	targetUserID := 2

	updateData := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "User",
		"status":     "suspended",
	}

	// Mock getting user by ID
	userRows := testutils.MockUserRows().AddRow(
		targetUserID, nil, "targetuser", "target@test.com", "Target", "User",
		nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
	)

	testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
		WithArgs(targetUserID).
		WillReturnRows(userRows)

	// Mock roles query for GetUserByID
	rolesRows := sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"})
	testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
		WithArgs(targetUserID).
		WillReturnRows(rolesRows)

	// Mock update user
	testDB.Mock.ExpectExec(`UPDATE users SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payloadBytes, err := json.Marshal(updateData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/users/"+strconv.Itoa(targetUserID), bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Set up chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(targetUserID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.User
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", response.FirstName)
	assert.Equal(t, "suspended", response.Status)

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestDeleteUser_AdminOnly(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	adminUser := testutils.CreateTestUser(t, "admin", "admin@test.com", "admin")
	targetUserID := 2

	// Mock delete user
	testDB.Mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(targetUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/users/"+strconv.Itoa(targetUserID), nil)

	// Set up chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(targetUserID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestAssignRole_AdminOnly(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	adminUser := testutils.CreateTestUser(t, "admin", "admin@test.com", "admin")
	targetUserID := 2
	roleID := 3

	assignData := map[string]int{
		"role_id": roleID,
	}

	// Mock assign role
	testDB.Mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs(targetUserID, roleID, adminUser.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payloadBytes, err := json.Marshal(assignData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/users/"+strconv.Itoa(targetUserID)+"/roles", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Set up chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(targetUserID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role assigned successfully", response["message"])

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestEnableMFA(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	testUser := testutils.CreateTestUser(t, "testuser", "test@example.com", "tenant")
	testUser.MFAEnabled = false // Ensure MFA is disabled initially

	// Mock update user for MFA enable
	testDB.Mock.ExpectExec(`UPDATE users SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/users/mfa/enable", nil)

	// Add user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "secret")
	assert.Contains(t, response, "qr_url")

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}

func TestLogout(t *testing.T) {
	router, testDB := setupUserAPITest(t)
	defer testDB.TeardownTestDB()

	testUser := testutils.CreateTestUser(t, "testuser", "test@example.com", "tenant")

	// Mock session deletion
	testDB.Mock.ExpectExec(`DELETE FROM user_sessions WHERE session_token = \$1`).
		WithArgs("test_session_token").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/users/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "test_session_token"})

	// Add user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, testUser)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])

	// Check that cookies are deleted
	cookies := rr.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "id_token" && cookie.MaxAge == -1 {
			found = true
			break
		}
	}
	assert.True(t, found, "id_token cookie should be deleted")

	require.NoError(t, testDB.Mock.ExpectationsWereMet())
}
