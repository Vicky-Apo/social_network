package utils

const (
	MsgInternalServerError   = "internal server error"
	MsgInvalidRequestBody    = "invalid request body"
	MsgUnauthorized          = "unauthorized"
	MsgForbidden             = "forbidden"
	MsgNotFound              = "not found"
	MsgInvalidPostID         = "invalid post id"
	MsgInvalidCommentID      = "invalid comment id"
	MsgInvalidUserID         = "invalid user id"
	MsgInvalidCategoryID     = "invalid category_id"
	MsgInvalidAuthorID       = "invalid author_id"
	MsgCommentsNotFound      = "comments not found"
	MsgProfileNotFound       = "profile not found"
	MsgPostNotFound          = "post not found"
	MsgUserNotFound          = "user not found"
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
