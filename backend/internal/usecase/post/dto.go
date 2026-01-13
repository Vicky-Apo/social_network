package post

import "time"

// PostDTO is the application-facing representation of a post.
type PostDTO struct {
	ID        int64     `json:"id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
