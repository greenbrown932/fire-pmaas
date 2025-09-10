package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
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
