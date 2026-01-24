package handler

import (
	"errors"
	"net/http"

	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/internal/transport/http/utils"
	usecasepost "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/logger"
)

// PostHandler serves REST endpoints for posts.
type PostHandler struct {
	service *usecasepost.Service
	log     logger.Logger
}

// NewPostHandler builds a PostHandler.
func NewPostHandler(service *usecasepost.Service, log logger.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "post")),
	}
}

// List handles GET /posts.
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.List(r.Context())
	if err != nil {
		h.log.Error("failed to list posts", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}

// GetByID handles GET /posts/{id}.
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(r.URL.Path, "/posts/")
	if !ok {
		h.log.Debug("invalid post id in path", logger.F("path", r.URL.Path))
		utils.RespondWithError(w, http.StatusNotFound, "post not found")
		return
	}

	post, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			h.log.Debug("post not found", logger.F("post_id", id))
			utils.RespondWithError(w, http.StatusNotFound, "post not found")
			return
		}
		h.log.Error("failed to get post", err, logger.F("post_id", id))
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, post)
}


