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
	id, remainder, ok := ParsePathIDAndRemainder(path, prefix)
	if !ok || remainder != "" {
		return 0, false
	}
	return id, true
}

// ParsePathIDAndRemainder extracts an int64 ID and remainder from a URL path with the given prefix.
// For example, ParsePathIDAndRemainder("/profiles/123/followers", "/profiles/") returns (123, "followers", true).
func ParsePathIDAndRemainder(path, prefix string) (int64, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, "", false
	}
	raw := strings.TrimPrefix(path, prefix)
	if raw == "" {
		return 0, "", false
	}
	parts := strings.Split(raw, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}
	remainder := ""
	if len(parts) > 1 {
		remainder = strings.Join(parts[1:], "/")
	}
	return id, remainder, true
}

// ParsePagination extracts limit/offset with defaults and bounds.
func ParsePagination(r *http.Request) (int, int, error) {
	limit := DefaultLimit
	offset := 0

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			return 0, 0, errors.New(MsgInvalidLimit)
		}
		if parsed > MaxLimit {
			parsed = MaxLimit
		}
		limit = parsed
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New(MsgInvalidOffset)
		}
		offset = parsed
	}

	return limit, offset, nil
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
