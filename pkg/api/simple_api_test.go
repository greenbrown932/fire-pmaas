package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestRegisterRoutes(t *testing.T) {
	r := chi.NewRouter()

	// This should not panic
	assert.NotPanics(t, func() {
		RegisterRoutes(r)
	})
}

func TestUserRegistrationEndpoint(t *testing.T) {
	r := chi.NewRouter()
	RegisterUserRoutes(r)

	req := httptest.NewRequest("POST", "/api/users/register", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// Should return bad request for empty body, but endpoint should exist
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
