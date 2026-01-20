package post

import "time"

// Post represents a social post in the domain layer.
type Post struct {
	ID        int64
	AuthorID  string
	Content   string
	Privacy   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
