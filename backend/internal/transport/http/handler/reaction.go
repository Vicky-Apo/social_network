package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	postID, ok := parsePostIDFromReactionsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.AddPostReaction(r.Context(), postID, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetPostReactions handles GET /posts/{id}/reactions
func (h *ReactionHandler) GetPostReactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	postID, ok := parsePostIDFromReactionsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reactions, err := h.service.GetPostReactions(r.Context(), postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, reactions)
}

// AddCommentReaction handles POST /comments/{id}/reactions
func (h *ReactionHandler) AddCommentReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	commentID, ok := parseCommentIDFromReactionsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req usecasereaction.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.AddCommentReaction(r.Context(), commentID, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetCommentReactions handles GET /comments/{id}/reactions
func (h *ReactionHandler) GetCommentReactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	commentID, ok := parseCommentIDFromReactionsPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reactions, err := h.service.GetCommentReactions(r.Context(), commentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, reactions)
}

// Helper to parse post ID from /posts/123/reactions
func parsePostIDFromReactionsPath(path string) (int64, bool) {
	if !strings.HasPrefix(path, "/posts/") {
		return 0, false
	}

	trimmed := strings.TrimPrefix(path, "/posts/")
	parts := strings.Split(trimmed, "/")

	if len(parts) < 2 || parts[0] == "" || parts[1] != "reactions" {
		return 0, false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

// Helper to parse comment ID from /comments/123/reactions
func parseCommentIDFromReactionsPath(path string) (int64, bool) {
	if !strings.HasPrefix(path, "/comments/") {
		return 0, false
	}

	trimmed := strings.TrimPrefix(path, "/comments/")
	parts := strings.Split(trimmed, "/")

	if len(parts) < 2 || parts[0] == "" || parts[1] != "reactions" {
		return 0, false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
