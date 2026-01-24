package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// ReadJSON decodes JSON request body into the destination struct
func ReadJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// ParsePathID extracts an int64 ID from a URL path with the given prefix.
// For example, ParsePathID("/posts/123", "/posts/") returns (123, true).
func ParsePathID(path, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(path, prefix)
	if raw == "" {
		return 0, false
	}
	if strings.Contains(raw, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// ExtractIPAddress extracts client IP from request, handling proxies
func ExtractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
