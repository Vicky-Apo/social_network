package handler

import (
	"net/http"
	"strconv"

	"social-network/backend/internal/transport/http/utils"
	usecasereaction "social-network/backend/internal/usecase/reaction"
	"social-network/backend/pkg/logger"
)

// ReactionHandler serves REST endpoints for reactions
type ReactionHandler struct {
	service *usecasereaction.Service
	log     logger.Logger
}

// NewReactionHandler creates a reaction handler
func NewReactionHandler(service *usecasereaction.Service, log logger.Logger) *ReactionHandler {
	return &ReactionHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "reaction")),
	}
}

// AddPostReaction handles POST /posts/{id}/reactions
func (h *ReactionHandler) AddPostReaction(w http.ResponseWriter, r *http.Request) {
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		logBadRequest(h.log, "reactions.post.toggle", logger.F("post_id", postIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "reactions.post.toggle", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	status, err := h.service.AddPostReaction(r.Context(), postID, req)
	if err != nil {
		logBadRequest(h.log, "reactions.post.toggle", logger.F("post_id", postID), logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": status})
}

// GetPostReactions handles GET /posts/{id}/reactions
func (h *ReactionHandler) GetPostReactions(w http.ResponseWriter, r *http.Request) {
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		logBadRequest(h.log, "reactions.post.list", logger.F("post_id", postIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}

	reactions, err := h.service.GetPostReactions(r.Context(), postID)
	if err != nil {
		logServerError(h.log, "reactions.post.list", err, logger.F("post_id", postID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, reactions)
}

// AddCommentReaction handles POST /comments/{id}/reactions
func (h *ReactionHandler) AddCommentReaction(w http.ResponseWriter, r *http.Request) {
	// Parse comment ID from path parameter
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		logBadRequest(h.log, "reactions.comment.toggle", logger.F("comment_id", commentIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "reactions.comment.toggle", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	status, err := h.service.AddCommentReaction(r.Context(), commentID, req)
	if err != nil {
		logBadRequest(h.log, "reactions.comment.toggle", logger.F("comment_id", commentID), logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": status})
}

// GetCommentReactions handles GET /comments/{id}/reactions
func (h *ReactionHandler) GetCommentReactions(w http.ResponseWriter, r *http.Request) {
	// Parse comment ID from path parameter
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		logBadRequest(h.log, "reactions.comment.list", logger.F("comment_id", commentIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}

	reactions, err := h.service.GetCommentReactions(r.Context(), commentID)
	if err != nil {
		logServerError(h.log, "reactions.comment.list", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, reactions)
}
