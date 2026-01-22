package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	domainuser "social-network/backend/internal/domain/user"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseID(r.URL.Path, "/profiles/")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	viewerID, ok := parseOptionalID(r.URL.Query().Get("viewer_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	profile, err := h.service.GetProfile(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// ListFollowers handles GET /profiles/{id}/followers.
func (h *ProfileHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "followers" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	viewerID, ok := parseOptionalID(r.URL.Query().Get("viewer_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	followers, err := h.service.ListFollowers(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, followers)
}

// ListFollowing handles GET /profiles/{id}/following.
func (h *ProfileHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "following" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	viewerID, ok := parseOptionalID(r.URL.Query().Get("viewer_id"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	following, err := h.service.ListFollowing(r.Context(), id, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, usecaseprofile.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, following)
}

// UpdateVisibility handles PATCH /profiles/{id}/visibility.
func (h *ProfileHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, remainder, ok := parseIDAndRemainder(r.URL.Path, "/profiles/")
	if !ok || remainder != "visibility" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var payload struct {
		IsPublic bool `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.SetVisibility(r.Context(), id, payload.IsPublic); err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
