package handler

import (
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/utils"
	usecaseuser "social-network/backend/internal/usecase/user"
)

// UserHandler serves REST endpoints for user listing/searching.
type UserHandler struct {
	service *usecaseuser.Service
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(service *usecaseuser.Service) *UserHandler {
	return &UserHandler{service: service}
}

// ListUsers handles GET /users.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, users)
}

// SearchUsers handles GET /users/search?q=...
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "q is required")
		return
	}

	users, err := h.service.SearchUsers(r.Context(), query)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, users)
}
