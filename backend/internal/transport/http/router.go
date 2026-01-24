package http

import (
	"net/http"

	"social-network/backend/internal/transport/http/handler"
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
func NewRouter(postHandler *handler.PostHandler, authHandler *handler.AuthHandler, commentHandler *handler.CommentHandler, reactionHandler *handler.ReactionHandler, mw Middlewares) http.Handler {
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
	mux.HandleFunc("GET /posts", postHandler.List)
	mux.HandleFunc("GET /posts/{id}", postHandler.GetByID)

	// Comment routes
	mux.HandleFunc("POST /posts/{id}/comments", commentHandler.Create)
	mux.HandleFunc("GET /posts/{id}/comments", commentHandler.GetByPostID)

	// Post reaction routes
	mux.HandleFunc("POST /posts/{id}/reactions", reactionHandler.AddPostReaction)
	mux.HandleFunc("GET /posts/{id}/reactions", reactionHandler.GetPostReactions)

	// Comment reaction routes
	mux.HandleFunc("POST /comments/{id}/reactions", reactionHandler.AddCommentReaction)
	mux.HandleFunc("GET /comments/{id}/reactions", reactionHandler.GetCommentReactions)

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
