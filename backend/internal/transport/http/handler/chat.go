package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainchat "social-network/backend/internal/domain/chat"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasechat "social-network/backend/internal/usecase/chat"
	"social-network/backend/pkg/logger"
)

// ChatHandler serves REST endpoints for conversations and messages.
type ChatHandler struct {
	service *usecasechat.Service
	log     logger.Logger
}

// NewChatHandler builds a ChatHandler.
func NewChatHandler(service *usecasechat.Service, log logger.Logger) *ChatHandler {
	return &ChatHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "chat")),
	}
}

// ListConversations handles GET /conversations.
func (h *ChatHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "conversations.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	conversations, err := h.service.ListConversations(r.Context(), userID)
	if err != nil {
		logServerError(h.log, "conversations.list", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, conversations)
}

// GetConversation handles GET /conversations/{id}.
func (h *ChatHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	conversationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || conversationID <= 0 {
		logBadRequest(h.log, "conversations.get", logger.F("conversation_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidConversationID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "conversations.get")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	conversation, err := h.service.GetConversationByID(r.Context(), userID, conversationID)
	if err != nil {
		switch {
		case errors.Is(err, usecasechat.ErrForbidden):
			logForbidden(h.log, "conversations.get", logger.F("conversation_id", conversationID), logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		case errors.Is(err, domainchat.ErrConversationNotFound):
			logNotFound(h.log, "conversations.get", logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgConversationNotFound)
		default:
			logServerError(h.log, "conversations.get", err, logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, conversation)
}

// ListMessages handles GET /conversations/{id}/messages.
func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	conversationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || conversationID <= 0 {
		logBadRequest(h.log, "messages.list", logger.F("conversation_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidConversationID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "messages.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "messages.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.service.GetConversationMessages(r.Context(), userID, conversationID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, usecasechat.ErrForbidden):
			logForbidden(h.log, "messages.list", logger.F("conversation_id", conversationID), logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		case errors.Is(err, domainchat.ErrConversationNotFound):
			logNotFound(h.log, "messages.list", logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgConversationNotFound)
		default:
			logServerError(h.log, "messages.list", err, logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, messages)
}

// MarkRead handles PATCH /conversations/{id}/read.
func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	conversationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || conversationID <= 0 {
		logBadRequest(h.log, "conversations.read", logger.F("conversation_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidConversationID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "conversations.read")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.MarkAsRead(r.Context(), userID, conversationID); err != nil {
		switch {
		case errors.Is(err, usecasechat.ErrForbidden):
			logForbidden(h.log, "conversations.read", logger.F("conversation_id", conversationID), logger.F("user_id", userID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		case errors.Is(err, domainchat.ErrConversationNotFound):
			logNotFound(h.log, "conversations.read", logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgConversationNotFound)
		default:
			logServerError(h.log, "conversations.read", err, logger.F("conversation_id", conversationID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "read"})
}

// UnreadCounts handles GET /conversations/unread-counts.
func (h *ChatHandler) UnreadCounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "conversations.unread_counts")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	counts, err := h.service.GetUnreadConversations(r.Context(), userID)
	if err != nil {
		logServerError(h.log, "conversations.unread_counts", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, counts)
}
