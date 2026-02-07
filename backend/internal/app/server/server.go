package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"social-network/backend/internal/config"
	transporthttp "social-network/backend/internal/transport/http"
	"social-network/backend/internal/transport/http/handler"
	"social-network/backend/internal/transport/http/middleware"
	transportws "social-network/backend/internal/transport/websocket"
	authusecase "social-network/backend/internal/usecase/auth"
	chatusecase "social-network/backend/internal/usecase/chat"
	commentusecase "social-network/backend/internal/usecase/comment"
	followusecase "social-network/backend/internal/usecase/follow"
	notificationusecase "social-network/backend/internal/usecase/notification"
	postusecase "social-network/backend/internal/usecase/post"
	profileusecase "social-network/backend/internal/usecase/profile"
	reactionusecase "social-network/backend/internal/usecase/reaction"
	userusecase "social-network/backend/internal/usecase/user"
	"social-network/backend/pkg/db/postgres"
	authrepo "social-network/backend/pkg/db/postgres/repositories/auth"
	chatrepo "social-network/backend/pkg/db/postgres/repositories/chat"
	commentrepo "social-network/backend/pkg/db/postgres/repositories/comment"
	followrepo "social-network/backend/pkg/db/postgres/repositories/follow"
	grouprepo "social-network/backend/pkg/db/postgres/repositories/group"
	notificationrepo "social-network/backend/pkg/db/postgres/repositories/notification"
	postrepo "social-network/backend/pkg/db/postgres/repositories/post"
	reactionrepo "social-network/backend/pkg/db/postgres/repositories/reaction"
	userrepo "social-network/backend/pkg/db/postgres/repositories/user"
	"social-network/backend/pkg/logger"
	"social-network/backend/pkg/utils"
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
	userRepository := userrepo.NewRepository(db)
	followRepository := followrepo.NewRepository(db)
	chatRepository := chatrepo.NewRepository(db)
	groupRepository := grouprepo.NewRepository(db)
	notificationRepository := notificationrepo.NewRepository(db)

	// WebSocket hub (needed for notification publisher)
	wsHub := transportws.NewHub(followRepository, log)
	go wsHub.Run()
	defer wsHub.Stop()

	// Services
	authService := authusecase.NewService(authRepository, cfg.Auth, log)
	postService := postusecase.NewService(postRepository, userRepository, followRepository, log)
	notificationPublisher := transportws.NewNotificationPublisher(wsHub)
	notificationService := notificationusecase.NewService(notificationRepository, notificationPublisher)
	commentService := commentusecase.NewService(commentRepository, postRepository, notificationService)
	reactionService := reactionusecase.NewService(reactionRepository, postRepository, commentRepository, notificationService)
	profileService := profileusecase.NewService(userRepository, followRepository)
	followService := followusecase.NewService(userRepository, followRepository, notificationService)
	userService := userusecase.NewService(userRepository)
	chatService := chatusecase.NewService(chatRepository, groupRepository, followRepository, log)

	// Handlers
	authHandlerCfg := handler.AuthHandlerConfig{
		CookieName: cfg.Auth.SessionCookieName,
		MaxAge:     cfg.Auth.SessionMaxAge,
	}
	authHandler := handler.NewAuthHandler(authService, log, authHandlerCfg)
	postHandler := handler.NewPostHandler(postService, log)
	commentHandler := handler.NewCommentHandler(commentService, log)
	reactionHandler := handler.NewReactionHandler(reactionService, log)
	profileHandler := handler.NewProfileHandler(profileService, log)
	followHandler := handler.NewFollowHandler(followService, log)
	userHandler := handler.NewUserHandler(userService, log)
	notificationHandler := handler.NewNotificationHandler(notificationService, log)

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

	// WebSocket handler
	wsHandler := transportws.NewHandler(
		wsHub,
		chatService,
		rateLimiter,
		authService,
		cfg.Auth.SessionCookieName,
		cfg.CORS.Enabled,
		cfg.CORS.AllowedOrigins,
		log,
	)
	log.Info("websocket hub started")

	// Create router with all handlers
	router := transporthttp.NewRouter(
		postHandler,
		authHandler,
		commentHandler,
		reactionHandler,
		profileHandler,
		followHandler,
		userHandler,
		notificationHandler,
		wsHandler,
		mw,
	)

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
