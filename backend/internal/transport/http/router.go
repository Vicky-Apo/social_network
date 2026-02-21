package http

import (
	"net/http"

	"social-network/backend/internal/transport/http/handler"
	transportws "social-network/backend/internal/transport/websocket"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Middlewares holds all middleware functions for the router
type Middlewares struct {
	Auth            Middleware
	RateLimit       Middleware
	CORS            Middleware
	SecurityHeaders Middleware
}

// NewRouter builds the HTTP router with all handlers.
// Middlewares are injected from outside, keeping the router decoupled from the usecase layer.
func NewRouter(
	postHandler *handler.PostHandler,
	authHandler *handler.AuthHandler,
	commentHandler *handler.CommentHandler,
	reactionHandler *handler.ReactionHandler,
	profileHandler *handler.ProfileHandler,
	followHandler *handler.FollowHandler,
	userHandler *handler.UserHandler,
	notificationHandler *handler.NotificationHandler,
	groupHandler *handler.GroupHandler,
	eventHandler *handler.EventHandler,
	chatHandler *handler.ChatHandler,
	messageReactionHandler *handler.MessageReactionHandler,
	uploadHandler *handler.UploadHandler,
	mediaHandler *handler.MediaHandler,
	uploadsDir string,
	wsHandler *transportws.Handler,
	mw Middlewares,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes (public)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// Auth routes (protected)
	mux.Handle("GET /auth/me", mw.Auth(http.HandlerFunc(authHandler.Me)))

	// Post routes
	mux.Handle("GET /posts", mw.Auth(http.HandlerFunc(postHandler.List)))
	mux.Handle("GET /posts/{id}", mw.Auth(http.HandlerFunc(postHandler.GetByID)))
	mux.Handle("POST /posts", mw.Auth(http.HandlerFunc(postHandler.Create)))

	// Comment routes (protected)
	mux.Handle("POST /posts/{id}/comments", mw.Auth(http.HandlerFunc(commentHandler.Create)))
	mux.Handle("GET /posts/{id}/comments", mw.Auth(http.HandlerFunc(commentHandler.GetByPostID)))

	// Post reaction routes (protected)
	mux.Handle("POST /posts/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddPostReaction)))
	mux.Handle("GET /posts/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.GetPostReactions)))

	// Comment reaction routes (protected)
	mux.Handle("POST /comments/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddCommentReaction)))
	mux.Handle("GET /comments/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.GetCommentReactions)))

	// Profile routes (protected)
	mux.Handle("GET /profiles/{id}", mw.Auth(http.HandlerFunc(profileHandler.GetProfile)))
	mux.Handle("GET /profiles/{id}/full", mw.Auth(http.HandlerFunc(profileHandler.GetProfileFull)))
	mux.Handle("GET /profiles/{id}/followers", mw.Auth(http.HandlerFunc(profileHandler.ListFollowers)))
	mux.Handle("GET /profiles/{id}/following", mw.Auth(http.HandlerFunc(profileHandler.ListFollowing)))
	mux.Handle("PATCH /profiles/{id}/visibility", mw.Auth(http.HandlerFunc(profileHandler.UpdateVisibility)))
	mux.Handle("PATCH /profiles/{id}", mw.Auth(http.HandlerFunc(profileHandler.UpdateProfile)))

	// Follow routes (protected)
	mux.Handle("GET /follow-requests", mw.Auth(http.HandlerFunc(followHandler.ListRequests)))
	mux.Handle("POST /follow-requests", mw.Auth(http.HandlerFunc(followHandler.CreateRequest)))
	mux.Handle("GET /follow-requests/sent", mw.Auth(http.HandlerFunc(followHandler.ListSentRequests)))
	mux.Handle("PATCH /follow-requests/{id}", mw.Auth(http.HandlerFunc(followHandler.UpdateRequest)))
	mux.Handle("DELETE /users/{id}/followers", mw.Auth(http.HandlerFunc(followHandler.Unfollow)))
	mux.Handle("DELETE /followers/{id}", mw.Auth(http.HandlerFunc(followHandler.RemoveFollower)))

	// User routes (protected)
	mux.Handle("GET /users", mw.Auth(http.HandlerFunc(userHandler.ListUsers)))

	// Group routes (protected)
	mux.Handle("POST /groups", mw.Auth(http.HandlerFunc(groupHandler.Create)))
	mux.Handle("GET /groups", mw.Auth(http.HandlerFunc(groupHandler.List)))
	mux.Handle("GET /groups/{id}", mw.Auth(http.HandlerFunc(groupHandler.GetByID)))
	mux.Handle("GET /groups/{id}/members", mw.Auth(http.HandlerFunc(groupHandler.ListMembers)))
	mux.Handle("GET /groups/{id}/posts", mw.Auth(http.HandlerFunc(postHandler.ListByGroup)))
	mux.Handle("POST /groups/{id}/posts", mw.Auth(http.HandlerFunc(postHandler.CreateInGroup)))
	mux.Handle("POST /groups/{id}/events", mw.Auth(http.HandlerFunc(eventHandler.Create)))
	mux.Handle("GET /groups/{id}/events", mw.Auth(http.HandlerFunc(eventHandler.ListByGroup)))
	mux.Handle("POST /groups/{id}/invitations", mw.Auth(http.HandlerFunc(groupHandler.Invite)))
	mux.Handle("GET /group-invitations", mw.Auth(http.HandlerFunc(groupHandler.ListInvitations)))
	mux.Handle("PATCH /group-invitations/{id}", mw.Auth(http.HandlerFunc(groupHandler.UpdateInvitation)))
	mux.Handle("POST /groups/{id}/join-requests", mw.Auth(http.HandlerFunc(groupHandler.RequestJoin)))
	mux.Handle("GET /groups/{id}/join-requests", mw.Auth(http.HandlerFunc(groupHandler.ListJoinRequests)))
	mux.Handle("PATCH /group-join-requests/{id}", mw.Auth(http.HandlerFunc(groupHandler.UpdateJoinRequest)))
	mux.Handle("DELETE /groups/{id}/members/me", mw.Auth(http.HandlerFunc(groupHandler.LeaveGroup)))

	// Event routes (protected)
	mux.Handle("GET /events/{id}", mw.Auth(http.HandlerFunc(eventHandler.GetByID)))
	mux.Handle("PATCH /events/{id}", mw.Auth(http.HandlerFunc(eventHandler.Update)))
	mux.Handle("DELETE /events/{id}", mw.Auth(http.HandlerFunc(eventHandler.Delete)))
	mux.Handle("POST /events/{id}/responses", mw.Auth(http.HandlerFunc(eventHandler.Respond)))
	mux.Handle("GET /events/{id}/responses", mw.Auth(http.HandlerFunc(eventHandler.ListResponses)))

	// Notification routes (protected)
	mux.Handle("GET /notifications", mw.Auth(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("GET /notifications/unread-count", mw.Auth(http.HandlerFunc(notificationHandler.UnreadCount)))
	mux.Handle("PATCH /notifications/{id}/read", mw.Auth(http.HandlerFunc(notificationHandler.MarkRead)))
	mux.Handle("PATCH /notifications/read-all", mw.Auth(http.HandlerFunc(notificationHandler.MarkAllRead)))

	// Chat routes (protected)
	mux.Handle("GET /conversations", mw.Auth(http.HandlerFunc(chatHandler.ListConversations)))
	mux.Handle("GET /conversations/unread-counts", mw.Auth(http.HandlerFunc(chatHandler.UnreadCounts)))
	mux.Handle("GET /conversations/{id}", mw.Auth(http.HandlerFunc(chatHandler.GetConversation)))
	mux.Handle("GET /conversations/{id}/messages", mw.Auth(http.HandlerFunc(chatHandler.ListMessages)))
	mux.Handle("PATCH /conversations/{id}/read", mw.Auth(http.HandlerFunc(chatHandler.MarkRead)))

	// Message reaction routes (protected)
	mux.Handle("POST /messages/{id}/reactions", mw.Auth(http.HandlerFunc(messageReactionHandler.Toggle)))
	mux.Handle("GET /messages/{id}/reactions", mw.Auth(http.HandlerFunc(messageReactionHandler.List)))

	// Upload routes (protected)
	mux.Handle("POST /uploads", mw.Auth(http.HandlerFunc(uploadHandler.Upload)))
	if uploadsDir != "" && mediaHandler != nil {
		mux.Handle("GET /uploads/{path...}", mw.Auth(http.HandlerFunc(mediaHandler.Serve)))
	}

	// WebSocket route (authentication handled inside the handler)
	mux.Handle("/ws", wsHandler)

	// Apply global middlewares (order: security headers -> CORS -> rate limiting)
	// Security headers are applied first so they're present on all responses
	// CORS is applied second to handle preflight requests
	// Rate limiting is applied last to protect the application
	var handler http.Handler = mux
	handler = mw.RateLimit(handler)
	handler = mw.CORS(handler)
	handler = mw.SecurityHeaders(handler)

	return handler
}
