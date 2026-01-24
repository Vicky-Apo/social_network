package http

import (
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/handler"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// NewRouter builds the HTTP router with all handlers.
func NewRouter(
	postHandler *handler.PostHandler,
	authHandler *handler.AuthHandler,
	profileHandler *handler.ProfileHandler,
	followHandler *handler.FollowHandler,
	authMiddleware Middleware,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/posts", postHandler.List)
	mux.HandleFunc("/posts/", postHandler.GetByID)

	// Auth routes (public)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// Auth routes (protected)
	mux.Handle("GET /auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Profile routes (protected)
	mux.Handle("/profiles", authMiddleware(http.HandlerFunc(profileHandler.GetProfile)))
	mux.Handle("/profiles/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/followers"):
			profileHandler.ListFollowers(w, r)
		case strings.HasSuffix(r.URL.Path, "/following"):
			profileHandler.ListFollowing(w, r)
		case strings.HasSuffix(r.URL.Path, "/visibility"):
			profileHandler.UpdateVisibility(w, r)
		default:
			profileHandler.GetProfile(w, r)
		}
	})))

	// Follow routes (protected)
	mux.Handle("/follow-requests", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			followHandler.ListRequests(w, r)
		case http.MethodPost:
			followHandler.CreateRequest(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/follow-requests/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/accept"):
			followHandler.AcceptRequest(w, r)
		case strings.HasSuffix(r.URL.Path, "/decline"):
			followHandler.DeclineRequest(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})))
	mux.Handle("/unfollow", authMiddleware(http.HandlerFunc(followHandler.Unfollow)))

	return mux
}
