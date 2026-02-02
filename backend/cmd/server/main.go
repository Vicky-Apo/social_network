package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"social-network/backend/internal/app/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
