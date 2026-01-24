package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	domainuser "social-network/backend/internal/domain/user"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecaseprofile "social-network/backend/internal/usecase/profile"
)

// ProfileHandler serves REST endpoints for profiles.
type ProfileHandler struct {
	service *usecaseprofile.Service
}

// NewProfileHandler builds a ProfileHandler.
func NewProfileHandler(service *usecaseprofile.Service) *ProfileHandler {
	return &ProfileHandler{service: service}
}

// GetProfile handles GET /profiles/{id}.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, ok := parseID(r.URL.Path, "/profiles/")
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "profile not found")
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "profile not found")
		case errors.Is(err, usecaseprofile.ErrForbidden):
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// ListFollowers handles GET /profiles/{id}/followers.
func (h *ProfileHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "followers" {
		utils.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followers, err := h.service.ListFollowers(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "profile not found")
		case errors.Is(err, usecaseprofile.ErrForbidden):
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, followers)
}

// ListFollowing handles GET /profiles/{id}/following.
func (h *ProfileHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "following" {
		utils.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	following, err := h.service.ListFollowing(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "profile not found")
		case errors.Is(err, usecaseprofile.ErrForbidden):
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, following)
}

// UpdateVisibility handles PATCH /profiles/{id}/visibility.
func (h *ProfileHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "visibility" {
		utils.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}

	var payload struct {
		IsPublic bool `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.SetVisibility(r.Context(), id, actorID, payload.IsPublic); err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "profile not found")
		case errors.Is(err, usecaseprofile.ErrForbidden):
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
