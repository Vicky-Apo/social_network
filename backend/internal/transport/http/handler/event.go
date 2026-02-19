package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainevent "social-network/backend/internal/domain/event"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecaseevent "social-network/backend/internal/usecase/event"
	"social-network/backend/pkg/logger"
)

// EventHandler serves REST endpoints for events.
type EventHandler struct {
	service *usecaseevent.Service
	log     logger.Logger
}

// NewEventHandler builds an EventHandler.
func NewEventHandler(service *usecaseevent.Service, log logger.Logger) *EventHandler {
	return &EventHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "event")),
	}
}

// Create handles POST /groups/{id}/events.
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "events.create", logger.F("group_id", groupIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	creatorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.create")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecaseevent.CreateEventRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "events.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	event, err := h.service.CreateEvent(r.Context(), creatorID, groupID, req)
	if err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.create", err, logger.F("group_id", groupID), logger.F("creator_id", creatorID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.create", logger.F("group_id", groupID), logger.F("creator_id", creatorID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.create", logger.F("group_id", groupID))
			default:
				logBadRequest(h.log, "events.create", logger.F("group_id", groupID), logger.F("creator_id", creatorID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, event)
}

// ListByGroup handles GET /groups/{id}/events.
func (h *EventHandler) ListByGroup(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "events.list", logger.F("group_id", groupIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "events.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	events, err := h.service.ListGroupEvents(r.Context(), groupID, viewerID, limit, offset)
	if err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.list", err, logger.F("group_id", groupID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.list", logger.F("group_id", groupID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.list", logger.F("group_id", groupID))
			default:
				logBadRequest(h.log, "events.list", logger.F("group_id", groupID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, events)
}

// GetByID handles GET /events/{id}.
func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		logBadRequest(h.log, "events.get", logger.F("event_id", eventIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.get")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	event, err := h.service.GetEvent(r.Context(), eventID, viewerID)
	if err != nil {
		switch {
		case errors.Is(err, domainevent.ErrNotFound):
			logNotFound(h.log, "events.get", logger.F("event_id", eventID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgEventNotFound)
		case errors.Is(err, usecaseevent.ErrForbidden):
			logForbidden(h.log, "events.get", logger.F("event_id", eventID), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
		default:
			logServerError(h.log, "events.get", err, logger.F("event_id", eventID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, event)
}

// Respond handles POST /events/{id}/responses.
func (h *EventHandler) Respond(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		logBadRequest(h.log, "events.respond", logger.F("event_id", eventIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.respond")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecaseevent.RespondRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "events.respond", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	resp, err := h.service.Respond(r.Context(), eventID, userID, req.Response)
	if err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.respond", err, logger.F("event_id", eventID), logger.F("user_id", userID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.respond", logger.F("event_id", eventID), logger.F("user_id", userID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.respond", logger.F("event_id", eventID))
			default:
				logBadRequest(h.log, "events.respond", logger.F("event_id", eventID), logger.F("user_id", userID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, resp)
}

// ListResponses handles GET /events/{id}/responses.
func (h *EventHandler) ListResponses(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		logBadRequest(h.log, "events.responses", logger.F("event_id", eventIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.responses")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	items, err := h.service.ListResponses(r.Context(), eventID, viewerID)
	if err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.responses", err, logger.F("event_id", eventID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.responses", logger.F("event_id", eventID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.responses", logger.F("event_id", eventID))
			default:
				logBadRequest(h.log, "events.responses", logger.F("event_id", eventID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, items)
}

// Update handles PATCH /events/{id}.
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		logBadRequest(h.log, "events.update", logger.F("event_id", eventIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.update")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecaseevent.UpdateEventRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "events.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	updated, err := h.service.UpdateEvent(r.Context(), eventID, userID, req)
	if err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.update", err, logger.F("event_id", eventID), logger.F("user_id", userID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.update", logger.F("event_id", eventID), logger.F("user_id", userID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.update", logger.F("event_id", eventID))
			default:
				logBadRequest(h.log, "events.update", logger.F("event_id", eventID), logger.F("user_id", userID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, updated)
}

// Delete handles DELETE /events/{id}.
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		logBadRequest(h.log, "events.delete", logger.F("event_id", eventIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "events.delete")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.DeleteEvent(r.Context(), eventID, userID); err != nil {
		status, message := mapEventError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "events.delete", err, logger.F("event_id", eventID), logger.F("user_id", userID))
		} else {
			switch status {
			case http.StatusForbidden:
				logForbidden(h.log, "events.delete", logger.F("event_id", eventID), logger.F("user_id", userID))
			case http.StatusNotFound:
				logNotFound(h.log, "events.delete", logger.F("event_id", eventID))
			default:
				logBadRequest(h.log, "events.delete", logger.F("event_id", eventID), logger.F("user_id", userID), logger.F("reason", message))
			}
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func mapEventError(err error) (int, string) {
	switch {
	case errors.Is(err, usecaseevent.ErrInvalidTitle):
		return http.StatusBadRequest, utils.MsgInvalidEventTitle
	case errors.Is(err, usecaseevent.ErrInvalidEventTime):
		return http.StatusBadRequest, utils.MsgInvalidEventTime
	case errors.Is(err, usecaseevent.ErrInvalidResponse):
		return http.StatusBadRequest, utils.MsgInvalidEventResponse
	case errors.Is(err, usecaseevent.ErrForbidden):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, usecaseevent.ErrNotCreator):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, domainevent.ErrNotFound):
		return http.StatusNotFound, utils.MsgEventNotFound
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}
