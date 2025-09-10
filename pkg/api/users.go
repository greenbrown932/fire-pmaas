package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/greenbrown932/fire-pmaas/pkg/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

// RegisterUserRoutes registers all user-related API routes
func RegisterUserRoutes(r chi.Router) {
	// Public routes (no authentication required)
	r.Post("/api/users/register", handleUserRegistration)
	r.Post("/api/users/login", handleUserLogin)
	r.Post("/api/users/password-reset/request", handlePasswordResetRequest)
	r.Post("/api/users/password-reset/confirm", handlePasswordResetConfirm)

	// Protected routes (authentication required)
	r.Group(func(auth chi.Router) {
		auth.Use(middleware.LoadUserFromToken)
		auth.Use(middleware.RequireLogin)

		// User profile management
		auth.Get("/api/users/profile", handleGetProfile)
		auth.Put("/api/users/profile", handleUpdateProfile)
		auth.Post("/api/users/logout", handleLogout)

		// MFA management
		auth.Post("/api/users/mfa/enable", handleEnableMFA)
		auth.Post("/api/users/mfa/disable", handleDisableMFA)
		auth.Post("/api/users/mfa/verify", handleVerifyMFA)

		// Admin-only routes
		auth.Group(func(admin chi.Router) {
			admin.Use(middleware.RequireAnyRole("admin", "property_manager"))

			admin.Get("/api/users", handleListUsers)
			admin.Get("/api/users/{id}", handleGetUser)
			admin.Put("/api/users/{id}", handleUpdateUser)
			admin.Delete("/api/users/{id}", handleDeleteUser)
			admin.Post("/api/users/{id}/roles", handleAssignRole)
			admin.Delete("/api/users/{id}/roles/{roleId}", handleRemoveRole)
			admin.Get("/api/roles", handleGetRoles)
		})
	})
}

// User Registration Handler
func handleUserRegistration(w http.ResponseWriter, r *http.Request) {
	var registration models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if registration.Username == "" || registration.Email == "" || registration.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Check if passwords match
	if registration.Password != registration.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	_, err := models.GetUserByEmail(registration.Email)
	if err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	_, err = models.GetUserByUsername(registration.Username)
	if err == nil {
		http.Error(w, "User with this username already exists", http.StatusConflict)
		return
	}

	// Create new user
	user := &models.User{
		Username:      registration.Username,
		Email:         registration.Email,
		FirstName:     registration.FirstName,
		LastName:      registration.LastName,
		PhoneNumber:   models.NullString(registration.PhoneNumber),
		EmailVerified: false, // Email verification would be implemented separately
		Status:        "active",
	}

	if err := models.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Assign default role
	defaultRole, err := models.GetRoleByName("tenant")
	if err == nil {
		models.AssignRole(user.ID, defaultRole.ID, nil)
	}

	// Return user info (without sensitive data)
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// User Login Handler (session-based)
func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	var login models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For now, redirect to OIDC login since we're using Keycloak
	// This handler could be extended to support direct password authentication
	http.Error(w, "Please use OIDC login flow", http.StatusNotImplemented)
}

// Get User Profile Handler
func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Update User Profile Handler
func handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	var updateData struct {
		FirstName         string `json:"first_name"`
		LastName          string `json:"last_name"`
		PhoneNumber       string `json:"phone_number"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update user fields
	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}
	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}
	user.PhoneNumber = models.NullString(updateData.PhoneNumber)
	user.ProfilePictureURL = models.NullString(updateData.ProfilePictureURL)

	if err := models.UpdateUser(user); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Logout Handler
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Delete session cookie if present
	sessionCookie, err := r.Cookie("session_token")
	if err == nil && sessionCookie.Value != "" {
		models.DeleteUserSession(sessionCookie.Value)
	}

	// Delete ID token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "id_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete the cookie
	})

	// Delete session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete the cookie
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// Enable MFA Handler
func handleEnableMFA(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	if user.MFAEnabled {
		http.Error(w, "MFA is already enabled", http.StatusBadRequest)
		return
	}

	// Generate MFA secret
	secret, err := models.GenerateMFASecret()
	if err != nil {
		http.Error(w, "Failed to generate MFA secret", http.StatusInternalServerError)
		return
	}

	user.MFASecret = models.NullString(secret)
	user.MFAEnabled = true

	if err := models.UpdateUser(user); err != nil {
		http.Error(w, "Failed to enable MFA", http.StatusInternalServerError)
		return
	}

	// Return QR code data for the user to scan
	response := map[string]interface{}{
		"secret": secret,
		"qr_url": fmt.Sprintf("otpauth://totp/Fire-PMAAS:%s?secret=%s&issuer=Fire-PMAAS", user.Email, secret),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Disable MFA Handler
func handleDisableMFA(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	var request struct {
		MFACode string `json:"mfa_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verify MFA code before disabling
	if !user.MFAEnabled || !user.MFASecret.Valid {
		http.Error(w, "MFA is not enabled", http.StatusBadRequest)
		return
	}

	if !models.ValidateMFACode(user.MFASecret.String, request.MFACode) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	user.MFAEnabled = false
	user.MFASecret = sql.NullString{Valid: false}

	if err := models.UpdateUser(user); err != nil {
		http.Error(w, "Failed to disable MFA", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "MFA disabled successfully"})
}

// Verify MFA Handler
func handleVerifyMFA(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	var request struct {
		MFACode string `json:"mfa_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !user.MFAEnabled || !user.MFASecret.Valid {
		http.Error(w, "MFA is not enabled", http.StatusBadRequest)
		return
	}

	valid := models.ValidateMFACode(user.MFASecret.String, request.MFACode)

	response := map[string]interface{}{
		"valid": valid,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Password Reset Request Handler
func handlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmail(request.Email)
	if err != nil {
		// Don't reveal if user exists or not for security
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a reset link has been sent"})
		return
	}

	// Generate reset token
	token, err := models.GenerateResetToken()
	if err != nil {
		http.Error(w, "Failed to generate reset token", http.StatusInternalServerError)
		return
	}

	user.PasswordResetToken = models.NullString(token)
	user.PasswordResetExpires = sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true}

	if err := models.UpdateUser(user); err != nil {
		http.Error(w, "Failed to save reset token", http.StatusInternalServerError)
		return
	}

	// TODO: Send email with reset link
	// For now, just log the token (in production, this would be sent via email)
	fmt.Printf("Password reset token for %s: %s\n", user.Email, token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a reset link has been sent"})
}

// Password Reset Confirm Handler
func handlePasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var reset models.PasswordReset
	if err := json.NewDecoder(r.Body).Decode(&reset); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For OIDC-based auth, password reset would typically be handled by Keycloak
	http.Error(w, "Password reset is handled by the authentication provider", http.StatusNotImplemented)
}

// List Users Handler (Admin only)
func handleListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement pagination
	query := `
		SELECT u.id, u.keycloak_id, u.username, u.email, u.first_name, u.last_name,
			   u.phone_number, u.email_verified, u.mfa_enabled, u.status, u.last_login, u.created_at
		FROM users u
		ORDER BY u.created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user struct {
			ID            int            `json:"id"`
			KeycloakID    sql.NullString `json:"keycloak_id,omitempty"`
			Username      string         `json:"username"`
			Email         string         `json:"email"`
			FirstName     string         `json:"first_name"`
			LastName      string         `json:"last_name"`
			PhoneNumber   sql.NullString `json:"phone_number,omitempty"`
			EmailVerified bool           `json:"email_verified"`
			MFAEnabled    bool           `json:"mfa_enabled"`
			Status        string         `json:"status"`
			LastLogin     sql.NullTime   `json:"last_login,omitempty"`
			CreatedAt     time.Time      `json:"created_at"`
		}

		err := rows.Scan(&user.ID, &user.KeycloakID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.PhoneNumber, &user.EmailVerified,
			&user.MFAEnabled, &user.Status, &user.LastLogin, &user.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan user data", http.StatusInternalServerError)
			return
		}

		// Convert to map for JSON response
		userMap := map[string]interface{}{
			"id":             user.ID,
			"username":       user.Username,
			"email":          user.Email,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"email_verified": user.EmailVerified,
			"mfa_enabled":    user.MFAEnabled,
			"status":         user.Status,
			"created_at":     user.CreatedAt,
		}

		if user.KeycloakID.Valid {
			userMap["keycloak_id"] = user.KeycloakID.String
		}
		if user.PhoneNumber.Valid {
			userMap["phone_number"] = user.PhoneNumber.String
		}
		if user.LastLogin.Valid {
			userMap["last_login"] = user.LastLogin.Time
		}

		users = append(users, userMap)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Get User Handler (Admin only)
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Update User Handler (Admin only)
func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		}
		return
	}

	var updateData struct {
		Username      string `json:"username"`
		Email         string `json:"email"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		PhoneNumber   string `json:"phone_number"`
		EmailVerified *bool  `json:"email_verified,omitempty"`
		Status        string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update fields
	if updateData.Username != "" {
		user.Username = updateData.Username
	}
	if updateData.Email != "" {
		user.Email = updateData.Email
	}
	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}
	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}
	user.PhoneNumber = models.NullString(updateData.PhoneNumber)
	if updateData.EmailVerified != nil {
		user.EmailVerified = *updateData.EmailVerified
	}
	if updateData.Status != "" {
		user.Status = updateData.Status
	}

	if err := models.UpdateUser(user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Delete User Handler (Admin only)
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteUser(userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Assign Role Handler (Admin only)
func handleAssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var request struct {
		RoleID int `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	currentUser, _ := middleware.GetUserFromContext(r.Context())
	assignedBy := &currentUser.ID

	if err := models.AssignRole(userID, request.RoleID, assignedBy); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			http.Error(w, "Role already assigned to user", http.StatusConflict)
		} else {
			http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Role assigned successfully"})
}

// Remove Role Handler (Admin only)
func handleRemoveRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	roleID, err := strconv.Atoi(chi.URLParam(r, "roleId"))
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	if err := models.RemoveRole(userID, roleID); err != nil {
		http.Error(w, "Failed to remove role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Role removed successfully"})
}

// Get Roles Handler (Admin only)
func handleGetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := models.GetAllRoles()
	if err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}
