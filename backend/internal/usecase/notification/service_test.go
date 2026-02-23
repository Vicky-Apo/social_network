package notification

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domainnotification "social-network/backend/internal/domain/notification"
)

type fakeNotificationRepo struct {
	created         domainnotification.Notification
	createErr       error
	list            []domainnotification.Notification
	listErr         error
	markReadCalls   int
	markAllCalls    int
	unreadCalls     int
	markReadErr     error
	markAllErr      error
	unreadErr       error
	markAllReturn   int64
	unreadReturn    int64
	lastMarkReadAt  time.Time
	lastMarkAllRead time.Time
	lastReadUserID  int64
	lastReadID      int64
	lastUserID      int64
	lastLimit       int
	lastOffset      int
	lastUnreadOnly  bool
}

func (r *fakeNotificationRepo) Create(ctx context.Context, n domainnotification.Notification) (domainnotification.Notification, error) {
	if r.createErr != nil {
		return domainnotification.Notification{}, r.createErr
	}
	n.ID = 1
	n.CreatedAt = time.Now()
	r.created = n
	return n, nil
}

func (r *fakeNotificationRepo) ListByUser(ctx context.Context, userID int64, limit, offset int, unreadOnly bool) ([]domainnotification.Notification, error) {
	r.lastUserID = userID
	r.lastLimit = limit
	r.lastOffset = offset
	r.lastUnreadOnly = unreadOnly
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.list, nil
}

func (r *fakeNotificationRepo) MarkRead(ctx context.Context, userID, notificationID int64, readAt time.Time) error {
	r.markReadCalls++
	r.lastReadUserID = userID
	r.lastReadID = notificationID
	r.lastMarkReadAt = readAt
	return r.markReadErr
}

func (r *fakeNotificationRepo) MarkAllRead(ctx context.Context, userID int64, readAt time.Time) (int64, error) {
	r.markAllCalls++
	r.lastReadUserID = userID
	r.lastMarkAllRead = readAt
	if r.markAllErr != nil {
		return 0, r.markAllErr
	}
	return r.markAllReturn, nil
}

func (r *fakeNotificationRepo) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	r.unreadCalls++
	r.lastUserID = userID
	if r.unreadErr != nil {
		return 0, r.unreadErr
	}
	return r.unreadReturn, nil
}

type fakePublisher struct {
	calls   int
	lastID  int64
	lastDTO NotificationDTO
	err     error
}

func (p *fakePublisher) Publish(ctx context.Context, userID int64, payload NotificationDTO) error {
	p.calls++
	p.lastID = userID
	p.lastDTO = payload
	return p.err
}

func TestCreateForUser_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  CreateRequest
	}{
		{
			name: "invalid_user",
			req:  CreateRequest{UserID: 0, Type: string(domainnotification.FollowRequest), EntityType: "user", EntityID: 1},
		},
		{
			name: "invalid_entity",
			req:  CreateRequest{UserID: 1, Type: string(domainnotification.FollowRequest), EntityType: "user", EntityID: 0},
		},
		{
			name: "missing_type",
			req:  CreateRequest{UserID: 1, EntityType: "user", EntityID: 1},
		},
		{
			name: "missing_entity_type",
			req:  CreateRequest{UserID: 1, Type: string(domainnotification.FollowRequest), EntityID: 1},
		},
		{
			name: "invalid_type",
			req:  CreateRequest{UserID: 1, Type: "unknown", EntityType: "user", EntityID: 1},
		},
	}

	svc := NewService(&fakeNotificationRepo{}, nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.CreateForUser(context.Background(), tt.req); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestCreateForUser_MetadataMarshalError(t *testing.T) {
	svc := NewService(&fakeNotificationRepo{}, nil, nil)
	_, err := svc.CreateForUser(context.Background(), CreateRequest{
		UserID:     1,
		Type:       string(domainnotification.FollowRequest),
		EntityType: "user",
		EntityID:   1,
		Metadata:   map[string]any{"bad": func() {}},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCreateForUser_SuccessPublishes(t *testing.T) {
	repo := &fakeNotificationRepo{}
	pub := &fakePublisher{}
	svc := NewService(repo, pub, nil)

	actorID := int64(2)
	meta := map[string]any{"group_id": float64(3)}
	dto, err := svc.CreateForUser(context.Background(), CreateRequest{
		UserID:     1,
		ActorID:    &actorID,
		Type:       string(domainnotification.GroupInvitation),
		EntityType: "group",
		EntityID:   3,
		Metadata:   meta,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created.UserID != 1 || repo.created.EntityID != 3 {
		t.Fatalf("unexpected create payload")
	}
	if len(repo.created.Metadata) == 0 {
		t.Fatalf("expected metadata to be stored")
	}
	if pub.calls != 1 || pub.lastID != 1 {
		t.Fatalf("expected publish call")
	}
	if dto.Type != string(domainnotification.GroupInvitation) || dto.EntityType != "group" || dto.EntityID != 3 {
		t.Fatalf("unexpected dto")
	}
	if dto.Metadata == nil || dto.Metadata["group_id"] != meta["group_id"] {
		t.Fatalf("expected metadata in dto")
	}
}

func TestCreateForUser_PublishErrorDoesNotFail(t *testing.T) {
	repo := &fakeNotificationRepo{}
	pub := &fakePublisher{err: errors.New("fail")}
	svc := NewService(repo, pub, nil)

	_, err := svc.CreateForUser(context.Background(), CreateRequest{
		UserID:     1,
		Type:       string(domainnotification.FollowRequest),
		EntityType: "user",
		EntityID:   2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList(t *testing.T) {
	repo := &fakeNotificationRepo{
		list: []domainnotification.Notification{
			{ID: 1, UserID: 1, Type: domainnotification.FollowRequest, EntityType: "user", EntityID: 2},
		},
	}
	svc := NewService(repo, nil, nil)

	items, err := svc.List(context.Background(), 1, 10, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("unexpected items")
	}
	if repo.lastUserID != 1 || repo.lastLimit != 10 || !repo.lastUnreadOnly {
		t.Fatalf("unexpected repo call")
	}

	if _, err := svc.List(context.Background(), 0, 10, 0, false); err == nil {
		t.Fatalf("expected error for invalid user id")
	}
}

func TestMarkRead(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := NewService(repo, nil, nil)

	if err := svc.MarkRead(context.Background(), 1, 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.markReadCalls != 1 || repo.lastReadUserID != 1 || repo.lastReadID != 2 {
		t.Fatalf("unexpected mark read call")
	}
	if repo.lastMarkReadAt.IsZero() {
		t.Fatalf("expected read time")
	}

	if err := svc.MarkRead(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected error for invalid ids")
	}
}

func TestMarkAllRead(t *testing.T) {
	repo := &fakeNotificationRepo{markAllReturn: 3}
	svc := NewService(repo, nil, nil)

	count, err := svc.MarkAllRead(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 || repo.markAllCalls != 1 || repo.lastReadUserID != 1 {
		t.Fatalf("unexpected mark all call")
	}
	if repo.lastMarkAllRead.IsZero() {
		t.Fatalf("expected read time")
	}

	if _, err := svc.MarkAllRead(context.Background(), 0); err == nil {
		t.Fatalf("expected error for invalid user id")
	}
}

func TestUnreadCount(t *testing.T) {
	repo := &fakeNotificationRepo{unreadReturn: 5}
	svc := NewService(repo, nil, nil)

	count, err := svc.UnreadCount(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 || repo.unreadCalls != 1 || repo.lastUserID != 1 {
		t.Fatalf("unexpected unread call")
	}

	if _, err := svc.UnreadCount(context.Background(), 0); err == nil {
		t.Fatalf("expected error for invalid user id")
	}
}

func TestMapNotification(t *testing.T) {
	meta, _ := json.Marshal(map[string]any{"k": "v"})
	n := domainnotification.Notification{
		ID:         1,
		UserID:     2,
		Type:       domainnotification.FollowRequest,
		EntityType: "user",
		EntityID:   3,
		Metadata:   meta,
	}
	dto := mapNotification(n)
	if dto.Metadata == nil || dto.Metadata["k"] != "v" {
		t.Fatalf("expected metadata to be decoded")
	}
}
