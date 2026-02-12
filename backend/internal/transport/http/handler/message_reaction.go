package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainchat "social-network/backend/internal/domain/chat"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasemessagereaction "social-network/backend/internal/usecase/message_reaction"
	"social-network/backend/pkg/logger"
)

// MessageReactionHandler serves REST endpoints for message reactions.
type MessageReactionHandler struct {
	service *usecasemessagereaction.Service
	log     logger.Logger
}

// NewMessageReactionHandler builds a MessageReactionHandler.
func NewMessageReactionHandler(service *usecasemessagereaction.Service, log logger.Logger) *MessageReactionHandler {
	return &MessageReactionHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "message_reaction")),
	}
}

// Toggle handles POST /messages/{id}/reactions.
func (h *MessageReactionHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	messageID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || messageID <= 0 {
		logBadRequest(h.log, "messages.reactions.toggle", logger.F("message_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidMessageID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "messages.reactions.toggle")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecasemessagereaction.ToggleReactionRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "messages.reactions.toggle", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	status, err := h.service.ToggleReaction(r.Context(), userID, messageID, req.Emoji)
	if err != nil {
		code, msg := mapMessageReactionError(err)
		if code >= http.StatusInternalServerError {
			logServerError(h.log, "messages.reactions.toggle", err, logger.F("message_id", messageID))
		} else {
			logBadRequest(h.log, "messages.reactions.toggle", logger.F("message_id", messageID), logger.F("reason", msg))
		}
		utils.RespondWithError(w, code, msg)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": status})
}

// List handles GET /messages/{id}/reactions.
func (h *MessageReactionHandler) List(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	messageID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || messageID <= 0 {
		logBadRequest(h.log, "messages.reactions.list", logger.F("message_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidMessageID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "messages.reactions.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	reactions, err := h.service.ListReactions(r.Context(), userID, messageID)
	if err != nil {
		code, msg := mapMessageReactionError(err)
		if code >= http.StatusInternalServerError {
			logServerError(h.log, "messages.reactions.list", err, logger.F("message_id", messageID))
		} else {
			logBadRequest(h.log, "messages.reactions.list", logger.F("message_id", messageID), logger.F("reason", msg))
		}
		utils.RespondWithError(w, code, msg)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, reactions)
}

func mapMessageReactionError(err error) (int, string) {
	switch {
	case errors.Is(err, usecasemessagereaction.ErrInvalidEmoji):
		return http.StatusBadRequest, utils.MsgInvalidEmoji
	case errors.Is(err, usecasemessagereaction.ErrForbidden):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, domainchat.ErrMessageNotFound):
		return http.StatusNotFound, utils.MsgMessageNotFound
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}
