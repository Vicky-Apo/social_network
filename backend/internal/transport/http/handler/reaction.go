package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"social-network/backend/internal/transport/http/utils"
	usecasereaction "social-network/backend/internal/usecase/reaction"
)

// ReactionHandler serves REST endpoints for reactions
type ReactionHandler struct {
	service *usecasereaction.Service
}

// NewReactionHandler creates a reaction handler
func NewReactionHandler(service *usecasereaction.Service) *ReactionHandler {
	return &ReactionHandler{service: service}
}

// AddPostReaction handles POST /posts/{id}/reactions
func (h *ReactionHandler) AddPostReaction(w http.ResponseWriter, r *http.Request) {
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.AddPostReaction(r.Context(), postID, req); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to add reaction")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, map[string]string{"message": "reaction added"})
}

// GetPostReactions handles GET /posts/{id}/reactions
func (h *ReactionHandler) GetPostReactions(w http.ResponseWriter, r *http.Request) {
	// Parse post ID from path parameter
	postIDStr := r.PathValue("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	reactions, err := h.service.GetPostReactions(r.Context(), postID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to get reactions")
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
		utils.RespondWithError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.AddCommentReaction(r.Context(), commentID, req); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to add reaction")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, map[string]string{"message": "reaction added"})
}

// GetCommentReactions handles GET /comments/{id}/reactions
func (h *ReactionHandler) GetCommentReactions(w http.ResponseWriter, r *http.Request) {
	// Parse comment ID from path parameter
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	reactions, err := h.service.GetCommentReactions(r.Context(), commentID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to get reactions")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, reactions)
}
