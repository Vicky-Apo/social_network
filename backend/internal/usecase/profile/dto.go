package profile

import "time"

// UserDTO is the application-facing representation of a user profile.
type UserDTO struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	AvatarPath  *string   `json:"avatar_path,omitempty"`
	Nickname    *string   `json:"nickname,omitempty"`
	About       *string   `json:"about,omitempty"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProfileDTO combines a user with follower stats.
type ProfileDTO struct {
	User           UserDTO `json:"user"`
	FollowersCount int64   `json:"followers_count"`
	FollowingCount int64   `json:"following_count"`
}
