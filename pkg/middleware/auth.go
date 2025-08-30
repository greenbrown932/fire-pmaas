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

// Call this in main() before starting your server.
func InitOIDC() error {
	ctx := context.Background()
	var err error
	provider, err = oidc.NewProvider(ctx, keycloakIssuer)
	if err != nil {
		return fmt.Errorf("could not connect to OIDC provider: %w", err)
	}
	oidcConfig = &oidc.Config{
		ClientID: clientID,
	}
	oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return nil
}

// Middleware that protects routes and enforces login via Keycloak OIDC.
func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If an ID token cookie is present, verify it before trusting.
		if c, err := r.Cookie("id_token"); err == nil && c.Value != "" {
			if _, err := provider.Verifier(oidcConfig).Verify(r.Context(), c.Value); err == nil {
				next.ServeHTTP(w, r)
				return
			}
			// If verification fails, fall through to start login.
		}

		// Generate a simple per-request state to prevent CSRF in the OAuth2 flow.
		// For production, store it server-side (session) and check on callback.
		state := fmt.Sprintf("%d", time.Now().UnixNano())

		// Build the authorization URL from discovery (authorization_endpoint)
		// via oauth2Config, ensuring it matches the provider's issuer.
		// Scopes were set in oauth2Config; add any extras as needed.
		authURL := oauth2Config.AuthCodeURL(
			state,
			// If redirectURL needs to be overridden per-request, use:
			// oauth2.SetAuthURLParam("redirect_uri", redirectURL),
			// You can also add a nonce:
			// oauth2.SetAuthURLParam("nonce", state),
		)

		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// Middleware that protects routes and enforces login via Keycloak OIDC.
// func RequireLogin(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Check for a valid ID token cookie first
// 		if cookie, err := r.Cookie("id_token"); err == nil && cookie.Value != "" {
// 			// Validate the ID token (important—never trust the cookie blindly)
// 			ctx := r.Context()
// 			_, err := provider.Verifier(oidcConfig).Verify(ctx, cookie.Value)
// 			if err == nil {
// 				next.ServeHTTP(w, r)
// 				return
// 			}
// 			// Invalid token—fall through to redirect
// 		}

// If not authenticated (or token not valid), redirect to Keycloak for login
// 		keycloakAuthURL := "http://localhost:8080/realms/pmaas/protocol/openid-connect/auth" +
// 			"?client_id=" + clientID +
// 			"&redirect_uri=" + redirectURL +
// 			"&response_type=code" +
// 			"&scope=openid%20email%20profile"
// 		http.Redirect(w, r, keycloakAuthURL, http.StatusFound)
// 	})
// }

// Handler for the OIDC callback from Keycloak.
// Exchanges the authorization code for tokens, validates them, and sets session cookie.
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
		Name:     "id_token",
		Value:    rawIDToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with https!
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 hour
	})

	// Redirect to the home/dashboard
	http.Redirect(w, r, "/", http.StatusFound)
}
