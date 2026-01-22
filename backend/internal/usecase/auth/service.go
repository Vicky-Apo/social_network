package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	domainauth "social-network/backend/internal/domain/auth"
)

const (
	bcryptCost        = 12                      // Security/performance balance
	sessionTokenBytes = 32                      // 256-bit entropy
	sessionDuration   = 30 * 24 * time.Hour     // 30 days
	minPasswordLength = 8                       // Minimum password length
	minAge            = 13                      // Minimum age in years
)

var (
	// Email validation regex (basic validation)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Common errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidDateFormat  = errors.New("invalid date format, expected DD/MM/YYYY")
	ErrInvalidDateOfBirth = errors.New("invalid date of birth")
	ErrUserTooYoung       = errors.New("user must be at least 13 years old")
)

// Service orchestrates authentication use cases
type Service struct {
	repo domainauth.Repository
}

// NewService creates a new authentication service
func NewService(repo domainauth.Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (UserDTO, error) {
	// Validate input
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if err := validateEmail(req.Email); err != nil {
		return UserDTO{}, err
	}

	if len(req.Password) < minPasswordLength {
		return UserDTO{}, ErrWeakPassword
	}

	if req.FirstName == "" || req.LastName == "" {
		return UserDTO{}, errors.New("first name and last name are required")
	}

	// Parse and validate date of birth
	dateOfBirth, err := parseDate(req.DateOfBirth)
	if err != nil {
		return UserDTO{}, err
	}

	if err := validateDateOfBirth(dateOfBirth); err != nil {
		return UserDTO{}, err
	}

	// Hash password
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return UserDTO{}, fmt.Errorf("hash password: %w", err)
	}

	// Create user entity
	user := domainauth.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  dateOfBirth,
		IsPublic:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domainauth.ErrEmailAlreadyExists) {
			return UserDTO{}, domainauth.ErrEmailAlreadyExists
		}
		return UserDTO{}, fmt.Errorf("create user: %w", err)
	}

	user.ID = userID

	return mapUserToDTO(user), nil
}

// Login authenticates user and creates session
func (s *Service) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (LoginResponse, error) {
	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, fmt.Errorf("get user: %w", err)
	}

	// Verify password
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	// Generate session token
	sessionToken, err := generateSessionToken()
	if err != nil {
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
		return LoginResponse{}, fmt.Errorf("create session: %w", err)
	}

	session.ID = sessionID

	return LoginResponse{
		User:  mapUserToDTO(user),
		Token: sessionToken,
	}, nil
}

// Logout destroys the current session
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return nil // Already logged out
	}

	if err := s.repo.DeleteSession(ctx, sessionToken); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// GetCurrentUser retrieves user by session token
func (s *Service) GetCurrentUser(ctx context.Context, sessionToken string) (UserDTO, error) {
	if sessionToken == "" {
		return UserDTO{}, domainauth.ErrSessionNotFound
	}

	// Get session
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, domainauth.ErrSessionNotFound) {
			return UserDTO{}, domainauth.ErrSessionNotFound
		}
		return UserDTO{}, fmt.Errorf("get session: %w", err)
	}

	// Check expiry (database query already filters, but double-check)
	if time.Now().After(session.ExpiresAt) {
		return UserDTO{}, domainauth.ErrSessionExpired
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, session.UserID)
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
		return 0, domainauth.ErrSessionNotFound
	}

	// Get session
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, domainauth.ErrSessionNotFound) {
			return 0, domainauth.ErrSessionNotFound
		}
		return 0, fmt.Errorf("get session: %w", err)
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		return 0, domainauth.ErrSessionExpired
	}

	return session.UserID, nil
}

// Helper functions

// generateSessionToken creates a secure random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// comparePassword verifies a password against a hash
func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// validateEmail checks if email format is valid
func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// parseDate converts "DD/MM/YYYY" string to time.Time
func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}
	return t, nil
}

// validateDateOfBirth checks if date of birth is valid
func validateDateOfBirth(dob time.Time) error {
	now := time.Now()

	// Check if date is in the future
	if dob.After(now) {
		return ErrInvalidDateOfBirth
	}

	// Check minimum age
	minDate := now.AddDate(-minAge, 0, 0)
	if dob.After(minDate) {
		return ErrUserTooYoung
	}

	// Check reasonable maximum age (150 years)
	maxDate := now.AddDate(-150, 0, 0)
	if dob.Before(maxDate) {
		return ErrInvalidDateOfBirth
	}

	return nil
}

// mapUserToDTO converts domain User to UserDTO
func mapUserToDTO(user domainauth.User) UserDTO {
	return UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth,
		AvatarPath:  user.AvatarPath,
		Nickname:    user.Nickname,
		About:       user.About,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
	}
}
