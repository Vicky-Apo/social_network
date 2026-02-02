package auth

import "time"

// User represents a registered user in the system
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	DateOfBirth  time.Time
	AvatarPath   *string // nullable
	Nickname     *string // nullable
	About        *string // nullable
	IsPublic     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session represents an authenticated session
type Session struct {
	ID           int64
	UserID       int64
	SessionToken string
	UserAgent    *string // nullable
	IPAddress    *string // nullable
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
