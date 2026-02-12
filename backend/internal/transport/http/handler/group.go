package handler

import (
	"errors"
	"net/http"
	"strconv"

	domaingroup "social-network/backend/internal/domain/group"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasegroup "social-network/backend/internal/usecase/group"
	"social-network/backend/pkg/logger"
)

// GroupHandler serves REST endpoints for groups.
type GroupHandler struct {
	service *usecasegroup.Service
	log     logger.Logger
}

// NewGroupHandler builds a GroupHandler.
func NewGroupHandler(service *usecasegroup.Service, log logger.Logger) *GroupHandler {
	return &GroupHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "group")),
	}
}

// Create handles POST /groups.
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.create")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecasegroup.CreateGroupRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	group, err := h.service.CreateGroup(r.Context(), creatorID, req)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.create", err, logger.F("creator_id", creatorID))
		} else {
			logBadRequest(h.log, "groups.create", logger.F("creator_id", creatorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, group)
}

// List handles GET /groups.
func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "groups.list", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	query := r.URL.Query().Get("q")

	groups, err := h.service.ListGroups(r.Context(), userID, query, limit, offset)
	if err != nil {
		logServerError(h.log, "groups.list", err, logger.F("query", query))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, groups)
}

// GetByID handles GET /groups/{id}.
func (h *GroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.get", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.get")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	group, err := h.service.GetGroup(r.Context(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, domaingroup.ErrGroupNotFound):
			logNotFound(h.log, "groups.get", logger.F("group_id", groupID))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgGroupNotFound)
		default:
			logServerError(h.log, "groups.get", err, logger.F("group_id", groupID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, group)
}

// ListMembers handles GET /groups/{id}/members.
func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.members", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.members")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	members, err := h.service.ListMembers(r.Context(), userID, groupID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.members", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.members", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, members)
}

// Invite handles POST /groups/{id}/invitations.
func (h *GroupHandler) Invite(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.invite", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	inviterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.invite")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var payload struct {
		InviteeID int64 `json:"invitee_id"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.invite", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if payload.InviteeID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}

	inv, err := h.service.InviteToGroup(r.Context(), inviterID, groupID, payload.InviteeID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.invite", err, logger.F("group_id", groupID), logger.F("invitee_id", payload.InviteeID))
		} else {
			logBadRequest(h.log, "groups.invite", logger.F("group_id", groupID), logger.F("invitee_id", payload.InviteeID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, inv)
}

// ListInvitations handles GET /group-invitations.
func (h *GroupHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.invitations.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	items, err := h.service.ListInvitations(r.Context(), userID)
	if err != nil {
		logServerError(h.log, "groups.invitations.list", err, logger.F("user_id", userID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, items)
}

// UpdateInvitation handles PATCH /group-invitations/{id}.
func (h *GroupHandler) UpdateInvitation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	invID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || invID <= 0 {
		logBadRequest(h.log, "groups.invitations.update", logger.F("invitation_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidInvitationID)
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.invitations.update")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.invitations.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if payload.Status == "" {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidStatus)
		return
	}

	if err := h.service.UpdateInvitation(r.Context(), invID, actorID, payload.Status); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.invitations.update", err, logger.F("invitation_id", invID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.invitations.update", logger.F("invitation_id", invID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": payload.Status})
}

// RequestJoin handles POST /groups/{id}/join-requests.
func (h *GroupHandler) RequestJoin(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.join.request", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.join.request")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	req, err := h.service.RequestJoin(r.Context(), groupID, userID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join.request", err, logger.F("group_id", groupID), logger.F("user_id", userID))
		} else {
			logBadRequest(h.log, "groups.join.request", logger.F("group_id", groupID), logger.F("user_id", userID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, req)
}

// ListJoinRequests handles GET /groups/{id}/join-requests.
func (h *GroupHandler) ListJoinRequests(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.join.list", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.join.list")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	items, err := h.service.ListJoinRequests(r.Context(), groupID, actorID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join.list", err, logger.F("group_id", groupID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.join.list", logger.F("group_id", groupID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, items)
}

// UpdateJoinRequest handles PATCH /group-join-requests/{id}.
func (h *GroupHandler) UpdateJoinRequest(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	reqID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || reqID <= 0 {
		logBadRequest(h.log, "groups.join.update", logger.F("request_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidJoinRequestID)
		return
	}

	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.join.update")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.join.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if payload.Status == "" {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidStatus)
		return
	}

	if err := h.service.UpdateJoinRequest(r.Context(), reqID, actorID, payload.Status); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join.update", err, logger.F("request_id", reqID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.join.update", logger.F("request_id", reqID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": payload.Status})
}

// LeaveGroup handles DELETE /groups/{id}/members/me.
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "groups.leave", logger.F("group_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "groups.leave")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	if err := h.service.LeaveGroup(r.Context(), groupID, userID); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.leave", err, logger.F("group_id", groupID), logger.F("user_id", userID))
		} else {
			logBadRequest(h.log, "groups.leave", logger.F("group_id", groupID), logger.F("user_id", userID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "left"})
}

func mapGroupError(err error) (int, string) {
	switch {
	case errors.Is(err, usecasegroup.ErrInvalidTitle):
		return http.StatusBadRequest, utils.MsgInvalidGroupTitle
	case errors.Is(err, usecasegroup.ErrCannotInviteSelf):
		return http.StatusBadRequest, utils.MsgCannotInviteSelf
	case errors.Is(err, usecasegroup.ErrAlreadyMember):
		return http.StatusConflict, utils.MsgAlreadyGroupMember
	case errors.Is(err, usecasegroup.ErrInvitationExists):
		return http.StatusConflict, utils.MsgGroupInvitationExists
	case errors.Is(err, usecasegroup.ErrJoinRequestExists):
		return http.StatusConflict, utils.MsgGroupJoinRequestExists
	case errors.Is(err, usecasegroup.ErrInvalidStatus):
		return http.StatusBadRequest, utils.MsgInvalidStatus
	case errors.Is(err, usecasegroup.ErrForbidden):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, usecasegroup.ErrCannotLeaveCreator):
		return http.StatusConflict, utils.MsgCannotLeaveGroup
	case errors.Is(err, domaingroup.ErrGroupNotFound):
		return http.StatusNotFound, utils.MsgGroupNotFound
	case errors.Is(err, domaingroup.ErrInvitationNotFound):
		return http.StatusNotFound, utils.MsgInvitationNotFound
	case errors.Is(err, domaingroup.ErrJoinRequestNotFound):
		return http.StatusNotFound, utils.MsgJoinRequestNotFound
	case errors.Is(err, domaingroup.ErrNotMember):
		return http.StatusForbidden, utils.MsgForbidden
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}
