package handler

import (
	"errors"
	"net/http"
	"strconv"

	domainpost "social-network/backend/internal/domain/post"
	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasepost "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/logger"
)

// PostHandler serves REST endpoints for posts.
type PostHandler struct {
	service *usecasepost.Service
	log     logger.Logger
}

// NewPostHandler builds a PostHandler.
func NewPostHandler(service *usecasepost.Service, log logger.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		log:     log.WithFields(logger.F("handler", "post")),
	}
}

// List handles GET /posts.
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	viewerID, _ := middleware.GetUserID(r.Context())

	if rawCategory := r.URL.Query().Get("category_id"); rawCategory != "" {
		categoryID, err := strconv.ParseInt(rawCategory, 10, 64)
		if err != nil || categoryID <= 0 {
			logBadRequest(h.log, "posts.list", logger.F("category_id", rawCategory))
			utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidCategoryID)
			return
		}
		posts, err := h.service.ListByCategory(r.Context(), categoryID, viewerID, limit, offset)
		if err != nil {
			logServerError(h.log, "posts.list", err, logger.F("category_id", categoryID))
			utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, posts)
		return
	}

	if rawAuthor := r.URL.Query().Get("author_id"); rawAuthor != "" {
		authorID, err := strconv.ParseInt(rawAuthor, 10, 64)
		if err != nil || authorID <= 0 {
			logBadRequest(h.log, "posts.list", logger.F("author_id", rawAuthor))
			utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidAuthorID)
			return
		}
		h.ListByAuthor(w, r, authorID, limit, offset)
		return
	}

	posts, err := h.service.List(r.Context(), viewerID, limit, offset)
	if err != nil {
		logServerError(h.log, "posts.list", err)
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}

// ListByGroup handles GET /groups/{id}/posts.
func (h *PostHandler) ListByGroup(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "posts.list_by_group", logger.F("group_id", groupIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	limit, offset, err := utils.ParsePagination(r)
	if err != nil {
		logBadRequest(h.log, "posts.list_by_group", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	viewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "posts.list_by_group")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	posts, err := h.service.ListByGroup(r.Context(), groupID, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			logForbidden(h.log, "posts.list_by_group", logger.F("group_id", groupID), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		logServerError(h.log, "posts.list_by_group", err, logger.F("group_id", groupID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}

// CreateInGroup handles POST /groups/{id}/posts.
func (h *PostHandler) CreateInGroup(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil || groupID <= 0 {
		logBadRequest(h.log, "posts.create_in_group", logger.F("group_id", groupIDStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidGroupID)
		return
	}

	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "posts.create_in_group")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecasepost.CreatePostRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "posts.create_in_group", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}
	req.GroupID = &groupID

	post, err := h.service.Create(r.Context(), authorID, req)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			logForbidden(h.log, "posts.create_in_group", logger.F("group_id", groupID), logger.F("author_id", authorID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		logBadRequest(h.log, "posts.create_in_group", logger.F("author_id", authorID), logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, post)
}

// Create handles POST /posts.
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	authorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "posts.create")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	var req usecasepost.CreatePostRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		logBadRequest(h.log, "posts.create", logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	post, err := h.service.Create(r.Context(), authorID, req)
	if err != nil {
		logBadRequest(h.log, "posts.create", logger.F("author_id", authorID), logger.F("error", err.Error()))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidRequestBody)
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, post)
}

// GetByID handles GET /posts/{id}.
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		logBadRequest(h.log, "posts.get", logger.F("post_id", idStr))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidPostID)
		return
	}

	viewerID, _ := middleware.GetUserID(r.Context())
	post, err := h.service.GetByID(r.Context(), id, viewerID)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			logForbidden(h.log, "posts.get", logger.F("post_id", id), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		if errors.Is(err, domainpost.ErrNotFound) {
			logNotFound(h.log, "posts.get", logger.F("post_id", id))
			utils.RespondWithError(w, http.StatusNotFound, utils.MsgPostNotFound)
			return
		}
		logServerError(h.log, "posts.get", err, logger.F("post_id", id))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, post)
}

// ListByAuthor handles GET /posts?author_id=...&limit=&offset=
func (h *PostHandler) ListByAuthor(w http.ResponseWriter, r *http.Request, authorID int64, limit, offset int) {
	viewerID, _ := middleware.GetUserID(r.Context())
	posts, err := h.service.ListByAuthor(r.Context(), authorID, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, usecasepost.ErrForbidden) {
			logForbidden(h.log, "posts.list_by_author", logger.F("author_id", authorID), logger.F("viewer_id", viewerID))
			utils.RespondWithError(w, http.StatusForbidden, utils.MsgForbidden)
			return
		}
		logServerError(h.log, "posts.list_by_author", err, logger.F("author_id", authorID))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, posts)
}
