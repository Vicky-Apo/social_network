package http

import (
	"net/http"
	"strings"

	"social-network/backend/internal/transport/http/handler"
)

// NewRouter builds the HTTP router with all handlers.
func NewRouter(
	postHandler *handler.PostHandler,
	profileHandler *handler.ProfileHandler,
	followHandler *handler.FollowHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/posts", postHandler.List)
	mux.HandleFunc("/posts/", postHandler.GetByID)

	mux.HandleFunc("/profiles", profileHandler.GetProfile)
	mux.HandleFunc("/profiles/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("/follow-requests", followHandler.CreateRequest)
	mux.HandleFunc("/follow-requests/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/accept"):
			followHandler.AcceptRequest(w, r)
		case strings.HasSuffix(r.URL.Path, "/decline"):
			followHandler.DeclineRequest(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	mux.HandleFunc("/unfollow", followHandler.Unfollow)

	return mux
}
