package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	domainnotification "social-network/backend/internal/domain/notification"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasenotification "social-network/backend/internal/usecase/notification"
	"social-network/backend/pkg/logger"
)

// NotificationHandler serves REST endpoints for notifications.
type NotificationHandler struct {
	service *usecasenotification.Service
	log     logger.Logger
}

// NewNotificationHandler creates a notification handler.
func NewNotificationHandler(service *usecasenotification.Service, log logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "notification")),
	}
}

// List handles GET /notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "notifications.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "notifications.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	unreadOnly := false
	if raw := r.URL.Query().Get("unread"); raw != "" {
		v := strings.ToLower(raw)
		switch v {
		case "true", "1", "yes":
			unreadOnly = true
		case "false", "0", "no":
			unreadOnly = false
		default:
			logBadRequest(h.log, "notifications.list", logger.F("unread", raw))
			utils.RespondWithError(w, http.StatusBadRequest, "invalid unread value")
			return
		}
	}

	items, err := h.service.List(r.Context(), userID, limit, offset, unreadOnly)
	if err != nil {
		logServerError(h.log, "notifications.list", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, items)
}

// UnreadCount handles GET /notifications/unread-count
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "notifications.unread_count")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	count, err := h.service.UnreadCount(r.Context(), userID)
	if err != nil {
		logServerError(h.log, "notifications.unread_count", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]int64{"count": count})
}

// MarkRead handles PATCH /notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "notifications.read")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "notifications.read", logger.F("notification_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidNotificationID)
		return
	}

	err = h.service.MarkRead(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, domainnotification.ErrNotFound) {
			logNotFound(h.log, "notifications.read", logger.F("notification_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgNotFound)
			return
		}
		logServerError(h.log, "notifications.read", err, logger.F("notification_id", id))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "read"})
}

// MarkAllRead handles PATCH /notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "notifications.read_all")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	count, err := h.service.MarkAllRead(r.Context(), userID)
	if err != nil {
		logServerError(h.log, "notifications.read_all", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]int64{"updated": count})
}
