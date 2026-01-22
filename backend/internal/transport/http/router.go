package http

import (
	"net/http"

	"social-network/backend/internal/transport/http/handler"
)

// NewRouter builds the HTTP router with all handlers.
func NewRouter(postHandler *handler.PostHandler, authHandler *handler.AuthHandler) http.Handler {
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
	mux.HandleFunc("GET /auth/me", authHandler.Me)

	// Post routes
	mux.HandleFunc("GET /posts", postHandler.List)
	mux.HandleFunc("GET /posts/", postHandler.GetByID)

	return mux
}
