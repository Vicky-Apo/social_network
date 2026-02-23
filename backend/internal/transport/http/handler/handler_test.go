package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/backend/internal/transport/http/middleware"
	"social-network/backend/internal/transport/http/utils"
	"social-network/backend/pkg/logger"
)

const testCookieName = "session"

type fakeSessionValidator struct {
	userID int64
	err    error
}

func (f fakeSessionValidator) ValidateSession(ctx context.Context, token string) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.userID, nil
}

func authWrap(handler http.Handler, userID int64) http.Handler {
	validator := fakeSessionValidator{userID: userID}
	return middleware.Auth(validator, testCookieName, logger.NewDefault(false))(handler)
}

func newJSONRequest(t *testing.T, method, path string, payload any) *http.Request {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeAPIResponse(t *testing.T, rr *httptest.ResponseRecorder) utils.APIResponse {
	t.Helper()
	var resp utils.APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}
