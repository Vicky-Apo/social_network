package utils

const (
	MsgInternalServerError   = "internal server error"
	MsgInvalidRequestBody    = "invalid request body"
	MsgUnauthorized          = "unauthorized"
	MsgForbidden             = "forbidden"
	MsgNotFound              = "not found"
	MsgInvalidPostID         = "invalid post id"
	MsgInvalidCommentID      = "invalid comment id"
	MsgCommentNotFound       = "comment not found"
	MsgInvalidUserID         = "invalid user id"
	MsgInvalidAuthorID       = "invalid author_id"
	MsgInvalidNotificationID = "invalid notification id"
	MsgCommentsNotFound      = "comments not found"
	MsgProfileNotFound       = "profile not found"
	MsgPostNotFound          = "post not found"
	MsgUserNotFound          = "user not found"
	MsgInvalidGroupID        = "invalid group id"
	MsgGroupNotFound         = "group not found"
	MsgInvalidInvitationID   = "invalid invitation id"
	MsgInvitationNotFound    = "invitation not found"
	MsgInvalidJoinRequestID  = "invalid join request id"
	MsgJoinRequestNotFound   = "join request not found"
	MsgInvalidEventID        = "invalid event id"
	MsgEventNotFound         = "event not found"
	MsgFollowRequestNotFound = "follow request not found"
	MsgFollowRequestExists   = "follow request already exists"
	MsgCannotFollowSelf      = "cannot follow self"
	MsgNotFollowing          = "not following"
	MsgFollowNotPending      = "follow request is not pending"
	MsgInvalidStatus         = "invalid status"
	MsgInvalidLimit          = "invalid limit"
	MsgInvalidOffset         = "invalid offset"
	MsgInvalidCredentials    = "invalid credentials"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)
