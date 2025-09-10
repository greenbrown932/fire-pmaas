package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
	"golang.org/x/oauth2"
)

var (
	// You'll want to load these from environment variables in production!
	clientID       = "pmaas-app"                        // Must match Keycloak client exactly
	clientSecret   = "CALTAkts8DxnpjCeD6xSDcavEetqMrxl" // Add your actual client secret here
	keycloakIssuer = os.Getenv("KEYCLOAK_ISSUER")
	redirectURL    = "http://localhost:8000/callback"

	provider     *oidc.Provider
	oidcConfig   *oidc.Config
	oauth2Config oauth2.Config
)

// InitOIDC initializes the OIDC provider and configuration.
// Call this in main() before starting your server.
func InitOIDC() error {
	if keycloakIssuer == "" {
		return fmt.Errorf("KEYCLOAK_ISSUER environment variable is not set")
	}

	ctx := context.Background() // Create a background context
	var err error               // Declare an error variable

	// Initialize the OIDC provider
	provider, err = oidc.NewProvider(ctx, keycloakIssuer)
	if err != nil {
		return fmt.Errorf("could not connect to OIDC provider: %w", err)
	}

	// Configure the OIDC client
	oidcConfig = &oidc.Config{
		ClientID: clientID, // Set the client ID
	}

	// Configure the OAuth2 settings
	oauth2Config = oauth2.Config{
		ClientID:     clientID,                                       // Set the client ID
		ClientSecret: clientSecret,                                   // Set the client secret
		Endpoint:     provider.Endpoint(),                            // Set the endpoint from the provider
		RedirectURL:  redirectURL,                                    // Set the redirect URL
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"}, // Set the scopes
	}
	return nil
}

// RequireLogin is a middleware that protects routes and enforces login via Keycloak OIDC.
func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If an ID token cookie is present, verify it before trusting.
		c, err := r.Cookie("id_token")
		if err == nil && c.Value != "" {
			fmt.Println("id_token cookie found:", c.Value)

			// Verify the ID token
			_, err = provider.Verifier(oidcConfig).Verify(r.Context(), c.Value)
			if err == nil {
				fmt.Println("id_token cookie is valid")
				// If verification is successful, serve the next handler
				next.ServeHTTP(w, r)
				return
			}
			fmt.Println("id_token cookie is invalid:", err)
			// If verification fails, fall through to start login.
		} else {
			fmt.Println("id_token cookie not found:", err)
		}

		// Generate a simple per-request state to prevent CSRF in the OAuth2 flow.
		// For production, store it server-side (session) and check on callback.
		state := fmt.Sprintf("%d", time.Now().UnixNano())

		// Build the authorization URL from discovery (authorization_endpoint)
		// via oauth2Config, ensuring it matches the provider's issuer.
		// Scopes were set in oauth2Config; add any extras as needed.
		authURL := oauth2Config.AuthCodeURL(
			state,
			oauth2.SetAuthURLParam("response_type", "code"),
		)

		// Redirect to the authorization URL
		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// HandleCallback is a handler for the OIDC callback from Keycloak.
// Exchanges the authorization code for tokens, validates them, and sets session cookie.
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Get the request context

	// Get the code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code in callback", http.StatusBadRequest)
		return
	}

	// Do the OAuth2 code-for-token exchange
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract and verify the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token in token response", http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	idToken, err := provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Invalid ID token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Optionally parse user info (claims) from the token
	var claims struct {
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the ID token in a secure httpOnly cookie (for demo only)
	http.SetCookie(w, &http.Cookie{
		Name:     "id_token",           // Cookie name
		Value:    rawIDToken,           // Cookie value
		Path:     "/",                  // Cookie path
		HttpOnly: true,                 // HttpOnly flag
		Secure:   false,                // Set to true in production with https!
		SameSite: http.SameSiteLaxMode, // SameSite attribute
		MaxAge:   3600,                 // 1 hour
	})

	// Redirect to the home/dashboard
	// Log the values for debugging
	fmt.Printf("Code: %s\n", code)
	fmt.Printf("Token: %+v\n", token)
	fmt.Printf("Raw ID Token: %s\n", rawIDToken)
	fmt.Printf("Claims: %+v\n", claims)

	// Get the state from the query parameters
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Missing state in callback", http.StatusBadRequest)
		return
	}
	// TODO: Verify the state against the stored state (session)

	// Set the ID token in a secure httpOnly cookie (for demo only)
	cookie := &http.Cookie{
		Name:     "id_token",           // Cookie name
		Value:    rawIDToken,           // Cookie value
		Path:     "/",                  // Cookie path
		HttpOnly: true,                 // HttpOnly flag
		Secure:   false,                // Set to true in production with https!
		SameSite: http.SameSiteLaxMode, // SameSite attribute
		MaxAge:   3600,                 // 1 hour
	}
	http.SetCookie(w, cookie)
	fmt.Println("Cookie set")

	// Redirect to the home/dashboard (clear query params to avoid loops)
	fmt.Println("Redirecting to /")
	http.Redirect(w, r, "/", http.StatusFound)
}

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// RequirePermission is a middleware that checks if the user has the required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !user.HasPermission(permission) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole is a middleware that checks if the user has the required role
func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !user.HasRole(roleName) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is a middleware that checks if the user has any of the required roles
func RequireAnyRole(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, roleName := range roleNames {
				if user.HasRole(roleName) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoadUserFromToken is a middleware that loads user information from OIDC token
func LoadUserFromToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get user from ID token cookie
		c, err := r.Cookie("id_token")
		if err == nil && c.Value != "" {
			// Verify the ID token
			idToken, err := provider.Verifier(oidcConfig).Verify(r.Context(), c.Value)
			if err == nil {
				// Extract claims from the token
				var claims struct {
					Subject           string `json:"sub"`
					Email             string `json:"email"`
					EmailVerified     bool   `json:"email_verified"`
					Name              string `json:"name"`
					PreferredUsername string `json:"preferred_username"`
					GivenName         string `json:"given_name"`
					FamilyName        string `json:"family_name"`
				}

				if err := idToken.Claims(&claims); err == nil {
					// Try to find existing user by Keycloak ID
					user, err := models.GetUserByKeycloakID(claims.Subject)
					if err != nil {
						// User doesn't exist, create one
						user = &models.User{
							KeycloakID:    models.NullString(claims.Subject),
							Username:      claims.PreferredUsername,
							Email:         claims.Email,
							FirstName:     claims.GivenName,
							LastName:      claims.FamilyName,
							EmailVerified: claims.EmailVerified,
							Status:        "active",
						}

						// Create the user in the database
						if err := models.CreateUser(user); err == nil {
							// Assign default role to new users
							defaultRole, err := models.GetRoleByName("tenant")
							if err == nil {
								models.AssignRole(user.ID, defaultRole.ID, nil)
								// Reload user with roles
								user, _ = models.GetUserByID(user.ID)
							}
						}
					}

					if user != nil {
						// Add user to request context
						ctx := context.WithValue(r.Context(), UserContextKey, user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SessionAuth is a middleware for session-based authentication (alternative to OIDC)
func SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for session token in cookie
		sessionCookie, err := r.Cookie("session_token")
		if err == nil && sessionCookie.Value != "" {
			session, err := models.GetUserSession(sessionCookie.Value)
			if err == nil {
				user, err := models.GetUserByID(session.UserID)
				if err == nil {
					// Add user to request context
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateUserSession creates a new session for a user
func CreateUserSession(userID int, r *http.Request) (*models.UserSession, error) {
	token, err := GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	session := &models.UserSession{
		UserID:       userID,
		SessionToken: token,
		IPAddress:    models.NullString(getClientIP(r)),
		UserAgent:    models.NullString(r.UserAgent()),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour session
	}

	err = models.CreateUserSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return strings.Split(forwarded, ",")[0]
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
