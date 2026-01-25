package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/internal/transport/http/middleware"
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
	limit, offset, err := parsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	viewerID, _ := middleware.GetUserID(r.Context())

	if rawCategory := r.URL.Query().Get("category_id"); rawCategory != "" {
		categoryID, err := strconv.ParseInt(rawCategory, 10, 64)
		if err != nil || categoryID <= 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		posts, err := h.service.ListByCategory(r.Context(), categoryID, viewerID, limit, offset)
		if err != nil {
			h.log.Error("failed to list posts by category", err, logger.F("category_id", categoryID))
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, posts)
		return
	}

	if rawAuthor := r.URL.Query().Get("author_id"); rawAuthor != "" {
		authorID, err := strconv.ParseInt(rawAuthor, 10, 64)
		if err != nil || authorID <= 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid author_id")
			return
		}
		h.ListByAuthor(w, r, authorID, limit, offset)
		return
	}

	posts, err := h.service.List(r.Context(), viewerID, limit, offset)
	if err != nil {
		h.log.Error("failed to list posts", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}

// Create handles POST /posts.
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req usecasepost.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := h.service.Create(r.Context(), authorID, req)
	if err != nil {
		h.log.Error("failed to create post", err, logger.F("author_id", authorID))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, post)
}

// GetByID handles GET /posts/{id}.
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(r.URL.Path, "/posts/")
	if !ok {
		h.log.Debug("invalid post id in path", logger.F("path", r.URL.Path))
		utils.RespondWithError(w, http.StatusNotFound, "post not found")
		return
	}

	viewerID, _ := middleware.GetUserID(r.Context())
	post, err := h.service.GetByID(r.Context(), id, viewerID)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
			return
		}
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

func parsePagination(r *http.Request) (int, int, error) {
	const (
		defaultLimit = 20
		maxLimit     = 100
	)

	limit := defaultLimit
	offset := 0

	if raw := r.URL.Query().Get("limit"); raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil || val <= 0 {
			return 0, 0, errors.New("invalid limit")
		}
		if val > maxLimit {
			val = maxLimit
		}
		limit = val
	}

	if raw := r.URL.Query().Get("offset"); raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil || val < 0 {
			return 0, 0, errors.New("invalid offset")
		}
		offset = val
	}

	return limit, offset, nil
}

// ListByAuthor handles GET /posts?author_id=...&limit=&offset=
func (h *PostHandler) ListByAuthor(w http.ResponseWriter, r *http.Request, authorID int64, limit, offset int) {
	viewerID, _ := middleware.GetUserID(r.Context())
	posts, err := h.service.ListByAuthor(r.Context(), authorID, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			utils.RespondWithError(w, http.StatusForbidden, "forbidden")
			return
		}
		h.log.Error("failed to list posts by author", err, logger.F("author_id", authorID))
		utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}
