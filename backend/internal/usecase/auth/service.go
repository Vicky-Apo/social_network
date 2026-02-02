package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"social-network/backend/internal/config"
	domainauth "social-network/backend/internal/domain/auth"
	"social-network/backend/pkg/logger"
)

// Service-level errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service orchestrates authentication use cases
type Service struct {
	repo   domainauth.Repository
	config config.AuthConfig
	log    logger.Logger
}

// NewService creates a new authentication service
func NewService(repo domainauth.Repository, cfg config.AuthConfig, log logger.Logger) *Service {
	return &Service{
		repo:   repo,
		config: cfg,
		log:    log.WithFields(logger.F("service", "auth")),
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (UserDTO, error) {
	s.log.Debug("registration attempt", logger.F("email", req.Email))

	// Trim whitespace from input
	TrimStrings(&req.Email, &req.FirstName, &req.LastName)

	// Validate email
	if err := ValidateEmail(req.Email); err != nil {
		s.log.Debug("registration failed: invalid email", logger.F("email", req.Email))
		return UserDTO{}, err
	}

	// Validate password
	if err := ValidatePassword(req.Password, s.config.MinPasswordLength); err != nil {
		s.log.Debug("registration failed: weak password", logger.F("email", req.Email))
		return UserDTO{}, err
	}

	// Validate name
	if err := ValidateName(req.FirstName, req.LastName); err != nil {
		s.log.Debug("registration failed: missing name", logger.F("email", req.Email))
		return UserDTO{}, err
	}

	// Parse and validate date of birth
	dateOfBirth, err := ParseDate(req.DateOfBirth)
	if err != nil {
		s.log.Debug("registration failed: invalid date format", logger.F("email", req.Email))
		return UserDTO{}, err
	}

	if err := ValidateDateOfBirth(dateOfBirth, s.config.MinUserAge); err != nil {
		s.log.Debug("registration failed: invalid date of birth", logger.F("email", req.Email))
		return UserDTO{}, err
	}

	// Hash password
	passwordHash, err := hashPassword(req.Password, s.config.BcryptCost)
	if err != nil {
		s.log.Error("registration failed: password hashing error", err, logger.F("email", req.Email))
		return UserDTO{}, fmt.Errorf("hash password: %w", err)
	}

	// Create user entity
	user := domainauth.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  dateOfBirth,
		AvatarPath:   nil,
		Nickname:     req.Nickname,
		About:        req.About,
		IsPublic:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domainauth.ErrEmailAlreadyExists) {
			s.log.Debug("registration failed: email exists", logger.F("email", req.Email))
			return UserDTO{}, domainauth.ErrEmailAlreadyExists
		}
		s.log.Error("registration failed: database error", err, logger.F("email", req.Email))
		return UserDTO{}, fmt.Errorf("create user: %w", err)
	}

	user.ID = userID
	s.log.Info("user registered successfully", logger.F("user_id", userID), logger.F("email", req.Email))

	return mapUserToDTO(user), nil
}

// Login authenticates user and creates session
func (s *Service) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (LoginResponse, error) {
	s.log.Debug("login attempt", logger.F("email", req.Email), logger.F("ip", ipAddress))

	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			s.log.Debug("login failed: user not found", logger.F("email", req.Email))
			return LoginResponse{}, ErrInvalidCredentials
		}
		s.log.Error("login failed: database error", err, logger.F("email", req.Email))
		return LoginResponse{}, fmt.Errorf("get user: %w", err)
	}

	// Verify password
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		s.log.Debug("login failed: invalid password", logger.F("email", req.Email))
		return LoginResponse{}, ErrInvalidCredentials
	}

	// Generate session token
	sessionToken, err := generateSessionToken(s.config.SessionTokenBytes)
	if err != nil {
		s.log.Error("login failed: token generation error", err, logger.F("email", req.Email))
		return LoginResponse{}, fmt.Errorf("generate session token: %w", err)
	}

	// Create session
	var userAgentPtr, ipAddressPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	sessionDuration := time.Duration(s.config.SessionDurationDays) * 24 * time.Hour
	session := domainauth.Session{
		UserID:       user.ID,
		SessionToken: sessionToken,
		UserAgent:    userAgentPtr,
		IPAddress:    ipAddressPtr,
		ExpiresAt:    time.Now().Add(sessionDuration),
		CreatedAt:    time.Now(),
	}

	sessionID, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		s.log.Error("login failed: session creation error", err, logger.F("email", req.Email))
		return LoginResponse{}, fmt.Errorf("create session: %w", err)
	}

	session.ID = sessionID
	s.log.Info("user logged in successfully", logger.F("user_id", user.ID), logger.F("session_id", sessionID))

	return LoginResponse{
		User:  mapUserToDTO(user),
		Token: sessionToken,
	}, nil
}

// Logout destroys the current session
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		s.log.Debug("logout called with empty session token")
		return nil // Already logged out
	}

	if err := s.repo.DeleteSession(ctx, sessionToken); err != nil {
		s.log.Error("failed to delete session", err)
		return fmt.Errorf("delete session: %w", err)
	}

	s.log.Debug("session deleted successfully")
	return nil
}

// GetUserByID retrieves a user by their ID (assumes session already validated)
func (s *Service) GetUserByID(ctx context.Context, userID int64) (UserDTO, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			return UserDTO{}, domainauth.ErrUserNotFound
		}
		return UserDTO{}, fmt.Errorf("get user: %w", err)
	}

	return mapUserToDTO(user), nil
}

// ValidateSession checks if session is valid and returns user ID
func (s *Service) ValidateSession(ctx context.Context, sessionToken string) (int64, error) {
	if sessionToken == "" {
		s.log.Debug("session validation failed: empty token")
		return 0, domainauth.ErrSessionNotFound
	}

	// Get session
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, domainauth.ErrSessionNotFound) {
			s.log.Debug("session validation failed: session not found")
			return 0, domainauth.ErrSessionNotFound
		}
		s.log.Error("session validation failed: database error", err)
		return 0, fmt.Errorf("get session: %w", err)
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		s.log.Debug("session validation failed: session expired", logger.F("user_id", session.UserID))
		return 0, domainauth.ErrSessionExpired
	}

	return session.UserID, nil
}

// Helper functions

// generateSessionToken creates a secure random session token
func generateSessionToken(tokenBytes int) (string, error) {
	bytes := make([]byte, tokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// comparePassword verifies a password against a hash
func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// mapUserToDTO converts domain User to UserDTO
func mapUserToDTO(user domainauth.User) UserDTO {
	return UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth.Format("02/01/2006"),
		AvatarPath:  user.AvatarPath,
		Nickname:    user.Nickname,
		About:       user.About,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
	}
}
