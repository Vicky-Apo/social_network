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

// ListUsers handles GET /users (optional search with ?q=).
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var (
		users []usecaseuser.UserListItemDTO
		err   error
	)
	if query == "" {
		users, err = h.service.ListUsers(r.Context())
	} else {
		users, err = h.service.SearchUsers(r.Context(), query)
	}
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, users)
}
