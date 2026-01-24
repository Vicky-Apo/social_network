package middleware

import (
	"net/http"
	"sync"
	"time"

	"social-network/backend/internal/transport/http/utils"
	"social-network/backend/pkg/logger"
)

// RateLimiter tracks request counts per client IP using a sliding window approach
type RateLimiter struct {
	mu                sync.RWMutex
	requests          map[string][]time.Time
	requestsPerMinute int
	enabled           bool
	cleanupInterval   time.Duration
	stopCleanup       chan struct{}
	log               logger.Logger
}

// NewRateLimiter creates a new rate limiter with the specified requests per minute limit
func NewRateLimiter(requestsPerMinute int, enabled bool, log logger.Logger) *RateLimiter {
	rl := &RateLimiter{
		requests:          make(map[string][]time.Time),
		requestsPerMinute: requestsPerMinute,
		enabled:           enabled,
		cleanupInterval:   time.Minute * 5,
		stopCleanup:       make(chan struct{}),
		log:               log.WithFields(logger.F("middleware", "rate_limit")),
	}

	// Start background cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup periodically removes expired entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			windowStart := now.Add(-time.Minute)

			for ip, timestamps := range rl.requests {
				// Filter out timestamps older than the window
				valid := make([]time.Time, 0)
				for _, t := range timestamps {
					if t.After(windowStart) {
						valid = append(valid, t)
					}
				}

				if len(valid) == 0 {
					delete(rl.requests, ip)
				} else {
					rl.requests[ip] = valid
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCleanup:
			return
		}
	}
}

// Stop stops the background cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// isAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Get existing timestamps and filter expired ones
	timestamps := rl.requests[ip]
	valid := make([]time.Time, 0, len(timestamps))
	for _, t := range timestamps {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	// Check if limit exceeded
	if len(valid) >= rl.requestsPerMinute {
		rl.requests[ip] = valid
		return false
	}

	// Add current request timestamp
	valid = append(valid, now)
	rl.requests[ip] = valid

	return true
}

// RateLimit returns middleware that limits requests per IP address
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if disabled
			if !limiter.enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract client IP
			ip := utils.ExtractIPAddress(r)

			// Check if request is allowed
			if !limiter.isAllowed(ip) {
				limiter.log.Warn("rate limit exceeded",
					logger.F("ip", ip),
					logger.F("path", r.URL.Path),
					logger.F("method", r.Method))
				w.Header().Set("Retry-After", "60")
				utils.RespondWithError(w, http.StatusTooManyRequests, "rate limit exceeded, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
