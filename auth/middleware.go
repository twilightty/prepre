package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"jinzmedia-atmt/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(authService *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("AUTH ERROR: Missing Authorization header for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
				writeErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Extract token from "Bearer <token>"
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				log.Printf("AUTH ERROR: Invalid auth header format for %s %s from %s: %s", r.Method, r.URL.Path, r.RemoteAddr, authHeader)
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			token := tokenParts[1]
			log.Printf("AUTH DEBUG: Validating token for %s %s from %s (token: %s...)", r.Method, r.URL.Path, r.RemoteAddr, token[:10])

			// Validate token
			user, err := authService.ValidateToken(token)
			if err != nil {
				log.Printf("AUTH ERROR: Token validation failed for %s %s from %s: %v", r.Method, r.URL.Path, r.RemoteAddr, err)
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			log.Printf("AUTH SUCCESS: User %s authenticated for %s %s", user.Email, r.Method, r.URL.Path)

			// Add user to request context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole is a middleware that requires a specific role
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if user.Role == string(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a middleware that requires admin role
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin, models.RoleSuper)
}

// RequireSuper is a middleware that requires super admin role
func RequireSuper() func(http.Handler) http.Handler {
	return RequireRole(models.RoleSuper)
}

// GetUserFromContext extracts the user from the request context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
