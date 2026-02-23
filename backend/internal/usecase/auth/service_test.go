package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"social-network/backend/internal/config"
	domainauth "social-network/backend/internal/domain/auth"
	"social-network/backend/pkg/logger"
)

type fakeAuthRepo struct {
	usersByEmail       map[string]domainauth.User
	usersByID          map[int64]domainauth.User
	sessionsByToken    map[string]domainauth.Session
	createUserErr      error
	createSessionErr   error
	deleteSessionErr   error
	lastCreatedUser    *domainauth.User
	lastCreatedSession *domainauth.Session
	nextUserID         int64
	nextSessionID      int64
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		usersByEmail:    make(map[string]domainauth.User),
		usersByID:       make(map[int64]domainauth.User),
		sessionsByToken: make(map[string]domainauth.Session),
		nextUserID:      1,
		nextSessionID:   1,
	}
}

func (r *fakeAuthRepo) CreateUser(ctx context.Context, user domainauth.User) (int64, error) {
	if r.createUserErr != nil {
		return 0, r.createUserErr
	}
	if _, exists := r.usersByEmail[user.Email]; exists {
		return 0, domainauth.ErrEmailAlreadyExists
	}
	id := r.nextUserID
	r.nextUserID++
	user.ID = id
	r.usersByEmail[user.Email] = user
	r.usersByID[id] = user
	r.lastCreatedUser = &user
	return id, nil
}

func (r *fakeAuthRepo) GetUserByEmail(ctx context.Context, email string) (domainauth.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeAuthRepo) GetUserByID(ctx context.Context, id int64) (domainauth.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeAuthRepo) CreateSession(ctx context.Context, session domainauth.Session) (int64, error) {
	if r.createSessionErr != nil {
		return 0, r.createSessionErr
	}
	id := r.nextSessionID
	r.nextSessionID++
	session.ID = id
	r.sessionsByToken[session.SessionToken] = session
	r.lastCreatedSession = &session
	return id, nil
}

func (r *fakeAuthRepo) GetSessionByToken(ctx context.Context, token string) (domainauth.Session, error) {
	sess, ok := r.sessionsByToken[token]
	if !ok {
		return domainauth.Session{}, domainauth.ErrSessionNotFound
	}
	return sess, nil
}

func (r *fakeAuthRepo) DeleteSession(ctx context.Context, token string) error {
	if r.deleteSessionErr != nil {
		return r.deleteSessionErr
	}
	delete(r.sessionsByToken, token)
	return nil
}

func (r *fakeAuthRepo) DeleteUserSessions(ctx context.Context, userID int64) error {
	for k, v := range r.sessionsByToken {
		if v.UserID == userID {
			delete(r.sessionsByToken, k)
		}
	}
	return nil
}

func TestRegister_Success(t *testing.T) {
	repo := newFakeAuthRepo()
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13}
	service := NewService(repo, cfg, logger.NewDefault(false))

	req := RegisterRequest{
		Email:       "  user@example.com ",
		Password:    "secret123",
		FirstName:   "  Jane ",
		LastName:    " Doe ",
		DateOfBirth: "01/01/2000",
		Nickname:    strPtr("jdoe"),
		About:       strPtr("hello"),
	}

	user, err := service.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("expected user id")
	}
	if repo.lastCreatedUser == nil {
		t.Fatalf("expected user to be created")
	}
	if repo.lastCreatedUser.Email != "user@example.com" {
		t.Fatalf("expected trimmed email, got %q", repo.lastCreatedUser.Email)
	}
	if repo.lastCreatedUser.FirstName != "Jane" || repo.lastCreatedUser.LastName != "Doe" {
		t.Fatalf("expected trimmed names")
	}
}

func TestRegister_EmailExists(t *testing.T) {
	repo := newFakeAuthRepo()
	repo.createUserErr = domainauth.ErrEmailAlreadyExists
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13}
	service := NewService(repo, cfg, logger.NewDefault(false))

	req := RegisterRequest{
		Email:       "user@example.com",
		Password:    "secret123",
		FirstName:   "Jane",
		LastName:    "Doe",
		DateOfBirth: "01/01/2000",
	}

	_, err := service.Register(context.Background(), req)
	if !errors.Is(err, domainauth.ErrEmailAlreadyExists) {
		t.Fatalf("expected email exists error, got %v", err)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	repo := newFakeAuthRepo()
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13, SessionTokenBytes: 16, SessionDurationDays: 1}
	service := NewService(repo, cfg, logger.NewDefault(false))

	_, err := service.Login(context.Background(), LoginRequest{Email: "missing@example.com", Password: "pw"}, "", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newFakeAuthRepo()
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13, SessionTokenBytes: 16, SessionDurationDays: 1}
	service := NewService(repo, cfg, logger.NewDefault(false))

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), 4)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user := domainauth.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash), FirstName: "Jane", LastName: "Doe", DateOfBirth: time.Now().AddDate(-20, 0, 0)}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.ID] = user

	resp, err := service.Login(context.Background(), LoginRequest{Email: user.Email, Password: "secret"}, "agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected session token")
	}
	if repo.lastCreatedSession == nil {
		t.Fatalf("expected session created")
	}
}

func TestValidateSession_Expired(t *testing.T) {
	repo := newFakeAuthRepo()
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13}
	service := NewService(repo, cfg, logger.NewDefault(false))

	repo.sessionsByToken["tok"] = domainauth.Session{
		UserID:       1,
		SessionToken: "tok",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}

	_, err := service.ValidateSession(context.Background(), "tok")
	if !errors.Is(err, domainauth.ErrSessionExpired) {
		t.Fatalf("expected session expired, got %v", err)
	}
}

func TestLogout_EmptyToken(t *testing.T) {
	repo := newFakeAuthRepo()
	cfg := config.AuthConfig{BcryptCost: 4, MinPasswordLength: 6, MinUserAge: 13}
	service := NewService(repo, cfg, logger.NewDefault(false))

	if err := service.Logout(context.Background(), ""); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func strPtr(s string) *string { return &s }
