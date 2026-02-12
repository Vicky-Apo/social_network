package media

import (
	"context"
	"errors"
)

// ErrNotFound is returned when no media record matches a path.
var ErrNotFound = errors.New("media not found")

// MediaType describes the owner of a media path.
type MediaType string

const (
	MediaTypePost    MediaType = "post"
	MediaTypeComment MediaType = "comment"
	MediaTypeMessage MediaType = "message"
	MediaTypeAvatar  MediaType = "avatar"
)

// MediaRef describes where a media path is used.
type MediaRef struct {
	Type           MediaType
	PostID         int64
	CommentID      int64
	MessageID      int64
	UserID         int64
	ConversationID int64
}

// Repository defines data access for media lookups.
type Repository interface {
	FindByPath(ctx context.Context, path string) (MediaRef, error)
}
