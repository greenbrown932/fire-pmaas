package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/api"
	"github.com/greenbrown932/fire-pmaas/pkg/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"github.com/greenbrown932/fire-pmaas/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// setupRBACTest creates a test environment for RBAC testing
func setupRBACTest(t *testing.T) (*chi.Mux, *testutils.TestDB) {
	testutils.SetTestEnv(t)
	testDB := testutils.SetupTestDB(t)

	router := chi.NewRouter()

	// Add middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would normally be done by LoadUserFromToken middleware
			// For testing, we'll set the user from a header
			userID := r.Header.Get("X-Test-User-ID")
			if userID != "" {
				id, _ := strconv.Atoi(userID)
				user := getUserForRBACTest(id)
				if user != nil {
					ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// Register routes
	api.RegisterRoutes(router)

	return router, testDB
}

// getUserForRBACTest returns test users with different roles
func getUserForRBACTest(userID int) *models.User {
	users := map[int]*models.User{
		1: testutils.CreateTestUser(nil, "admin", "admin@test.com", "admin"),
		2: testutils.CreateTestUser(nil, "manager", "manager@test.com", "property_manager"),
		3: testutils.CreateTestUser(nil, "tenant", "tenant@test.com", "tenant"),
		4: testutils.CreateTestUser(nil, "viewer", "viewer@test.com", "viewer"),
	}

	if user, exists := users[userID]; exists {
		user.ID = userID
		return user
	}
	return nil
}

func TestAdminAccessControl(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	adminUserID := "1" // Admin user

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		// Admin should have access to all endpoints
		{
			name:           "admin_can_list_users",
			method:         "GET",
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to list all users",
		},
		{
			name:           "admin_can_get_user",
			method:         "GET",
			path:           "/api/users/2",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to get any user details",
		},
		{
			name:           "admin_can_update_user",
			method:         "PUT",
			path:           "/api/users/2",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to update any user",
		},
		{
			name:           "admin_can_delete_user",
			method:         "DELETE",
			path:           "/api/users/2",
			expectedStatus: http.StatusNoContent,
			description:    "Admin should be able to delete users",
		},
		{
			name:           "admin_can_access_properties",
			method:         "GET",
			path:           "/properties",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to access properties",
		},
		{
			name:           "admin_can_access_tenants",
			method:         "GET",
			path:           "/tenants",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to access tenant management",
		},
		{
			name:           "admin_can_access_maintenance",
			method:         "GET",
			path:           "/maintenance",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to access maintenance",
		},
		{
			name:           "admin_can_access_admin_panel",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to access admin panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForRBACTest(testDB, tt.method, tt.path)

			var req *http.Request
			if tt.method == "PUT" {
				updateData := map[string]string{"first_name": "Updated"}
				body, _ := json.Marshal(updateData)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			req.Header.Set("X-Test-User-ID", adminUserID)

			// Set up chi URL params for paths with parameters
			if tt.path == "/api/users/2" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", "2")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestPropertyManagerAccessControl(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	managerUserID := "2" // Property Manager user

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		// Property Manager should have limited access
		{
			name:           "manager_can_list_users",
			method:         "GET",
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to list users",
		},
		{
			name:           "manager_can_access_properties",
			method:         "GET",
			path:           "/properties",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to access properties",
		},
		{
			name:           "manager_can_access_tenants",
			method:         "GET",
			path:           "/tenants",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to access tenants",
		},
		{
			name:           "manager_can_access_maintenance",
			method:         "GET",
			path:           "/maintenance",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to access maintenance",
		},
		{
			name:           "manager_can_access_admin_panel",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to access admin panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForRBACTest(testDB, tt.method, tt.path)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Test-User-ID", managerUserID)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestTenantAccessControl(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	tenantUserID := "3" // Tenant user

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		// Tenant should have very limited access
		{
			name:           "tenant_cannot_list_users",
			method:         "GET",
			path:           "/api/users",
			expectedStatus: http.StatusForbidden,
			description:    "Tenant should not be able to list users",
		},
		{
			name:           "tenant_cannot_access_properties_management",
			method:         "GET",
			path:           "/properties",
			expectedStatus: http.StatusForbidden,
			description:    "Tenant should not be able to access property management",
		},
		{
			name:           "tenant_cannot_access_tenant_management",
			method:         "GET",
			path:           "/tenants",
			expectedStatus: http.StatusForbidden,
			description:    "Tenant should not be able to access tenant management",
		},
		{
			name:           "tenant_can_access_maintenance",
			method:         "GET",
			path:           "/maintenance",
			expectedStatus: http.StatusOK,
			description:    "Tenant should be able to access maintenance (to submit requests)",
		},
		{
			name:           "tenant_cannot_access_admin_panel",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusForbidden,
			description:    "Tenant should not be able to access admin panel",
		},
		{
			name:           "tenant_can_access_profile",
			method:         "GET",
			path:           "/api/users/profile",
			expectedStatus: http.StatusOK,
			description:    "Tenant should be able to access their own profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForRBACTest(testDB, tt.method, tt.path)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Test-User-ID", tenantUserID)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestViewerAccessControl(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	viewerUserID := "4" // Viewer user

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		// Viewer should have read-only access to basic information
		{
			name:           "viewer_cannot_list_users",
			method:         "GET",
			path:           "/api/users",
			expectedStatus: http.StatusForbidden,
			description:    "Viewer should not be able to list users",
		},
		{
			name:           "viewer_can_access_properties",
			method:         "GET",
			path:           "/properties",
			expectedStatus: http.StatusOK,
			description:    "Viewer should be able to view properties",
		},
		{
			name:           "viewer_can_access_tenants",
			method:         "GET",
			path:           "/tenants",
			expectedStatus: http.StatusOK,
			description:    "Viewer should be able to view tenants",
		},
		{
			name:           "viewer_can_access_maintenance",
			method:         "GET",
			path:           "/maintenance",
			expectedStatus: http.StatusOK,
			description:    "Viewer should be able to view maintenance requests",
		},
		{
			name:           "viewer_cannot_access_admin_panel",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusForbidden,
			description:    "Viewer should not be able to access admin panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForRBACTest(testDB, tt.method, tt.path)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Test-User-ID", viewerUserID)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestUnauthenticatedAccessControl(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "unauthenticated_redirected_from_dashboard",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusFound, // Redirect to login
			description:    "Unauthenticated users should be redirected from dashboard",
		},
		{
			name:           "unauthenticated_redirected_from_properties",
			method:         "GET",
			path:           "/properties",
			expectedStatus: http.StatusFound, // Redirect to login
			description:    "Unauthenticated users should be redirected from properties",
		},
		{
			name:           "unauthenticated_can_access_health",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Unauthenticated users should be able to access health endpoint",
		},
		{
			name:           "unauthenticated_can_register",
			method:         "POST",
			path:           "/api/users/register",
			expectedStatus: http.StatusBadRequest, // Bad request due to empty body, but accessible
			description:    "Unauthenticated users should be able to access registration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			// No X-Test-User-ID header = unauthenticated

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestAPIPermissionLevels(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name           string
		userRole       string
		userID         string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		// Test role assignment endpoints
		{
			name:           "admin_can_assign_roles",
			userRole:       "admin",
			userID:         "1",
			method:         "POST",
			path:           "/api/users/2/roles",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to assign roles",
		},
		{
			name:           "manager_can_assign_roles",
			userRole:       "property_manager",
			userID:         "2",
			method:         "POST",
			path:           "/api/users/3/roles",
			expectedStatus: http.StatusOK,
			description:    "Property Manager should be able to assign roles",
		},
		{
			name:           "tenant_cannot_assign_roles",
			userRole:       "tenant",
			userID:         "3",
			method:         "POST",
			path:           "/api/users/4/roles",
			expectedStatus: http.StatusForbidden,
			description:    "Tenant should not be able to assign roles",
		},
		{
			name:           "viewer_cannot_assign_roles",
			userRole:       "viewer",
			userID:         "4",
			method:         "POST",
			path:           "/api/users/2/roles",
			expectedStatus: http.StatusForbidden,
			description:    "Viewer should not be able to assign roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForRoleAssignment(testDB)

			assignData := map[string]int{"role_id": 2}
			body, _ := json.Marshal(assignData)

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User-ID", tt.userID)

			// Set up chi URL params
			rctx := chi.NewRouteContext()
			if tt.path == "/api/users/2/roles" {
				rctx.URLParams.Add("id", "2")
			} else if tt.path == "/api/users/3/roles" {
				rctx.URLParams.Add("id", "3")
			} else if tt.path == "/api/users/4/roles" {
				rctx.URLParams.Add("id", "4")
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

func TestCrossUserDataAccess(t *testing.T) {
	router, testDB := setupRBACTest(t)
	defer testDB.TeardownTestDB()

	tests := []struct {
		name           string
		userID         string
		targetUserID   string
		expectedStatus int
		description    string
	}{
		{
			name:           "user_can_access_own_profile",
			userID:         "3",
			targetUserID:   "3",
			expectedStatus: http.StatusOK,
			description:    "User should be able to access their own profile",
		},
		{
			name:           "user_cannot_access_other_profile",
			userID:         "3",
			targetUserID:   "4",
			expectedStatus: http.StatusForbidden,
			description:    "User should not be able to access another user's profile",
		},
		{
			name:           "admin_can_access_any_profile",
			userID:         "1",
			targetUserID:   "3",
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to access any user's profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMockForUserProfile(testDB, tt.targetUserID)

			req := httptest.NewRequest("GET", "/api/users/"+tt.targetUserID, nil)
			req.Header.Set("X-Test-User-ID", tt.userID)

			// Set up chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.targetUserID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.description)
		})
	}
}

// Helper functions for setting up mocks

func setupMockForRBACTest(testDB *testutils.TestDB, method, path string) {
	switch {
	case path == "/api/users" && method == "GET":
		rows := testutils.MockUserRows()
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM users u ORDER BY`).
			WillReturnRows(rows)

	case path == "/api/users/2" && method == "GET":
		userRows := testutils.MockUserRows().AddRow(
			2, nil, "user2", "user2@test.com", "User", "Two",
			nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
		)
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
			WithArgs(2).
			WillReturnRows(userRows)

		rolesRows := sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"})
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
			WithArgs(2).
			WillReturnRows(rolesRows)

	case path == "/api/users/2" && method == "PUT":
		userRows := testutils.MockUserRows().AddRow(
			2, nil, "user2", "user2@test.com", "User", "Two",
			nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
		)
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
			WithArgs(2).
			WillReturnRows(userRows)

		rolesRows := sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"})
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
			WithArgs(2).
			WillReturnRows(rolesRows)

		testDB.Mock.ExpectExec(`UPDATE users SET`).
			WillReturnResult(sqlmock.NewResult(1, 1))

	case path == "/api/users/2" && method == "DELETE":
		testDB.Mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
			WithArgs(2).
			WillReturnResult(sqlmock.NewResult(1, 1))

	case path == "/properties" && method == "GET":
		rows := testutils.MockPropertyDetailRows()
		testDB.Mock.ExpectQuery(`SELECT (.+) FROM properties p`).
			WillReturnRows(rows)

	case path == "/api/users/profile" && method == "GET":
		// No mock needed as user comes from context
	}
}

func setupMockForRoleAssignment(testDB *testutils.TestDB) {
	testDB.Mock.ExpectExec(`INSERT INTO user_roles`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func setupMockForUserProfile(testDB *testutils.TestDB, userID string) {
	id, _ := strconv.Atoi(userID)
	userRows := testutils.MockUserRows().AddRow(
		id, nil, "user"+userID, "user"+userID+"@test.com", "User", "Name",
		nil, nil, true, false, nil, "active", nil, nil, nil, time.Now(), time.Now(),
	)
	testDB.Mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(userRows)

	rolesRows := sqlmock.NewRows([]string{"id", "name", "display_name", "description", "permissions", "created_at", "updated_at"})
	testDB.Mock.ExpectQuery(`SELECT (.+) FROM roles r JOIN user_roles ur`).
		WithArgs(id).
		WillReturnRows(rolesRows)
}
