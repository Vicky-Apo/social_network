package handler

import (
	"errors"
	"net/http"

	domainauth "social-network/backend/internal/domain/auth"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
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
		logBadRequest(h.log, "auth.register", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainauth.ErrEmailAlreadyExists):
			logBadRequest(h.log, "auth.register", logger.F("reason", "email_exists"), logger.F("email", req.Email))
			utils.RespondWithError(w, http.StatusConflict, "email already exists")
		case errors.Is(err, usecaseauth.ErrInvalidEmail):
			logBadRequest(h.log, "auth.register", logger.F("reason", "invalid_email"), logger.F("email", req.Email))
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrWeakPassword):
			logBadRequest(h.log, "auth.register", logger.F("reason", "weak_password"))
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecaseauth.ErrInvalidDateFormat),
			errors.Is(err, usecaseauth.ErrInvalidDateOfBirth),
			errors.Is(err, usecaseauth.ErrUserTooYoung):
			logBadRequest(h.log, "auth.register", logger.F("reason", "invalid_birthdate"))
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			logServerError(h.log, "auth.register", err, logger.F("email", req.Email))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecaseauth.LoginRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "auth.login", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	// Extract user agent and IP address from request
	userAgent := r.UserAgent()
	ipAddress := utils.ExtractIPAddress(r)

	loginResp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, usecaseauth.ErrInvalidCredentials) {
			logUnauthorized(h.log, "auth.login", logger.F("reason", "invalid_credentials"), logger.F("email", req.Email), logger.F("ip", ipAddress))
			utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgInvalidCredentials)
		} else {
			logServerError(h.log, "auth.login", err, logger.F("email", req.Email), logger.F("ip", ipAddress))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	// Set session cookie
	h.setSessionCookie(w, loginResp.Token)

	utils.RespondWithSuccess(w, http.StatusOK, loginResp)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var sessionToken string
	if cookie, err := r.Cookie(h.config.CookieName); err == nil {
		sessionToken = cookie.Value
	}
	if sessionToken != "" {
		if err := h.service.Logout(r.Context(), sessionToken); err != nil {
			logServerError(h.log, "auth.logout", err)
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
			return
		}
		h.log.Info("logout succeeded", logger.F("action", "auth.logout"))
	}

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
		logUnauthorized(h.log, "auth.me")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			logNotFound(h.log, "auth.me", logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgUserNotFound)
		} else {
			logServerError(h.log, "auth.me", err, logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
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
