package profile

import "time"

// UserDTO is the application-facing representation of a user profile.
type UserDTO struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"` // "DD/MM/YYYY" format
	AvatarPath  *string   `json:"avatar_path,omitempty"`
	Nickname    *string   `json:"nickname,omitempty"`
	About       *string   `json:"about,omitempty"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LimitedUserDTO is a reduced profile view for private profiles.
type LimitedUserDTO struct {
	ID         int64   `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Nickname   *string `json:"nickname,omitempty"`
	AvatarPath *string `json:"avatar_path,omitempty"`
	IsPublic   bool    `json:"is_public"`
}

// ProfileUserDTO is a superset user view for profile responses.
// Optional fields are omitted when unavailable (private profile).
type ProfileUserDTO struct {
	ID          int64      `json:"id"`
	Email       *string    `json:"email,omitempty"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	DateOfBirth *string    `json:"date_of_birth,omitempty"` // "DD/MM/YYYY" format
	AvatarPath  *string    `json:"avatar_path,omitempty"`
	Nickname    *string    `json:"nickname,omitempty"`
	About       *string    `json:"about,omitempty"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// ProfileDTO combines a user with follower stats.
type ProfileDTO struct {
	User           ProfileUserDTO `json:"user"`
	FollowersCount *int64 `json:"followers_count,omitempty"`
	FollowingCount *int64 `json:"following_count,omitempty"`
	IsFollowing    bool   `json:"is_following"`
	IsFollowedBy   bool   `json:"is_followed_by"`
	Limited        bool   `json:"limited,omitempty"`
}

// UpdateProfileRequest represents profile updates.
type UpdateProfileRequest struct {
	Nickname   *string `json:"nickname,omitempty"`
	About      *string `json:"about,omitempty"`
	AvatarPath *string `json:"avatar_path,omitempty"`
}
