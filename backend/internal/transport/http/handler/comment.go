package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	domaincomment "social-network/backend/internal/domain/comment"
	"social-network/backend/internal/transport/http/utils"
	usecasecomment "social-network/backend/internal/usecase/comment"
)

// CommentHandler serves REST endpoints for comments
type CommentHandler struct {
	service *usecasecomment.Service
}

// NewCommentHandler creates a comment handler
func NewCommentHandler(service *usecasecomment.Service) *CommentHandler {
	return &CommentHandler{service: service}
}

// Create handles POST /posts/{id}/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse post ID from URL like /posts/123/comments
	postID, ok := parsePostIDFromCommentsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse request body
	var req usecasecomment.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.PostID = postID

	// Create comment
	comment, err := h.service.Create(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, comment)
}

// GetByPostID handles GET /posts/{id}/comments
func (h *CommentHandler) GetByPostID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse post ID from URL
	postID, ok := parsePostIDFromCommentsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get comments
	comments, err := h.service.GetByPostID(r.Context(), postID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "comments not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to get comments")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, comments)
}

// Helper to parse post ID from /posts/123/comments
func parsePostIDFromCommentsPath(path string) (int64, bool) {
	// Expected format: /posts/123/comments
	if !strings.HasPrefix(path, "/posts/") {
		return 0, false
	}

	trimmed := strings.TrimPrefix(path, "/posts/")
	parts := strings.Split(trimmed, "/")

	if len(parts) < 2 || parts[0] == "" || parts[1] != "comments" {
		return 0, false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
