package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecaseprofile "social-network/backend/internal/usecase/profile"
	"social-network/backend/pkg/logger"
)

// ProfileHandler serves REST endpoints for profiles.
type ProfileHandler struct {
	service *usecaseprofile.Service
	log     logger.Logger
}

// NewProfileHandler builds a ProfileHandler.
func NewProfileHandler(service *usecaseprofile.Service, log logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "profile")),
	}
}

// GetProfile handles GET /profiles/{id}.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "profiles.get", logger.F("user_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "profiles.get")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	profile, err := h.service.GetProfile(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			logNotFound(h.log, "profiles.get", logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgProfileNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			logForbidden(h.log, "profiles.get", logger.F("profile_id", id), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		default:
			logServerError(h.log, "profiles.get", err, logger.F("profile_id", id), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, profile)
}

// ListFollowers handles GET /profiles/{id}/followers.
func (h *ProfileHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "profiles.followers", logger.F("user_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "profiles.followers")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	followers, err := h.service.ListFollowers(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			logNotFound(h.log, "profiles.followers", logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgProfileNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			logForbidden(h.log, "profiles.followers", logger.F("profile_id", id), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		default:
			logServerError(h.log, "profiles.followers", err, logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, followers)
}

// ListFollowing handles GET /profiles/{id}/following.
func (h *ProfileHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "profiles.following", logger.F("user_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "profiles.following")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	following, err := h.service.ListFollowing(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			logNotFound(h.log, "profiles.following", logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgProfileNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			logForbidden(h.log, "profiles.following", logger.F("profile_id", id), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		default:
			logServerError(h.log, "profiles.following", err, logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, following)
}

// UpdateVisibility handles PATCH /profiles/{id}/visibility.
func (h *ProfileHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "profiles.visibility", logger.F("user_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	var payload struct {
		IsPublic bool `json:"is_public"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "profiles.visibility", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "profiles.visibility")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.SetVisibility(r.Context(), id, actorID, payload.IsPublic); err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			logNotFound(h.log, "profiles.visibility", logger.F("profile_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgProfileNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			logForbidden(h.log, "profiles.visibility", logger.F("profile_id", id), logger.F("actor_id", actorID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		default:
			logServerError(h.log, "profiles.visibility", err, logger.F("profile_id", id), logger.F("actor_id", actorID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]any{
		"status":    "updated",
		"is_public": payload.IsPublic,
	})
}
