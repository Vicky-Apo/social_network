package handler

import (
	"errors"
	"net/http"
	"strconv"

	domaincomment "social-network/backend/internal/domain/comment"
	"social-network/backend/internal/transport/http/utils"
	usecasecomment "social-network/backend/internal/usecase/comment"
	"social-network/backend/pkg/logger"
)

// CommentHandler serves REST endpoints for comments
type CommentHandler struct {
	service *usecasecomment.Service
	log     logger.Logger
}

// NewCommentHandler creates a comment handler
func NewCommentHandler(service *usecasecomment.Service, log logger.Logger) *CommentHandler {
	return &CommentHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "comment")),
	}
}

// Create handles POST /posts/{id}/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		logBadRequest(h.log, "comments.create", logger.F("post_id", postIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}

	// Parse request body
	var req usecasecomment.CreateCommentRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "comments.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	req.PostID = postID

	// Create comment
	comment, err := h.service.Create(r.Context(), req)
	if err != nil {
		logServerError(h.log, "comments.create", err, logger.F("post_id", postID), logger.F("author_id", req.AuthorID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
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
		logBadRequest(h.log, "comments.list", logger.F("post_id", postIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "comments.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Get comments
	comments, err := h.service.GetByPostID(r.Context(), postID, limit, offset)
	if err != nil {
		if errors.Is(err, domaincomment.ErrNotFound) {
			logNotFound(h.log, "comments.list", logger.F("post_id", postID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentsNotFound)
			return
		}
		logServerError(h.log, "comments.list", err, logger.F("post_id", postID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, comments)
}
