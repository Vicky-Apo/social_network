package handler

import (
	"strconv"
	"strings"
)

func parseIDAndRemainder(path, prefix string) (int64, string, bool) {
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

func parseOptionalID(value string) (int64, bool) {
	if value == "" {
		return 0, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
