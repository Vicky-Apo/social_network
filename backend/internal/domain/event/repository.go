package event

import (
	"context"
	"errors"
)

// ErrNotFound is returned when an event does not exist.
var ErrNotFound = errors.New("event not found")

// Repository defines the data access contract for events.
type Repository interface {
	Create(ctx context.Context, e Event) (Event, error)
	GetByID(ctx context.Context, id int64) (Event, error)
	ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]Event, error)
	Update(ctx context.Context, e Event) (Event, error)
	Delete(ctx context.Context, id int64) error
	UpsertResponse(ctx context.Context, eventID, userID int64, response string) (EventResponse, error)
	ListResponses(ctx context.Context, eventID int64) ([]EventResponseInfo, error)
}
