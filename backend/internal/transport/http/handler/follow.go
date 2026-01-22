package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	domainfollow "social-network/backend/internal/domain/follow"
	domainuser "social-network/backend/internal/domain/user"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		RequesterID int64 `json:"requester_id"`
		TargetID    int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.RequesterID == 0 || payload.TargetID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.service.RequestFollow(r.Context(), payload.RequesterID, payload.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecasefollow.ErrAlreadyFollowing),
			errors.Is(err, usecasefollow.ErrRequestExists):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, usecasefollow.ErrCannotFollowSelf):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// AcceptRequest handles POST /follow-requests/{id}/accept.
func (h *FollowHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/follow-requests/")
	if !ok || remainder != "accept" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var payload struct {
		ActorID int64 `json:"actor_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.ActorID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.AcceptRequest(r.Context(), id, payload.ActorID); err != nil {
		switch {
		case errors.Is(err, domainfollow.ErrRequestNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecasefollow.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeclineRequest handles POST /follow-requests/{id}/decline.
func (h *FollowHandler) DeclineRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/follow-requests/")
	if !ok || remainder != "decline" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var payload struct {
		ActorID int64 `json:"actor_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.ActorID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.DeclineRequest(r.Context(), id, payload.ActorID); err != nil {
		switch {
		case errors.Is(err, domainfollow.ErrRequestNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecasefollow.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unfollow handles POST /unfollow.
func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		FollowerID  int64 `json:"follower_id"`
		FollowingID int64 `json:"following_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.FollowerID == 0 || payload.FollowingID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Unfollow(r.Context(), payload.FollowerID, payload.FollowingID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
