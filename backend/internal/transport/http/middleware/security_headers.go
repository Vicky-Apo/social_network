package middleware

import (
	"net/http"

	"social-network/backend/internal/config"
)

// SecurityHeaders returns middleware that adds security headers to responses
func SecurityHeaders(cfg config.SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip security headers if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Content Security Policy - helps prevent XSS attacks
			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			// X-Frame-Options - prevents clickjacking
			if cfg.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}

			// X-Content-Type-Options - prevents MIME type sniffing
			if cfg.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", cfg.XContentTypeOptions)
			}

			// Referrer-Policy - controls referrer information
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			// Permissions-Policy - controls browser features
			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			// Strict-Transport-Security - enforces HTTPS
			if cfg.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
			}

			next.ServeHTTP(w, r)
		})
	}
}
