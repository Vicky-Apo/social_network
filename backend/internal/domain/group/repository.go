package group

import (
	"context"
	"errors"
)

// Errors for the group domain.
var (
	ErrGroupNotFound = errors.New("group not found")
	ErrNotMember     = errors.New("user is not a member of this group")
	ErrInvitationNotFound = errors.New("group invitation not found")
	ErrJoinRequestNotFound = errors.New("group join request not found")
	ErrEventNotFound = errors.New("event not found")
	ErrEventResponseNotFound = errors.New("event response not found")
)

// Repository defines the data access contract for group operations.
type Repository interface {
	GetByID(ctx context.Context, id int64) (Group, error)
	List(ctx context.Context, query string, limit, offset int) ([]Group, error)
	Create(ctx context.Context, group Group) (Group, error)
	Update(ctx context.Context, group Group) (Group, error)
	Delete(ctx context.Context, id int64) error

	IsMember(ctx context.Context, groupID, userID int64) (bool, error)
	GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error)
	ListMembers(ctx context.Context, groupID int64, limit, offset int) ([]GroupMember, error)
	AddMember(ctx context.Context, groupID, userID int64) error
	RemoveMember(ctx context.Context, groupID, userID int64) error

	CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (GroupInvitation, error)
	GetInvitationByID(ctx context.Context, id int64) (GroupInvitation, error)
	ListInvitationsByInvitee(ctx context.Context, inviteeID int64, limit, offset int) ([]GroupInvitation, error)
	ListInvitationsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]GroupInvitation, error)
	DeleteInvitation(ctx context.Context, id int64) error

	CreateJoinRequest(ctx context.Context, groupID, userID int64) (GroupJoinRequest, error)
	GetJoinRequestByID(ctx context.Context, id int64) (GroupJoinRequest, error)
	ListJoinRequestsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]GroupJoinRequest, error)
	DeleteJoinRequest(ctx context.Context, id int64) error

	CreateEvent(ctx context.Context, event GroupEvent) (GroupEvent, error)
	GetEventByID(ctx context.Context, id int64) (GroupEvent, error)
	ListEventsByGroup(ctx context.Context, groupID int64, limit, offset int) ([]GroupEvent, error)
	UpdateEvent(ctx context.Context, event GroupEvent) (GroupEvent, error)
	DeleteEvent(ctx context.Context, id int64) error

	UpsertEventResponse(ctx context.Context, eventID, userID int64, response string) error
	GetEventResponse(ctx context.Context, eventID, userID int64) (GroupEventResponse, error)
	ListEventResponses(ctx context.Context, eventID int64) ([]GroupEventResponse, error)
}
