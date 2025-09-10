package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"github.com/stretchr/testify/assert"
)

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

func TestRequirePermission(t *testing.T) {
	// Create a test handler that will be called if middleware passes
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test with user that has permission
	user := &models.User{
		ID: 1,
		Roles: []models.Role{
			{
				Name:        "admin",
				Permissions: models.StringArray{"users.read", "users.create"},
			},
		},
	}

	handler := RequirePermission("users.read")(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestRequirePermissionDenied(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test with user that lacks permission
	user := &models.User{
		ID: 1,
		Roles: []models.Role{
			{
				Name:        "viewer",
				Permissions: models.StringArray{"properties.read"},
			},
		},
	}

	handler := RequirePermission("users.delete")(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireRole(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	user := &models.User{
		ID: 1,
		Roles: []models.Role{
			{Name: "admin"},
			{Name: "user"},
		},
	}

	handler := RequireRole("admin")(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
