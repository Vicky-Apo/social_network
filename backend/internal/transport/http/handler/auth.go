package handler

import (
	"errors"
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/response"
	domainauth "social-network/backend/internal/domain/auth"
	usecaseauth "social-network/backend/internal/usecase/auth"
)

const (
	sessionCookieName = "session_token"
	sessionMaxAge     = 30 * 24 * 60 * 60 // 30 days in seconds
)

// AuthHandler serves REST endpoints for authentication
type AuthHandler struct {
	service *usecaseauth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *usecaseauth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req usecaseauth.RegisterRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainauth.ErrEmailAlreadyExists):
			response.RespondWithError(w, http.StatusConflict, "email already exists")
		case errors.Is(err, usecaseauth.ErrInvalidEmail):
			response.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrWeakPassword):
			response.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrInvalidDateFormat),
			errors.Is(err, usecaseauth.ErrInvalidDateOfBirth),
			errors.Is(err, usecaseauth.ErrUserTooYoung):
			response.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.RespondWithSuccess(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req usecaseauth.LoginRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract user agent and IP address from request
	userAgent := r.UserAgent()
	ipAddress := extractIPAddress(r)

	loginResp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, usecaseauth.ErrInvalidCredentials) {
			response.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Set session cookie
	setSessionCookie(w, loginResp.Token)

	response.RespondWithSuccess(w, http.StatusOK, loginResp)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	// Extract session token from cookie
	sessionToken := extractSessionToken(r)

	if err := h.service.Logout(r.Context(), sessionToken); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Clear session cookie
	clearSessionCookie(w)

	response.RespondWithSuccess(w, http.StatusOK, nil)
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {

	// Extract session token from cookie
	sessionToken := extractSessionToken(r)

	user, err := h.service.GetCurrentUser(r.Context(), sessionToken)
	if err != nil {
		if errors.Is(err, domainauth.ErrSessionNotFound) ||
			errors.Is(err, domainauth.ErrSessionExpired) {
			response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.RespondWithSuccess(w, http.StatusOK, user)
}

// Helper functions

// extractSessionToken retrieves session token from cookie
func extractSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// extractIPAddress extracts client IP from request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// setSessionCookie sets the session cookie
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}

// clearSessionCookie removes the session cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}
