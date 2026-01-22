package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	domainauth "social-network/backend/internal/domain/auth"
	usecaseauth "social-network/backend/internal/usecase/auth"
)

type contextKey string

const (
	userIDKey         contextKey = "userID"
	sessionCookieName            = "session_token"
)

// Auth returns middleware that validates sessions and protects routes
func Auth(authService *usecaseauth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session token from cookie
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				writeUnauthorized(w, "unauthorized")
				return
			}

			sessionToken := cookie.Value
			if sessionToken == "" {
				writeUnauthorized(w, "unauthorized")
				return
			}

			// Validate session and get user ID
			userID, err := authService.ValidateSession(r.Context(), sessionToken)
			if err != nil {
				if errors.Is(err, domainauth.ErrSessionNotFound) ||
					errors.Is(err, domainauth.ErrSessionExpired) {
					writeUnauthorized(w, "unauthorized")
				} else {
					writeInternalError(w, "internal server error")
				}
				return
			}

			// Inject user ID into context
			ctx := context.WithValue(r.Context(), userIDKey, userID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from the request context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// Helper functions

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func writeInternalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
