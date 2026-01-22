package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"social-network/backend/internal/transport/http/response"
	domainpost "social-network/backend/internal/domain/post"
	usecasepost "social-network/backend/internal/usecase/post"
)

// PostHandler serves REST endpoints for posts.
type PostHandler struct {
	service *usecasepost.Service
}

// NewPostHandler builds a PostHandler.
func NewPostHandler(service *usecasepost.Service) *PostHandler {
	return &PostHandler{service: service}
}

// List handles GET /posts.
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {

	posts, err := h.service.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.RespondWithSuccess(w, http.StatusOK, posts)
}

// GetByID handles GET /posts/{id}.
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	id, ok := parseID(r.URL.Path, "/posts/")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	post, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.RespondWithSuccess(w, http.StatusOK, post)
}

func parseID(path, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(path, prefix)
	if raw == "" {
		return 0, false
	}
	if strings.Contains(raw, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}


