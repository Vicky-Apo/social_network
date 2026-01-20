package http

import (
	"net/http"

	"social-network/backend/internal/transport/http/handler"
)

// NewRouter builds the HTTP router with all handlers.
func NewRouter(postHandler *handler.PostHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/posts", postHandler.List)
	mux.HandleFunc("/posts/", postHandler.GetByID)

	return mux
}
