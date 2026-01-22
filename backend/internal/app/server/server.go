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
	followusecase "social-network/backend/internal/usecase/follow"
	postusecase "social-network/backend/internal/usecase/post"
	profileusecase "social-network/backend/internal/usecase/profile"
	"social-network/backend/pkg/db/postgres"
	followrepo "social-network/backend/pkg/db/postgres/repositories/follow"
	postrepo "social-network/backend/pkg/db/postgres/repositories/post"
	userrepo "social-network/backend/pkg/db/postgres/repositories/user"
	"social-network/backend/pkg/utils"
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

	postRepository := postrepo.NewRepository(db)
	postService := postusecase.NewService(postRepository)
	postHandler := handler.NewPostHandler(postService)

	userRepository := userrepo.NewRepository(db)
	followRepository := followrepo.NewRepository(db)

	profileService := profileusecase.NewService(userRepository, followRepository)
	followService := followusecase.NewService(userRepository, followRepository)

	profileHandler := handler.NewProfileHandler(profileService)
	followHandler := handler.NewFollowHandler(followService)

	router := transporthttp.NewRouter(postHandler, profileHandler, followHandler)

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
