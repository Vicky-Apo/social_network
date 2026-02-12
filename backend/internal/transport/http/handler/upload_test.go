package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"social-network/backend/internal/transport/http/utils"
	"social-network/backend/pkg/logger"
)

func TestUpload_AcceptsImagesAndSaves(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		payload    []byte
		wantType   string
		wantExt    string
		wantPrefix []byte
	}{
		{
			name:       "png",
			filename:   "image.png",
			payload:    append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0x00}, 10)...),
			wantType:   "image/png",
			wantExt:    ".png",
			wantPrefix: []byte{0x89, 0x50, 0x4E, 0x47},
		},
		{
			name:       "jpeg",
			filename:   "image.jpg",
			payload:    append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}, bytes.Repeat([]byte{0x00}, 10)...),
			wantType:   "image/jpeg",
			wantExt:    ".jpg",
			wantPrefix: []byte{0xFF, 0xD8, 0xFF},
		},
		{
			name:       "gif",
			filename:   "image.gif",
			payload:    append([]byte("GIF89a"), bytes.Repeat([]byte{0x00}, 10)...),
			wantType:   "image/gif",
			wantExt:    ".gif",
			wantPrefix: []byte("GIF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			h := NewUploadHandler(tempDir, 1024*1024, logger.NewDefault(false))

			body, contentType := buildMultipart(t, "post", tt.filename, tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/uploads", body)
			req.Header.Set("Content-Type", contentType)

			rr := httptest.NewRecorder()
			h.Upload(rr, req)

			if rr.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tt.wantType) {
				t.Fatalf("expected content type %s in response", tt.wantType)
			}

			files, err := os.ReadDir(filepath.Join(tempDir, "post"))
			if err != nil || len(files) != 1 {
				t.Fatalf("expected uploaded file in directory")
			}
			if !strings.HasSuffix(files[0].Name(), tt.wantExt) {
				t.Fatalf("expected file extension %s, got %s", tt.wantExt, files[0].Name())
			}

			data, err := os.ReadFile(filepath.Join(tempDir, "post", files[0].Name()))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			if !bytes.HasPrefix(data, tt.wantPrefix) {
				t.Fatalf("expected file content to match uploaded bytes")
			}
		})
	}
}

func TestUpload_RejectsInvalidType(t *testing.T) {
	tempDir := t.TempDir()
	h := NewUploadHandler(tempDir, 1024*1024, logger.NewDefault(false))

	body, contentType := buildMultipart(t, "", "file.txt", []byte("not an image"))

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(utils.MsgInvalidUploadType)) {
		t.Fatalf("expected invalid upload type message")
	}
}

func TestUpload_RejectsInvalidKind(t *testing.T) {
	tempDir := t.TempDir()
	h := NewUploadHandler(tempDir, 1024*1024, logger.NewDefault(false))

	body, contentType := buildMultipart(t, "not-valid", "image.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(utils.MsgInvalidUploadKind)) {
		t.Fatalf("expected invalid upload kind message")
	}
}

func TestUpload_RejectsTooLarge(t *testing.T) {
	tempDir := t.TempDir()
	h := NewUploadHandler(tempDir, 16, logger.NewDefault(false))

	payload := append([]byte("GIF89a"), bytes.Repeat([]byte{0x00}, 100)...)
	body, contentType := buildMultipart(t, "post", "image.gif", payload)

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(utils.MsgInvalidUpload)) {
		t.Fatalf("expected invalid upload message")
	}
}

func TestUpload_MissingFileField(t *testing.T) {
	tempDir := t.TempDir()
	h := NewUploadHandler(tempDir, 1024*1024, logger.NewDefault(false))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("kind", "post"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(utils.MsgInvalidUpload)) {
		t.Fatalf("expected invalid upload message")
	}
}

func TestUpload_MissingMultipartBoundary(t *testing.T) {
	tempDir := t.TempDir()
	h := NewUploadHandler(tempDir, 1024*1024, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodPost, "/uploads", bytes.NewBufferString("no-boundary"))
	req.Header.Set("Content-Type", "multipart/form-data")

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(utils.MsgInvalidUpload)) {
		t.Fatalf("expected invalid upload message")
	}
}

func buildMultipart(t *testing.T, kind, filename string, payload []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if kind != "" {
		if err := writer.WriteField("kind", kind); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, writer.FormDataContentType()
}
