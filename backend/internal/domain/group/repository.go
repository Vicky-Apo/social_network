package group

import (
	"context"
	"errors"
)

// Errors for the group domain.
var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrNotMember           = errors.New("user is not a member of this group")
	ErrInvitationNotFound  = errors.New("group invitation not found")
	ErrJoinRequestNotFound = errors.New("group join request not found")
	ErrInvitationExists    = errors.New("group invitation already exists")
	ErrJoinRequestExists   = errors.New("group join request already exists")
)

// Repository defines the data access contract for group operations.
type Repository interface {
	Create(ctx context.Context, creatorID int64, title string, description *string) (Group, error)
	List(ctx context.Context, userID int64, limit, offset int) ([]GroupSummary, error)
	Search(ctx context.Context, userID int64, query string, limit, offset int) ([]GroupSummary, error)
	GetWithMeta(ctx context.Context, userID, id int64) (GroupSummary, error)
	GetByID(ctx context.Context, id int64) (Group, error)
	IsMember(ctx context.Context, groupID, userID int64) (bool, error)
	GetMemberIDs(ctx context.Context, groupID int64) ([]int64, error)
	ListMembers(ctx context.Context, groupID int64) ([]GroupMemberInfo, error)
	AddMember(ctx context.Context, groupID, userID int64) error
	RemoveMember(ctx context.Context, groupID, userID int64) error

	InvitationExists(ctx context.Context, groupID, inviteeID int64) (bool, error)
	CreateInvitation(ctx context.Context, groupID, inviterID, inviteeID int64) (GroupInvitation, error)
	GetInvitationByID(ctx context.Context, id int64) (GroupInvitation, error)
	ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]GroupInvitation, error)
	DeleteInvitation(ctx context.Context, id int64) error

	JoinRequestExists(ctx context.Context, groupID, userID int64) (bool, error)
	CreateJoinRequest(ctx context.Context, groupID, userID int64) (GroupJoinRequest, error)
	GetJoinRequestByID(ctx context.Context, id int64) (GroupJoinRequest, error)
	ListJoinRequestsByGroup(ctx context.Context, groupID int64) ([]GroupJoinRequest, error)
	DeleteJoinRequest(ctx context.Context, id int64) error
}
