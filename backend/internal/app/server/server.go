package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"social-network/backend/internal/config"
	"social-network/backend/internal/transport/http/handler"
	"social-network/backend/internal/transport/http/middleware"
	transporthttp "social-network/backend/internal/transport/http"
	commentusecase "social-network/backend/internal/usecase/comment"
	authusecase "social-network/backend/internal/usecase/auth"
	postusecase "social-network/backend/internal/usecase/post"
	"social-network/backend/pkg/db/postgres"
	reactionusecase "social-network/backend/internal/usecase/reaction"
	commentrepo "social-network/backend/pkg/db/postgres/repositories/comment"
	authrepo "social-network/backend/pkg/db/postgres/repositories/auth"
	postrepo "social-network/backend/pkg/db/postgres/repositories/post"
	"social-network/backend/pkg/logger"
	"social-network/backend/pkg/utils"
	reactionrepo "social-network/backend/pkg/db/postgres/repositories/reaction"
)

const envFileName = ".env"

// Run boots the application and blocks until the context is canceled.
func Run(ctx context.Context) error {
	// Load environment variables from .env file
	_ = utils.LoadDotEnv(envFileName)

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logger
	log := logger.NewDefault(cfg.Server.Debug)
	log.Info("starting application", logger.F("debug", cfg.Server.Debug))

	// Open database connection
	db, err := postgres.Open(postgres.WithDefaults(cfg.Database.URL))
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	log.Debug("database connection pool configured",
		logger.F("max_open", cfg.Database.MaxOpenConns),
		logger.F("max_idle", cfg.Database.MaxIdleConns))

	// Apply database migrations
	migrationsPath, err := filepath.Abs(cfg.Database.MigrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	if err := postgres.ApplyMigrations(db, sourceURL); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	log.Info("database migrations applied")

	// Repositories
	authRepository := authrepo.NewRepository(db)
	postRepository := postrepo.NewRepository(db)
	commentRepository := commentrepo.NewRepository(db)
	reactionRepository := reactionrepo.NewRepository(db)

	// Services
	authService := authusecase.NewService(authRepository, cfg.Auth, log)
	postService := postusecase.NewService(postRepository, log)
	commentService := commentusecase.NewService(commentRepository)
	reactionService := reactionusecase.NewService(reactionRepository)

	// Handlers
	authHandlerCfg := handler.AuthHandlerConfig{
		CookieName: cfg.Auth.SessionCookieName,
		MaxAge:     cfg.Auth.SessionMaxAge,
	}
	authHandler := handler.NewAuthHandler(authService, log, authHandlerCfg)
	postHandler := handler.NewPostHandler(postService, log)
	commentHandler := handler.NewCommentHandler(commentService)
	reactionHandler := handler.NewReactionHandler(reactionService)

	// Middleware (authService implements middleware.SessionValidator)
	authMiddleware := middleware.Auth(authService, cfg.Auth.SessionCookieName, log)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Enabled, log)
	defer rateLimiter.Stop()
	rateLimitMiddleware := middleware.RateLimit(rateLimiter)
	log.Info("rate limiter initialized",
		logger.F("enabled", cfg.RateLimit.Enabled),
		logger.F("requests_per_minute", cfg.RateLimit.RequestsPerMinute))

	// CORS middleware
	corsMiddleware := middleware.CORS(cfg.CORS)
	log.Info("CORS middleware initialized",
		logger.F("enabled", cfg.CORS.Enabled),
		logger.F("allowed_origins", cfg.CORS.AllowedOrigins))

	// Security headers middleware
	securityHeadersMiddleware := middleware.SecurityHeaders(cfg.SecurityHeaders)
	log.Info("security headers middleware initialized",
		logger.F("enabled", cfg.SecurityHeaders.Enabled))

	// Build middlewares struct
	mw := transporthttp.Middlewares{
		Auth:            authMiddleware,
		RateLimit:       rateLimitMiddleware,
		CORS:            corsMiddleware,
		SecurityHeaders: securityHeadersMiddleware,
	}

	// Create router with all handlers
	router := transporthttp.NewRouter(postHandler, authHandler, commentHandler, reactionHandler, mw)

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info("server boot completed, starting HTTP listener", logger.F("addr", cfg.Server.Addr))

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown requested")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", err)
	}
	log.Info("server stopped")
	return nil
}
