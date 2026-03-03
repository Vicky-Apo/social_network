package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"social-network/backend/internal/app/server"
)

func main() {
	// closing context for safe shutdown on SIGINT or SIGTERM (Ctrl+C and kill command)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Run the server and block until the context is canceled
	if err := server.Run(ctx); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
