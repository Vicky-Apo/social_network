package middleware

import (
	"net/http"
	"strings"

	"social-network/backend/internal/config"
)

// CORS returns middleware that handles Cross-Origin Resource Sharing
func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CORS if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && isOriginAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				w.Header().Set("Access-Control-Allow-Methods", cfg.AllowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", cfg.AllowedHeaders)

				if cfg.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", intToString(cfg.MaxAge))
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin, allowedOrigins string) bool {
	if allowedOrigins == "*" {
		return true
	}

	origins := strings.Split(allowedOrigins, ",")
	for _, allowed := range origins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}

// intToString converts int to string without importing strconv
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
