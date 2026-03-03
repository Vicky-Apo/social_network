package follow

import "time"

// FollowRequestDTO is the application-facing representation of a follow request.
type FollowRequestDTO struct {
	ID          int64     `json:"id"`
	RequesterID int64     `json:"requester_id"`
	TargetID    int64     `json:"target_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	Requester   *UserDTO  `json:"requester,omitempty"`
	Target      *UserDTO  `json:"target,omitempty"`
}

// FollowResultDTO describes the outcome of a follow attempt.
type FollowResultDTO struct {
	Status  string            `json:"status"`
	Request *FollowRequestDTO `json:"request,omitempty"`
}

// UserDTO is a lightweight user view for follow requests.
type UserDTO struct {
	ID         int64   `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Nickname   *string `json:"nickname,omitempty"`
	AvatarPath *string `json:"avatar_path,omitempty"`
}
