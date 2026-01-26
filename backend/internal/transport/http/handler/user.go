package handler

import (
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/utils"
	usecaseuser "social-network/backend/internal/usecase/user"
	"social-network/backend/pkg/logger"
)

// UserHandler serves REST endpoints for user listing/searching.
type UserHandler struct {
	service *usecaseuser.Service
	log     logger.Logger
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(service *usecaseuser.Service, log logger.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "user")),
	}
}

// ListUsers handles GET /users (optional search with ?q=).
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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
		logServerError(h.log, "users.list", err, logger.F("query", query))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, users)
}
