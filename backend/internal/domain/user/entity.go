package user

import "time"

// User represents a user profile in the domain layer.
type User struct {
	ID          int64
	Email       string
	FirstName   string
	LastName    string
	DateOfBirth time.Time
	AvatarPath  *string
	Nickname    *string
	About       *string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
