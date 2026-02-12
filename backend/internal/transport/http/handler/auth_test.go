package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"social-network/backend/internal/config"
	domainauth "social-network/backend/internal/domain/auth"
	usecaseauth "social-network/backend/internal/usecase/auth"
	"social-network/backend/pkg/logger"
)

type fakeAuthRepo struct {
	userByEmail map[string]domainauth.User
	userByID    map[int64]domainauth.User
	nextID      int64
	sessions    map[string]domainauth.Session
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		userByEmail: make(map[string]domainauth.User),
		userByID:    make(map[int64]domainauth.User),
		nextID:      1,
		sessions:    make(map[string]domainauth.Session),
	}
}

func (r *fakeAuthRepo) CreateUser(ctx context.Context, user domainauth.User) (int64, error) {
	if _, exists := r.userByEmail[user.Email]; exists {
		return 0, domainauth.ErrEmailAlreadyExists
	}
	user.ID = r.nextID
	r.nextID++
	r.userByEmail[user.Email] = user
	r.userByID[user.ID] = user
	return user.ID, nil
}

func (r *fakeAuthRepo) GetUserByEmail(ctx context.Context, email string) (domainauth.User, error) {
	u, ok := r.userByEmail[email]
	if !ok {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeAuthRepo) GetUserByID(ctx context.Context, id int64) (domainauth.User, error) {
	u, ok := r.userByID[id]
	if !ok {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeAuthRepo) CreateSession(ctx context.Context, session domainauth.Session) (int64, error) {
	session.ID = int64(len(r.sessions) + 1)
	r.sessions[session.SessionToken] = session
	return session.ID, nil
}

func (r *fakeAuthRepo) GetSessionByToken(ctx context.Context, token string) (domainauth.Session, error) {
	s, ok := r.sessions[token]
	if !ok || time.Now().After(s.ExpiresAt) {
		return domainauth.Session{}, domainauth.ErrSessionNotFound
	}
	return s, nil
}

func (r *fakeAuthRepo) DeleteSession(ctx context.Context, token string) error {
	delete(r.sessions, token)
	return nil
}

func (r *fakeAuthRepo) DeleteUserSessions(ctx context.Context, userID int64) error {
	for k, v := range r.sessions {
		if v.UserID == userID {
			delete(r.sessions, k)
		}
	}
	return nil
}

func authTestConfig() config.AuthConfig {
	return config.AuthConfig{
		BcryptCost:          bcrypt.MinCost,
		SessionTokenBytes:   16,
		SessionDurationDays: 1,
		SessionCookieName:   testCookieName,
		SessionMaxAge:       86400,
		MinPasswordLength:   6,
		MinUserAge:          13,
	}
}

func TestAuthRegister_InvalidJSON(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := usecaseauth.NewService(repo, authTestConfig(), logger.NewDefault(false))
	h := NewAuthHandler(svc, logger.NewDefault(false), AuthHandlerConfig{CookieName: testCookieName, MaxAge: 3600})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", http.NoBody)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAuthLogin_SetsCookie(t *testing.T) {
	repo := newFakeAuthRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	repo.userByEmail["user@example.com"] = domainauth.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: string(hash),
		FirstName:    "U",
		LastName:     "One",
		DateOfBirth:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	repo.userByID[1] = repo.userByEmail["user@example.com"]

	svc := usecaseauth.NewService(repo, authTestConfig(), logger.NewDefault(false))
	h := NewAuthHandler(svc, logger.NewDefault(false), AuthHandlerConfig{CookieName: testCookieName, MaxAge: 3600})

	req := newJSONRequest(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    "user@example.com",
		"password": "password",
	})
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != testCookieName {
		t.Fatalf("expected session cookie")
	}
}

func TestAuthMe_Unauthorized(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := usecaseauth.NewService(repo, authTestConfig(), logger.NewDefault(false))
	h := NewAuthHandler(svc, logger.NewDefault(false), AuthHandlerConfig{CookieName: testCookieName, MaxAge: 3600})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", http.NoBody)
	rr := httptest.NewRecorder()
	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthRegister_Success(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := usecaseauth.NewService(repo, authTestConfig(), logger.NewDefault(false))
	h := NewAuthHandler(svc, logger.NewDefault(false), AuthHandlerConfig{CookieName: testCookieName, MaxAge: 3600})

	req := newJSONRequest(t, http.MethodPost, "/auth/register", map[string]string{
		"email":         "new@example.com",
		"password":      "strongpass",
		"first_name":    "New",
		"last_name":     "User",
		"date_of_birth": "01/01/2000",
	})
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}
