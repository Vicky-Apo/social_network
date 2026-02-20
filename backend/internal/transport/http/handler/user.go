package handler

import (
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/middleware"
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
	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "users.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "users.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var users []usecaseuser.UserListItemDTO
	if query == "" {
		users, err = h.service.ListUsers(r.Context(), viewerID, limit, offset)
	} else {
		users, err = h.service.SearchUsers(r.Context(), viewerID, query, limit, offset)
	}
	if err != nil {
		logServerError(h.log, "users.list", err, logger.F("query", query))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, users)
}
