package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"social-network/backend/internal/transport/http/handler"
	transporthttp "social-network/backend/internal/transport/http"
	commentusecase "social-network/backend/internal/usecase/comment"
	postusecase "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/db/postgres"
	reactionusecase "social-network/backend/internal/usecase/reaction"
	commentrepo "social-network/backend/pkg/db/postgres/repositories/comment"
	postrepo "social-network/backend/pkg/db/postgres/repositories/post"
	"social-network/backend/pkg/utils"
	reactionrepo "social-network/backend/pkg/db/postgres/repositories/reaction"
)

const envFileName = ".env"

// Run boots the application and blocks until the context is canceled.
func Run(ctx context.Context) error {
	if err := utils.LoadDotEnv(envFileName); err != nil {
		log.Printf("warning: could not load %s: %v", envFileName, err)
	}

	dbURL := utils.GetString("DATABASE_URL", "")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is required (set in .env or environment)")
	}

	db, err := postgres.Open(postgres.WithDefaults(dbURL))
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer db.Close()

	if maxOpen := utils.GetInt("MAX_OPEN_CONNS", 0); maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdle := utils.GetInt("MAX_IDLE_CONNS", 0); maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}

	migrationsDir := utils.GetString("MIGRATIONS_PATH", "")
	if migrationsDir == "" {
		return fmt.Errorf("MIGRATIONS_PATH is required (set in .env or environment)")
	}
	migrationsPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	if err := postgres.ApplyMigrations(db, sourceURL); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	// Initialize repositories
	postRepository := postrepo.NewRepository(db)
	commentRepository := commentrepo.NewRepository(db)
	reactionRepository := reactionrepo.NewRepository(db)

	// Initialize services
	postService := postusecase.NewService(postRepository)
	commentService := commentusecase.NewService(commentRepository)
	reactionService := reactionusecase.NewService(reactionRepository)

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	reactionHandler := handler.NewReactionHandler(reactionService)

	// Create router with all handlers
	router := transporthttp.NewRouter(postHandler, commentHandler, reactionHandler)

	addr, err := requiredString("SERVER_ADDR")
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server boot completed, starting HTTP listener")

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown requested")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	return nil
}

func requiredString(key string) (string, error) {
	val := utils.GetString(key, "")
	if val == "" {
		return "", fmt.Errorf("%s is required (set in .env or environment)", key)
	}
	return val, nil
}
