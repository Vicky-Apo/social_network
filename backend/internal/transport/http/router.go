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
	mux.HandleFunc("GET /posts/", postHandler.GetByID)

	// Comment and reaction routes
	mux.HandleFunc("/posts/", routePostPaths(postHandler, commentHandler, reactionHandler))
	mux.HandleFunc("/comments/", routeCommentPaths(reactionHandler))

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

// Route /posts/* paths
func routePostPaths(postHandler *handler.PostHandler, commentHandler *handler.CommentHandler, reactionHandler *handler.ReactionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check for /posts/{id}/comments
		if len(path) > 9 && path[len(path)-9:] == "/comments" {
			if r.Method == http.MethodGet {
				commentHandler.GetByPostID(w, r)
			} else if r.Method == http.MethodPost {
				commentHandler.Create(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// Check for /posts/{id}/reactions
		if len(path) > 10 && path[len(path)-10:] == "/reactions" {
			if r.Method == http.MethodGet {
				reactionHandler.GetPostReactions(w, r)
			} else if r.Method == http.MethodPost {
				reactionHandler.AddPostReaction(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// Default: /posts/{id}
		postHandler.GetByID(w, r)
	}
}

// Route /comments/* paths
func routeCommentPaths(reactionHandler *handler.ReactionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check for /comments/{id}/reactions
		if len(path) > 10 && path[len(path)-10:] == "/reactions" {
			if r.Method == http.MethodGet {
				reactionHandler.GetCommentReactions(w, r)
			} else if r.Method == http.MethodPost {
				reactionHandler.AddCommentReaction(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}
