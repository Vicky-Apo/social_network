package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasefollow "social-network/backend/internal/usecase/follow"
	"social-network/backend/pkg/logger"
)

// FollowHandler serves REST endpoints for follows and follow requests.
type FollowHandler struct {
	service *usecasefollow.Service
	log     logger.Logger
}

// NewFollowHandler builds a FollowHandler.
func NewFollowHandler(service *usecasefollow.Service, log logger.Logger) *FollowHandler {
	return &FollowHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "follow")),
	}
}

// CreateRequest handles POST /follow-requests.
func (h *FollowHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "follow.request", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if payload.TargetID == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "target_id is required")
		return
	}

	result, err := h.service.RequestFollow(r.Context(), requesterID, payload.TargetID)
	if err != nil {
		status, message := mapFollowError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "follow.request", err, logger.F("requester_id", requesterID), logger.F("target_id", payload.TargetID))
		} else {
			logBadRequest(h.log, "follow.request", logger.F("requester_id", requesterID), logger.F("target_id", payload.TargetID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, result)
}

// ListRequests handles GET /follow-requests.
func (h *FollowHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	targetID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	requests, err := h.service.ListRequests(r.Context(), targetID)
	if err != nil {
		logServerError(h.log, "follow.requests_incoming", err, logger.F("target_id", targetID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, requests)
}

// ListSentRequests handles GET /follow-requests/sent.
func (h *FollowHandler) ListSentRequests(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	requests, err := h.service.ListSentRequests(r.Context(), requesterID)
	if err != nil {
		logServerError(h.log, "follow.requests_sent", err, logger.F("requester_id", requesterID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, requests)
}

// UpdateRequest handles PATCH /follow-requests/{id}.
func (h *FollowHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "follow.request_update", logger.F("request_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgNotFound)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "follow.request_update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if payload.Status == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "status is required")
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.UpdateRequest(r.Context(), id, actorID, payload.Status); err != nil {
		status, message := mapFollowError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "follow.request_update", err, logger.F("request_id", id), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "follow.request_update", logger.F("request_id", id), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": payload.Status})
}

// Unfollow handles DELETE /users/{id}/followers.
func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	followingIDStr := r.PathValue("id")
	followingID, err := strconv.ParseInt(followingIDStr, 10, 64)
	if err != nil || followingID <= 0 {
		logBadRequest(h.log, "follow.unfollow", logger.F("user_id", followingIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	if err := h.service.Unfollow(r.Context(), followerID, followingID); err != nil {
		status, message := mapFollowError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "follow.unfollow", err, logger.F("follower_id", followerID), logger.F("following_id", followingID))
		} else {
			logBadRequest(h.log, "follow.unfollow", logger.F("follower_id", followerID), logger.F("following_id", followingID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "unfollowed"})
}

func mapFollowError(err error) (int, string) {
	switch {
	case errors.Is(err, domainuser.ErrNotFound):
		return http.StatusNotFound, utils.MsgUserNotFound
	case errors.Is(err, domainfollow.ErrRequestNotFound):
		return http.StatusNotFound, utils.MsgFollowRequestNotFound
	case errors.Is(err, usecasefollow.ErrAlreadyFollowing),
		errors.Is(err, usecasefollow.ErrRequestExists):
		return http.StatusConflict, utils.MsgFollowRequestExists
	case errors.Is(err, usecasefollow.ErrCannotFollowSelf):
		return http.StatusBadRequest, utils.MsgCannotFollowSelf
	case errors.Is(err, usecasefollow.ErrForbidden):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, usecasefollow.ErrRequestNotPending):
		return http.StatusConflict, utils.MsgFollowNotPending
	case errors.Is(err, usecasefollow.ErrNotFollowing):
		return http.StatusConflict, utils.MsgNotFollowing
	case errors.Is(err, usecasefollow.ErrInvalidStatus):
		return http.StatusBadRequest, utils.MsgInvalidStatus
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}
