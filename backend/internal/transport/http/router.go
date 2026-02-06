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

	// Comment routes
	// Comment routes (protected)
mux.Handle("POST /posts/{id}/comments", mw.Auth(http.HandlerFunc(commentHandler.Create)))
mux.HandleFunc("GET /posts/{id}/comments", commentHandler.GetByPostID)  // Can be public

// Post reaction routes (protected)
mux.Handle("POST /posts/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddPostReaction)))
mux.HandleFunc("GET /posts/{id}/reactions", reactionHandler.GetPostReactions)  // Can be public

// Comment reaction routes (protected)
mux.Handle("POST /comments/{id}/reactions", mw.Auth(http.HandlerFunc(reactionHandler.AddCommentReaction)))
mux.HandleFunc("GET /comments/{id}/reactions", reactionHandler.GetCommentReactions)  // Can be public

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
