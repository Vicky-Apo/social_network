package http

import (
	"net/http"

	"social-network/backend/internal/transport/http/handler"
)

// NewRouter builds the HTTP router with all handlers.
func NewRouter(postHandler *handler.PostHandler, commentHandler *handler.CommentHandler, reactionHandler *handler.ReactionHandler) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Post endpoints
	mux.HandleFunc("/posts", postHandler.List)
	mux.HandleFunc("/posts/", routePostPaths(postHandler, commentHandler, reactionHandler))

	// Comment reaction endpoints
	mux.HandleFunc("/comments/", routeCommentPaths(reactionHandler))

	return mux
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
