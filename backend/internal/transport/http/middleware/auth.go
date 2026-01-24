package middleware

import (
	"context"
	"net/http"

	"social-network/backend/internal/transport/http/utils"
	"social-network/backend/pkg/logger"
)

type contextKey string

const (
	userIDKey       contextKey = "userID"
	sessionTokenKey contextKey = "sessionToken"
)

// SessionValidator defines what the auth middleware needs to validate sessions.
// This interface allows the middleware to remain decoupled from the usecase layer.
type SessionValidator interface {
	ValidateSession(ctx context.Context, token string) (userID int64, err error)
}

// Auth returns middleware that validates sessions and protects routes.
// It accepts any SessionValidator implementation, keeping the transport layer
// decoupled from the usecase layer.
func Auth(validator SessionValidator, cookieName string, log logger.Logger) func(http.Handler) http.Handler {
	log = log.WithFields(logger.F("middleware", "auth"))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.ExtractIPAddress(r)

			// Extract session token from cookie
			var sessionToken string
			cookie, err := r.Cookie(cookieName)
			if err == nil {
				sessionToken = cookie.Value
			}

			if sessionToken == "" {
				log.Debug("missing session token", logger.F("ip", ip), logger.F("path", r.URL.Path))
				utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// Validate session and get user ID
			userID, err := validator.ValidateSession(r.Context(), sessionToken)
			if err != nil {
				log.Debug("session validation failed", logger.F("ip", ip), logger.F("path", r.URL.Path), logger.F("error", err.Error()))
				utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// Inject user ID and session token into context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, sessionTokenKey, sessionToken)

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

// GetSessionToken extracts the session token from the request context
func GetSessionToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(sessionTokenKey).(string)
	return token, ok
}
