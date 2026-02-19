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
	wsHandler *transportws.Handler,
	mw Middlewares,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
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
	mux.Handle("PATCH /posts/{id}", mw.Auth(http.HandlerFunc(postHandler.Update)))
	mux.Handle("DELETE /posts/{id}", mw.Auth(http.HandlerFunc(postHandler.Delete)))

	// Comment routes (protected)
	mux.Handle("POST /posts/{id}/comments", mw.Auth(http.HandlerFunc(commentHandler.Create)))
	mux.Handle("GET /posts/{id}/comments", mw.Auth(http.HandlerFunc(commentHandler.GetByPostID)))
	mux.Handle("PATCH /comments/{id}", mw.Auth(http.HandlerFunc(commentHandler.Update)))
	mux.Handle("DELETE /comments/{id}", mw.Auth(http.HandlerFunc(commentHandler.Delete)))

	// Post reaction routes (protected)
	mux.Handle("POST /posts/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddPostReaction)))
	mux.Handle("GET /posts/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.GetPostReactions)))

	// Comment reaction routes (protected)
	mux.Handle("POST /comments/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddCommentReaction)))
	mux.Handle("GET /comments/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.GetCommentReactions)))

	// Profile routes (protected)
	mux.Handle("GET /profiles/{id}", mw.Auth(http.HandlerFunc(profileHandler.GetProfile)))
	mux.Handle("GET /profiles/{id}/followers", mw.Auth(http.HandlerFunc(profileHandler.ListFollowers)))
	mux.Handle("GET /profiles/{id}/following", mw.Auth(http.HandlerFunc(profileHandler.ListFollowing)))
	mux.Handle("PATCH /profiles/{id}/visibility", mw.Auth(http.HandlerFunc(profileHandler.UpdateVisibility)))

	// Follow routes (protected)
	mux.Handle("GET /follow-requests", mw.Auth(http.HandlerFunc(followHandler.ListRequests)))
	mux.Handle("POST /follow-requests", mw.Auth(http.HandlerFunc(followHandler.CreateRequest)))
	mux.Handle("GET /follow-requests/sent", mw.Auth(http.HandlerFunc(followHandler.ListSentRequests)))
	mux.Handle("PATCH /follow-requests/{id}", mw.Auth(http.HandlerFunc(followHandler.UpdateRequest)))
	mux.Handle("DELETE /users/{id}/followers", mw.Auth(http.HandlerFunc(followHandler.Unfollow)))

	// User routes (protected)
	mux.Handle("GET /users", mw.Auth(http.HandlerFunc(userHandler.ListUsers)))

	// Notification routes (protected)
	mux.Handle("GET /notifications", mw.Auth(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("GET /notifications/unread-count", mw.Auth(http.HandlerFunc(notificationHandler.UnreadCount)))
	mux.Handle("PATCH /notifications/{id}/read", mw.Auth(http.HandlerFunc(notificationHandler.MarkRead)))
	mux.Handle("PATCH /notifications/read-all", mw.Auth(http.HandlerFunc(notificationHandler.MarkAllRead)))

	// Group routes (protected)
	mux.Handle("POST /groups", mw.Auth(http.HandlerFunc(groupHandler.Create)))
	mux.Handle("GET /groups", mw.Auth(http.HandlerFunc(groupHandler.List)))
	mux.Handle("GET /groups/{id}", mw.Auth(http.HandlerFunc(groupHandler.Get)))
	mux.Handle("PATCH /groups/{id}", mw.Auth(http.HandlerFunc(groupHandler.Update)))
	mux.Handle("DELETE /groups/{id}", mw.Auth(http.HandlerFunc(groupHandler.Delete)))
	mux.Handle("GET /groups/{id}/members", mw.Auth(http.HandlerFunc(groupHandler.ListMembers)))
	mux.Handle("DELETE /groups/{id}/members/me", mw.Auth(http.HandlerFunc(groupHandler.Leave)))
	mux.Handle("DELETE /groups/{id}/members/{user_id}", mw.Auth(http.HandlerFunc(groupHandler.RemoveMember)))
	mux.Handle("POST /groups/{id}/invitations", mw.Auth(http.HandlerFunc(groupHandler.Invite)))
	mux.Handle("GET /group-invitations", mw.Auth(http.HandlerFunc(groupHandler.ListInvitations)))
	mux.Handle("PATCH /group-invitations/{id}", mw.Auth(http.HandlerFunc(groupHandler.RespondInvitation)))
	mux.Handle("POST /groups/{id}/join-requests", mw.Auth(http.HandlerFunc(groupHandler.RequestJoin)))
	mux.Handle("GET /groups/{id}/join-requests", mw.Auth(http.HandlerFunc(groupHandler.ListJoinRequests)))
	mux.Handle("PATCH /groups/{id}/join-requests/{request_id}", mw.Auth(http.HandlerFunc(groupHandler.RespondJoinRequest)))
	mux.Handle("POST /groups/{id}/posts", mw.Auth(http.HandlerFunc(groupHandler.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts", mw.Auth(http.HandlerFunc(groupHandler.ListGroupPosts)))
	mux.Handle("POST /groups/{id}/posts/{post_id}/comments", mw.Auth(http.HandlerFunc(groupHandler.CreateGroupComment)))
	mux.Handle("GET /groups/{id}/posts/{post_id}/comments", mw.Auth(http.HandlerFunc(groupHandler.ListGroupComments)))
	mux.Handle("PATCH /group-comments/{id}", mw.Auth(http.HandlerFunc(groupHandler.UpdateGroupComment)))
	mux.Handle("DELETE /group-comments/{id}", mw.Auth(http.HandlerFunc(groupHandler.DeleteGroupComment)))
	mux.Handle("POST /groups/{id}/events", mw.Auth(http.HandlerFunc(groupHandler.CreateEvent)))
	mux.Handle("GET /groups/{id}/events", mw.Auth(http.HandlerFunc(groupHandler.ListEvents)))
	mux.Handle("POST /groups/{id}/events/{event_id}/rsvp", mw.Auth(http.HandlerFunc(groupHandler.RSVPEvent)))

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
