package handler

import (
	"errors"
	"net/http"
	"strconv"

	domaincomment "social-network/backend/internal/domain/comment"
	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/internal/transport/http/middleware"
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
	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "comments.create")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil || postID <= 0 {
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
	req.AuthorID = authorID

	// Create comment
	comment, err := h.service.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, usecasecomment.ErrForbidden) {
			logForbidden(h.log, "comments.create", logger.F("post_id", postID), logger.F("author_id", req.AuthorID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domainpost.ErrNotFound) {
			logNotFound(h.log, "comments.create", logger.F("post_id", postID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgPostNotFound)
			return
		}
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
	if err != nil || postID <= 0 {
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
	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "comments.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	// Get comments
	comments, err := h.service.GetByPostID(r.Context(), postID, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, usecasecomment.ErrForbidden) {
			logForbidden(h.log, "comments.list", logger.F("post_id", postID), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domainpost.ErrNotFound) || errors.Is(err, domaincomment.ErrNotFound) {
			logNotFound(h.log, "comments.list", logger.F("post_id", postID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgPostNotFound)
			return
		}
		logServerError(h.log, "comments.list", err, logger.F("post_id", postID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, comments)
}

// Update handles PATCH /comments/{id}
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		logBadRequest(h.log, "comments.update", logger.F("comment_id", commentIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}

	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "comments.update")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecasecomment.UpdateCommentRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "comments.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	updated, err := h.service.Update(r.Context(), commentID, authorID, req)
	if err != nil {
		if errors.Is(err, usecasecomment.ErrForbidden) {
			logForbidden(h.log, "comments.update", logger.F("comment_id", commentID), logger.F("author_id", authorID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domaincomment.ErrNotFound) {
			logNotFound(h.log, "comments.update", logger.F("comment_id", commentID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgNotFound)
			return
		}
		logBadRequest(h.log, "comments.update", logger.F("comment_id", commentID), logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, updated)
}

// Delete handles DELETE /comments/{id}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		logBadRequest(h.log, "comments.delete", logger.F("comment_id", commentIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}

	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "comments.delete")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.Delete(r.Context(), commentID, authorID); err != nil {
		if errors.Is(err, usecasecomment.ErrForbidden) {
			logForbidden(h.log, "comments.delete", logger.F("comment_id", commentID), logger.F("author_id", authorID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domaincomment.ErrNotFound) {
			logNotFound(h.log, "comments.delete", logger.F("comment_id", commentID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgNotFound)
			return
		}
		logServerError(h.log, "comments.delete", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
}
