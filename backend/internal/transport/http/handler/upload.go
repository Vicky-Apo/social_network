package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	httputils "social-network/backend/internal/transport/http/utils"
	"social-network/backend/pkg/logger"
	pkgutils "social-network/backend/pkg/utils"
)

// UploadHandler serves file uploads.
type UploadHandler struct {
	uploadDir string
	maxBytes  int64
	log       logger.Logger
}

// NewUploadHandler builds an UploadHandler.
func NewUploadHandler(uploadDir string, maxBytes int64, log logger.Logger) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
		maxBytes:  maxBytes,
		log:       log.WithFields(logger.F("handler", "upload")),
	}
}

// Upload handles POST /uploads.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	}

	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		logBadRequest(h.log, "uploads.create", logger.F("error", err.Error()))
		httputils.RespondWithError(w, http.StatusBadRequest, httputils.MsgInvalidUpload)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logBadRequest(h.log, "uploads.create", logger.F("error", err.Error()))
		httputils.RespondWithError(w, http.StatusBadRequest, httputils.MsgInvalidUpload)
		return
	}
	defer file.Close()

	kind := strings.TrimSpace(r.FormValue("kind"))
	if kind == "" {
		kind = "media"
	}
	if !isAllowedKind(kind) {
		httputils.RespondWithError(w, http.StatusBadRequest, httputils.MsgInvalidUploadKind)
		return
	}

	contentType, sniffed, err := sniffContentType(file)
	if err != nil {
		logBadRequest(h.log, "uploads.create", logger.F("error", err.Error()))
		httputils.RespondWithError(w, http.StatusBadRequest, httputils.MsgInvalidUpload)
		return
	}
	ext, ok := extensionForContentType(contentType)
	if !ok {
		logBadRequest(h.log, "uploads.create", logger.F("content_type", contentType))
		httputils.RespondWithError(w, http.StatusBadRequest, httputils.MsgInvalidUploadType)
		return
	}

	name, err := safeFilename(ext)
	if err != nil {
		logServerError(h.log, "uploads.create", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, httputils.MsgInternalServerError)
		return
	}

	dir := filepath.Join(h.uploadDir, kind)
	if err := pkgutils.EnsureDir(dir); err != nil {
		logServerError(h.log, "uploads.create", err, logger.F("dir", dir))
		httputils.RespondWithError(w, http.StatusInternalServerError, httputils.MsgInternalServerError)
		return
	}

	fullPath := filepath.Join(dir, name)
	out, err := pkgutils.SafeCreateFile(fullPath)
	if err != nil {
		logServerError(h.log, "uploads.create", err, logger.F("path", fullPath))
		httputils.RespondWithError(w, http.StatusInternalServerError, httputils.MsgInternalServerError)
		return
	}
	defer out.Close()

	if _, err := out.Write(sniffed); err != nil {
		logServerError(h.log, "uploads.create", err, logger.F("path", fullPath))
		httputils.RespondWithError(w, http.StatusInternalServerError, httputils.MsgInternalServerError)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		logServerError(h.log, "uploads.create", err, logger.F("path", fullPath))
		httputils.RespondWithError(w, http.StatusInternalServerError, httputils.MsgInternalServerError)
		return
	}

	_ = header // kept for future use (size/name validation)

	publicPath := "/uploads/" + kind + "/" + name
	httputils.RespondWithSuccess(w, http.StatusCreated, map[string]any{
		"path":         publicPath,
		"content_type": contentType,
	})
}

func sniffContentType(file io.Reader) (string, []byte, error) {
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", nil, err
	}
	buf = buf[:n]
	return http.DetectContentType(buf), buf, nil
}

func extensionForContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}

func isAllowedKind(kind string) bool {
	switch kind {
	case "media", "avatar", "post", "comment", "message", "group":
		return true
	default:
		return false
	}
}

func safeFilename(ext string) (string, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	rnd := hex.EncodeToString(randBytes)
	stamp := time.Now().UTC().Format("20060102T150405")
	return stamp + "_" + rnd + ext, nil
}
