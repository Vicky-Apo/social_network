package config

import (
	"fmt"

	"social-network/backend/pkg/utils"
)

// Config holds all application configuration
type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	Auth            AuthConfig
	RateLimit       RateLimitConfig
	CORS            CORSConfig
	SecurityHeaders SecurityHeadersConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	Enabled           bool
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
	MaxAge           int
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	Enabled                 bool
	ContentSecurityPolicy   string
	XFrameOptions           string
	XContentTypeOptions     string
	ReferrerPolicy          string
	PermissionsPolicy       string
	StrictTransportSecurity string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Addr           string
	Debug          bool
	UploadDir      string
	MaxUploadBytes int64
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL            string
	MigrationsPath string
	MaxOpenConns   int
	MaxIdleConns   int
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	BcryptCost          int
	SessionTokenBytes   int
	SessionDurationDays int
	SessionCookieName   string
	SessionMaxAge       int // Derived from SessionDurationDays (in seconds)
	MinPasswordLength   int
	MinUserAge          int
}

// Load reads configuration from environment variables
// Returns an error if any required values are missing
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Addr:           utils.GetString("SERVER_ADDR", ""),
			Debug:          utils.GetBool("DEBUG", false),
			UploadDir:      utils.GetString("UPLOAD_DIR", "backend/uploads"),
			MaxUploadBytes: int64(utils.GetInt("MAX_UPLOAD_MB", 10)) * 1024 * 1024,
		},
		Database: DatabaseConfig{
			URL:            utils.GetString("DATABASE_URL", ""),
			MigrationsPath: utils.GetString("MIGRATIONS_PATH", ""),
			MaxOpenConns:   utils.GetInt("MAX_OPEN_CONNS", 10),
			MaxIdleConns:   utils.GetInt("MAX_IDLE_CONNS", 5),
		},
		Auth: AuthConfig{
			BcryptCost:          utils.GetInt("BCRYPT_COST", 12),
			SessionTokenBytes:   utils.GetInt("SESSION_TOKEN_BYTES", 32),
			SessionDurationDays: utils.GetInt("SESSION_DURATION_DAYS", 30),
			SessionCookieName:   utils.GetString("SESSION_COOKIE_NAME", "session_token"),
			MinPasswordLength:   utils.GetInt("MIN_PASSWORD_LENGTH", 8),
			MinUserAge:          utils.GetInt("MIN_USER_AGE", 13),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: utils.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 1000),
			Enabled:           utils.GetBool("RATE_LIMIT_ENABLED", true),
		},
		CORS: CORSConfig{
			Enabled:          utils.GetBool("CORS_ENABLED", true),
			AllowedOrigins:   utils.GetString("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			AllowedMethods:   utils.GetString("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders:   utils.GetString("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
			AllowCredentials: utils.GetBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           utils.GetInt("CORS_MAX_AGE", 86400),
		},
		SecurityHeaders: SecurityHeadersConfig{
			Enabled:                 utils.GetBool("SECURITY_HEADERS_ENABLED", true),
			ContentSecurityPolicy:   utils.GetString("SECURITY_CSP", "default-src 'self'"),
			XFrameOptions:           utils.GetString("SECURITY_X_FRAME_OPTIONS", "DENY"),
			XContentTypeOptions:     utils.GetString("SECURITY_X_CONTENT_TYPE_OPTIONS", "nosniff"),
			ReferrerPolicy:          utils.GetString("SECURITY_REFERRER_POLICY", "strict-origin-when-cross-origin"),
			PermissionsPolicy:       utils.GetString("SECURITY_PERMISSIONS_POLICY", "geolocation=(), microphone=(), camera=()"),
			StrictTransportSecurity: utils.GetString("SECURITY_HSTS", "max-age=31536000; includeSubDomains"),
		},
	}

	// Derive SessionMaxAge from SessionDurationDays (convert days to seconds)
	cfg.Auth.SessionMaxAge = cfg.Auth.SessionDurationDays * 24 * 60 * 60

	// Validate required fields
	if cfg.Server.Addr == "" {
		return nil, fmt.Errorf("SERVER_ADDR is required")
	}
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Database.MigrationsPath == "" {
		return nil, fmt.Errorf("MIGRATIONS_PATH is required")
	}

	return cfg, nil
}
