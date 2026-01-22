package auth

import "time"

// Request DTOs

// RegisterRequest represents the user registration payload
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"` // "DD/MM/YYYY" format
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response DTOs

// UserDTO represents user data for API responses (excludes sensitive fields)
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
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	User  UserDTO `json:"user"`
	Token string  `json:"token"` // session token for client storage
}
