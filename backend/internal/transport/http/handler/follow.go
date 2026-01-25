package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasefollow "social-network/backend/internal/usecase/follow"
)

// FollowHandler serves REST endpoints for follows and follow requests.
type FollowHandler struct {
	service *usecasefollow.Service
}

// NewFollowHandler builds a FollowHandler.
func NewFollowHandler(service *usecasefollow.Service) *FollowHandler {
	return &FollowHandler{service: service}
}

// CreateRequest handles POST /follow-requests.
func (h *FollowHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.TargetID == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "target_id is required")
		return
	}

	result, err := h.service.RequestFollow(r.Context(), requesterID, payload.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, usecasefollow.ErrAlreadyFollowing),
			errors.Is(err, usecasefollow.ErrRequestExists):
			utils.RespondWithError(w, http.StatusConflict, "follow request already exists")
		case errors.Is(err, usecasefollow.ErrCannotFollowSelf):
			utils.RespondWithError(w, http.StatusBadRequest, "cannot follow self")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListRequests handles GET /follow-requests.
func (h *FollowHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	targetID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := h.service.ListRequests(r.Context(), targetID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

// ListSentRequests handles GET /follow-requests/sent.
func (h *FollowHandler) ListSentRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	requesterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := h.service.ListSentRequests(r.Context(), requesterID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

// UpdateRequest handles PATCH /follow-requests/{id}.
func (h *FollowHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/follow-requests/")
	if !ok || remainder != "" {
		utils.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.Status == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "status is required")
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.UpdateRequest(r.Context(), id, actorID, payload.Status); err != nil {
		switch {
		case errors.Is(err, domainfollow.ErrRequestNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "follow request not found")
		case errors.Is(err, usecasefollow.ErrForbidden):
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
		default:
			utils.RespondWithError(w, http.StatusBadRequest, "invalid status")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": payload.Status})
}

// Unfollow handles DELETE /users/{id}/followers.
func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	followerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followingIDStr := r.PathValue("id")
	followingID, err := strconv.ParseInt(followingIDStr, 10, 64)
	if err != nil || followingID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Unfollow(r.Context(), followerID, followingID); err != nil {
		switch {
		case errors.Is(err, usecasefollow.ErrCannotFollowSelf):
			utils.RespondWithError(w, http.StatusBadRequest, "cannot unfollow self")
		case errors.Is(err, usecasefollow.ErrNotFollowing):
			utils.RespondWithError(w, http.StatusConflict, "not following")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "unfollowed"})
}
