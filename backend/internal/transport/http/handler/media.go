package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	usecasemedia "social-network/backend/internal/usecase/media"
	"social-network/backend/pkg/logger"
)

// MediaHandler serves authenticated media files with access checks.
type MediaHandler struct {
	service   *usecasemedia.Service
	uploadDir string
	log       logger.Logger
}

// NewMediaHandler builds a MediaHandler.
func NewMediaHandler(service *usecasemedia.Service, uploadDir string, log logger.Logger) *MediaHandler {
	return &MediaHandler{
		service:   service,
		uploadDir: uploadDir,
		log:       log.WithFields(logger.F("handler", "media")),
	}
}

// Serve handles GET /uploads/{path...}
func (h *MediaHandler) Serve(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.PathValue("path"))
	if path == "" {
		utils.RespondWithError(w, http.StatusNotFound, utils.MsgNotFound)
		return
	}

	// Reject any traversal attempts early.
	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		logBadRequest(h.log, "uploads.get", logger.F("path", path))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUpload)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logUnauthorized(h.log, "uploads.get")
		utils.RespondWithError(w, http.StatusUnauthorized, utils.MsgUnauthorized)
		return
	}

	publicPath := "/uploads/" + path
	allowed, err := h.service.CanAccess(r.Context(), userID, publicPath)
	if err != nil {
		logServerError(h.log, "uploads.get", err, logger.F("user_id", userID), logger.F("path", publicPath))
		utils.RespondWithError(w, http.StatusInternalServerError, utils.MsgInternalServerError)
		return
	}
	if !allowed {
		// Avoid leaking whether a file exists.
		logForbidden(h.log, "uploads.get", logger.F("user_id", userID), logger.F("path", publicPath))
		utils.RespondWithError(w, http.StatusNotFound, utils.MsgNotFound)
		return
	}

	fullPath := filepath.Join(h.uploadDir, filepath.FromSlash(path))
	cleanBase := filepath.Clean(h.uploadDir) + string(filepath.Separator)
	cleanFull := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanFull, cleanBase) {
		logBadRequest(h.log, "uploads.get", logger.F("path", path))
		utils.RespondWithError(w, http.StatusBadRequest, utils.MsgInvalidUpload)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	http.ServeFile(w, r, fullPath)
}
