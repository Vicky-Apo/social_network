package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	domaincomment "social-network/backend/internal/domain/comment"
	domaingroup "social-network/backend/internal/domain/group"
	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasecomment "social-network/backend/internal/usecase/comment"
	usecasegroup "social-network/backend/internal/usecase/group"
	usecasepost "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/logger"
)

// GroupHandler serves REST endpoints for groups.
type GroupHandler struct {
	service        *usecasegroup.Service
	postService    *usecasepost.Service
	commentService *usecasecomment.Service
	log            logger.Logger
}

// NewGroupHandler builds a GroupHandler.
func NewGroupHandler(service *usecasegroup.Service, postService *usecasepost.Service, commentService *usecasecomment.Service, log logger.Logger) *GroupHandler {
	return &GroupHandler{
		service:        service,
		postService:    postService,
		commentService: commentService,
		log:            log.WithFields(logger.F("handler", "group")),
	}
}

// Create handles POST /groups.
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var req usecasegroup.CreateGroupRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	group, err := h.service.Create(r.Context(), creatorID, req)
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
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	groups, err := h.service.List(r.Context(), query, limit, offset)
	if err != nil {
		logServerError(h.log, "groups.list", err)
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, groups)
}

// Get handles GET /groups/{id}.
func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	group, err := h.service.Get(r.Context(), groupID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.get", err, logger.F("group_id", groupID))
		} else {
			logNotFound(h.log, "groups.get", logger.F("group_id", groupID))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, group)
}

// Update handles PATCH /groups/{id}.
func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var req usecasegroup.UpdateGroupRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	group, err := h.service.Update(r.Context(), groupID, actorID, req)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.update", err, logger.F("group_id", groupID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.update", logger.F("group_id", groupID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, group)
}

// Delete handles DELETE /groups/{id}.
func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	if err := h.service.Delete(r.Context(), groupID, actorID); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.delete", err, logger.F("group_id", groupID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.delete", logger.F("group_id", groupID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListMembers handles GET /groups/{id}/members.
func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	members, err := h.service.ListMembers(r.Context(), groupID, actorID, limit, offset)
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

// Leave handles DELETE /groups/{id}/members/me.
func (h *GroupHandler) Leave(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	if err := h.service.Leave(r.Context(), groupID, actorID); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.leave", err, logger.F("group_id", groupID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.leave", logger.F("group_id", groupID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "left"})
}

// RemoveMember handles DELETE /groups/{id}/members/{user_id}.
func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	memberStr := r.PathValue("user_id")
	memberID, err := strconv.ParseInt(memberStr, 10, 64)
	if err != nil || memberID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	if err := h.service.RemoveMember(r.Context(), groupID, actorID, memberID); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.remove_member", err, logger.F("group_id", groupID), logger.F("actor_id", actorID))
		} else {
			logBadRequest(h.log, "groups.remove_member", logger.F("group_id", groupID), logger.F("actor_id", actorID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "removed"})
}

// Invite handles POST /groups/{id}/invitations.
func (h *GroupHandler) Invite(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	inviterID, ok := middleware.GetUserID(r.Context())
	if !ok {
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
	if payload.InviteeID == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUserID)
		return
	}
	inv, err := h.service.Invite(r.Context(), groupID, inviterID, payload.InviteeID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.invite", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.invite", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]any{"invitation": inv})
}

// ListInvitations handles GET /group-invitations.
func (h *GroupHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	invs, err := h.service.ListInvitations(r.Context(), actorID, limit, offset)
	if err != nil {
		logServerError(h.log, "groups.invitations_list", err, logger.F("actor_id", actorID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, invs)
}

// RespondInvitation handles PATCH /group-invitations/{id}.
func (h *GroupHandler) RespondInvitation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	invitationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || invitationID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidInvitationID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.invitation_respond", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if err := h.service.RespondInvitation(r.Context(), invitationID, actorID, payload.Status); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.invitation_respond", err, logger.F("invitation_id", invitationID))
		} else {
			logBadRequest(h.log, "groups.invitation_respond", logger.F("invitation_id", invitationID), logger.F("reason", message))
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
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	req, err := h.service.RequestJoin(r.Context(), groupID, actorID)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join_request", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.join_request", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]any{"request": req})
}

// ListJoinRequests handles GET /groups/{id}/join-requests.
func (h *GroupHandler) ListJoinRequests(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	reqs, err := h.service.ListJoinRequests(r.Context(), groupID, actorID, limit, offset)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join_requests", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.join_requests", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, reqs)
}

// RespondJoinRequest handles PATCH /groups/{id}/join-requests/{request_id}.
func (h *GroupHandler) RespondJoinRequest(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	reqStr := r.PathValue("request_id")
	requestID, err := strconv.ParseInt(reqStr, 10, 64)
	if err != nil || requestID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidJoinRequestID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.join_request_respond", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if err := h.service.RespondJoinRequest(r.Context(), groupID, requestID, actorID, payload.Status); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.join_request_respond", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.join_request_respond", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": payload.Status})
}

// CreateGroupPost handles POST /groups/{id}/posts.
func (h *GroupHandler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var req usecasepost.CreatePostRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.posts.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	post, err := h.postService.CreateGroupPost(r.Context(), authorID, groupID, req)
	if err != nil {
		status, message := mapPostError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.posts.create", err, logger.F("group_id", groupID), logger.F("author_id", authorID))
		} else {
			logBadRequest(h.log, "groups.posts.create", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusCreated, post)
}

// ListGroupPosts handles GET /groups/{id}/posts.
func (h *GroupHandler) ListGroupPosts(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}
	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	posts, err := h.postService.ListByGroup(r.Context(), groupID, viewerID, limit, offset)
	if err != nil {
		status, message := mapPostError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.posts.list", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.posts.list", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, posts)
}

// CreateGroupComment handles POST /groups/{id}/posts/{post_id}/comments.
func (h *GroupHandler) CreateGroupComment(w http.ResponseWriter, r *http.Request) {
	groupID, postID, ok := parseGroupPostIDs(w, r)
	if !ok {
		return
	}
	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	post, err := h.postService.GetByID(r.Context(), postID, authorID)
	if err != nil {
		status, message := mapPostError(err)
		utils.RespondWithError(w, status, message)
		return
	}
	if post.GroupID == nil || *post.GroupID != groupID {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}
	var req usecasecomment.CreateCommentRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.comments.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	req.PostID = postID
	req.AuthorID = authorID
	comment, err := h.commentService.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, usecasecomment.ErrInvalidRequest) {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		logServerError(h.log, "groups.comments.create", err, logger.F("post_id", postID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusCreated, comment)
}

// ListGroupComments handles GET /groups/{id}/posts/{post_id}/comments.
func (h *GroupHandler) ListGroupComments(w http.ResponseWriter, r *http.Request) {
	groupID, postID, ok := parseGroupPostIDs(w, r)
	if !ok {
		return
	}
	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	post, err := h.postService.GetByID(r.Context(), postID, viewerID)
	if err != nil {
		status, message := mapPostError(err)
		utils.RespondWithError(w, status, message)
		return
	}
	if post.GroupID == nil || *post.GroupID != groupID {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	comments, err := h.commentService.GetByPostID(r.Context(), postID, limit, offset)
	if err != nil {
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentsNotFound)
			return
		}
		logServerError(h.log, "groups.comments.list", err, logger.F("post_id", postID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, comments)
}

// UpdateGroupComment handles PATCH /group-comments/{id}.
func (h *GroupHandler) UpdateGroupComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	comment, err := h.commentService.GetByID(r.Context(), commentID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentNotFound)
			return
		}
		logServerError(h.log, "groups.comments.update", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	post, err := h.postService.GetByID(r.Context(), comment.PostID, actorID)
	if err != nil {
		status, message := mapPostError(err)
		utils.RespondWithError(w, status, message)
		return
	}
	if post.GroupID == nil {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}
	var req usecasecomment.UpdateCommentRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.comments.update", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	updated, err := h.commentService.Update(r.Context(), commentID, actorID, req)
	if err != nil {
		if errors.Is(err, usecasecomment.ErrInvalidRequest) {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, usecasecomment.ErrForbidden) {
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentNotFound)
			return
		}
		logServerError(h.log, "groups.comments.update", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, updated)
}

// DeleteGroupComment handles DELETE /group-comments/{id}.
func (h *GroupHandler) DeleteGroupComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCommentID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	comment, err := h.commentService.GetByID(r.Context(), commentID)
	if err != nil {
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentNotFound)
			return
		}
		logServerError(h.log, "groups.comments.delete", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	post, err := h.postService.GetByID(r.Context(), comment.PostID, actorID)
	if err != nil {
		status, message := mapPostError(err)
		utils.RespondWithError(w, status, message)
		return
	}
	if post.GroupID == nil {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}
	if err := h.commentService.Delete(r.Context(), commentID, actorID); err != nil {
		if errors.Is(err, usecasecomment.ErrForbidden) {
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domaincomment.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgCommentNotFound)
			return
		}
		logServerError(h.log, "groups.comments.delete", err, logger.F("comment_id", commentID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// CreateEvent handles POST /groups/{id}/events.
func (h *GroupHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseGroupID(w, r)
	if !ok {
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var req usecasegroup.CreateEventRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "groups.events.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	event, err := h.service.CreateEvent(r.Context(), groupID, actorID, req)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.events.create", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.events.create", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusCreated, event)
}

// ListEvents handles GET /groups/{id}/events.
func (h *GroupHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseGroupID(w, r)
	if !ok {
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, err := h.service.ListEvents(r.Context(), groupID, actorID, limit, offset)
	if err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.events.list", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.events.list", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, events)
}

// RSVPEvent handles POST /groups/{id}/events/{event_id}/rsvp.
func (h *GroupHandler) RSVPEvent(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseGroupID(w, r)
	if !ok {
		return
	}
	eventStr := r.PathValue("event_id")
	eventID, err := strconv.ParseInt(eventStr, 10, 64)
	if err != nil || eventID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidEventID)
		return
	}
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}
	var payload struct {
		Response string `json:"response"`
	}
	if err := utils.ReadJSON(r, &payload); err != nil {
		logBadRequest(h.log, "groups.events.rsvp", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	if err := h.service.RSVP(r.Context(), groupID, eventID, actorID, payload.Response); err != nil {
		status, message := mapGroupError(err)
		if status >= http.StatusInternalServerError {
			logServerError(h.log, "groups.events.rsvp", err, logger.F("group_id", groupID))
		} else {
			logBadRequest(h.log, "groups.events.rsvp", logger.F("group_id", groupID), logger.F("reason", message))
		}
		utils.RespondWithError(w, status, message)
		return
	}
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"response": payload.Response})
}

func parseGroupID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	idStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || groupID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return 0, false
	}
	return groupID, true
}

func parseGroupPostIDs(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	groupID, ok := parseGroupID(w, r)
	if !ok {
		return 0, 0, false
	}
	postStr := r.PathValue("post_id")
	postID, err := strconv.ParseInt(postStr, 10, 64)
	if err != nil || postID <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return 0, 0, false
	}
	return groupID, postID, true
}

func mapGroupError(err error) (int, string) {
	switch {
	case errors.Is(err, usecasegroup.ErrInvalidRequest), errors.Is(err, usecasegroup.ErrInvalidStatus), errors.Is(err, usecasegroup.ErrInvalidResponse):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, usecasegroup.ErrForbidden), errors.Is(err, domaingroup.ErrNotMember):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, domaingroup.ErrGroupNotFound):
		return http.StatusNotFound, utils.MsgGroupNotFound
	case errors.Is(err, domaingroup.ErrInvitationNotFound):
		return http.StatusNotFound, utils.MsgInvitationNotFound
	case errors.Is(err, domaingroup.ErrJoinRequestNotFound):
		return http.StatusNotFound, utils.MsgJoinRequestNotFound
	case errors.Is(err, domaingroup.ErrEventNotFound):
		return http.StatusNotFound, utils.MsgEventNotFound
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}

func mapPostError(err error) (int, string) {
	switch {
	case errors.Is(err, usecasepost.ErrForbidden):
		return http.StatusForbidden, utils.MsgForbidden
	case errors.Is(err, domainpost.ErrNotFound):
		return http.StatusNotFound, utils.MsgPostNotFound
	case errors.Is(err, usecasepost.ErrInvalidRequest):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, utils.MsgInternalServerError
	}
}
