package handler

import (
	"errors"
	"net/http"

	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	domainauth "social-network/backend/internal/domain/auth"
	usecaseauth "social-network/backend/internal/usecase/auth"
	"social-network/backend/pkg/logger"
)

// AuthHandlerConfig holds configuration for the auth handler
type AuthHandlerConfig struct {
	CookieName string
	MaxAge     int
}

// AuthHandler serves REST endpoints for authentication
type AuthHandler struct {
	service *usecaseauth.Service
	log     logger.Logger
	config  AuthHandlerConfig
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *usecaseauth.Service, log logger.Logger, cfg AuthHandlerConfig) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "auth")),
		config:  cfg,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req usecaseauth.RegisterRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		h.log.Debug("invalid request body", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainauth.ErrEmailAlreadyExists):
			h.log.Debug("registration failed: email exists", logger.F("email", req.Email))
			utils.RespondWithError(w, http.StatusConflict, "email already exists")
		case errors.Is(err, usecaseauth.ErrInvalidEmail):
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrWeakPassword):
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrInvalidDateFormat),
			errors.Is(err, usecaseauth.ErrInvalidDateOfBirth),
			errors.Is(err, usecaseauth.ErrUserTooYoung):
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.log.Error("registration failed", err, logger.F("email", req.Email))
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecaseauth.LoginRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		h.log.Debug("invalid request body", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract user agent and IP address from request
	userAgent := r.UserAgent()
	ipAddress := utils.ExtractIPAddress(r)

	loginResp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, usecaseauth.ErrInvalidCredentials) {
			h.log.Debug("login failed: invalid credentials", logger.F("email", req.Email), logger.F("ip", ipAddress))
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
		} else {
			h.log.Error("login failed", err, logger.F("email", req.Email), logger.F("ip", ipAddress))
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Set session cookie
	h.setSessionCookie(w, loginResp.Token)

	utils.RespondWithSuccess(w, http.StatusOK, loginResp)
}

// Logout handles POST /auth/logout
// Expects Auth middleware to have validated the session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session token from context (middleware already validated it)
	sessionToken, ok := middleware.GetSessionToken(r.Context())
	if !ok {
		h.log.Warn("logout attempt without session token")
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	if err := h.service.Logout(r.Context(), sessionToken); err != nil {
		h.log.Error("logout failed", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.log.Info("user logged out", logger.F("user_id", userID))

	// Clear session cookie
	h.clearSessionCookie(w)

	utils.RespondWithSuccess(w, http.StatusOK, nil)
}

// Me handles GET /auth/me
// Expects Auth middleware to have validated the session
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (middleware already validated session)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Warn("me endpoint called without user context")
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			h.log.Warn("user not found for valid session", logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
		} else {
			h.log.Error("failed to get user", err, logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, user)
}

// Helper functions

// setSessionCookie sets the session cookie
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   h.config.MaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}

// clearSessionCookie removes the session cookie
func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}
