package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainnotification "social-network/backend/internal/domain/notification"
	usecasenotification "social-network/backend/internal/usecase/notification"
	"social-network/backend/pkg/logger"
)

func TestNotificationList_Unauthorized(t *testing.T) {
	h := NewNotificationHandler(nil, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

type fakeNotificationRepo struct{}

func (r *fakeNotificationRepo) Create(ctx context.Context, n domainnotification.Notification) (domainnotification.Notification, error) {
	return n, nil
}
func (r *fakeNotificationRepo) ListByUser(ctx context.Context, userID int64, limit, offset int, unreadOnly bool) ([]domainnotification.Notification, error) {
	return []domainnotification.Notification{{ID: 1, UserID: userID, Type: "follow_request", EntityType: "follow_request", EntityID: 1}}, nil
}
func (r *fakeNotificationRepo) MarkRead(ctx context.Context, userID, notificationID int64, readAt time.Time) error {
	return nil
}
func (r *fakeNotificationRepo) MarkAllRead(ctx context.Context, userID int64, readAt time.Time) (int64, error) {
	return 0, nil
}
func (r *fakeNotificationRepo) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}

func TestNotificationList_Success(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := usecasenotification.NewService(repo, nil, logger.NewDefault(false))
	h := NewNotificationHandler(svc, logger.NewDefault(false))

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "token"})
	rr := httptest.NewRecorder()

	handler := authWrap(http.HandlerFunc(h.List), 1)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
