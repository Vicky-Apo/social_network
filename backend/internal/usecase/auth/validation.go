package auth

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// Validation errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrMissingName        = errors.New("first name and last name are required")
	ErrInvalidDateFormat  = errors.New("invalid date format, expected DD/MM/YYYY")
	ErrInvalidDateOfBirth = errors.New("invalid date of birth")
	ErrUserTooYoung       = errors.New("user must be at least 13 years old")
)

// Email validation regex (basic validation)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword checks if password meets minimum requirements
func ValidatePassword(password string, minLength int) error {
	if len(password) < minLength {
		return ErrWeakPassword
	}
	return nil
}

// ValidateName checks if first and last name are provided
func ValidateName(firstName, lastName string) error {
	if firstName == "" || lastName == "" {
		return ErrMissingName
	}
	return nil
}

// ParseDate converts "DD/MM/YYYY" string to time.Time
func ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}
	return t, nil
}

// ValidateDateOfBirth checks if date of birth is valid
func ValidateDateOfBirth(dob time.Time, minAge int) error {
	now := time.Now()

	// Check if date is in the future
	if dob.After(now) {
		return ErrInvalidDateOfBirth
	}

	// Check minimum age
	minDate := now.AddDate(-minAge, 0, 0)
	if dob.After(minDate) {
		return ErrUserTooYoung
	}

	// Check reasonable maximum age (150 years)
	maxDate := now.AddDate(-150, 0, 0)
	if dob.Before(maxDate) {
		return ErrInvalidDateOfBirth
	}

	return nil
}

// TrimStrings trims whitespace from common registration fields
func TrimStrings(email, firstName, lastName *string) {
	*email = strings.TrimSpace(*email)
	*firstName = strings.TrimSpace(*firstName)
	*lastName = strings.TrimSpace(*lastName)
}
