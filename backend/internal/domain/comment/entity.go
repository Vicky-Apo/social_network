package comment

import "time"

// Comment represents a comment on a post
type Comment struct {
	ID        int64
	PostID    int64
	AuthorID  int64
	Content   string
	MediaPath string
	CreatedAt time.Time
	UpdatedAt time.Time
}