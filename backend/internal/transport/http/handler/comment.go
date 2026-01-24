package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid post id")
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
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid post id")
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
