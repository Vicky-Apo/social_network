package shared

import (
	"database/sql"
	"strings"
)

// NullableString converts a sql.NullString to *string.
func NullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

// NullableStringValue trims and normalizes a *string for persistence.
func NullableStringValue(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}
