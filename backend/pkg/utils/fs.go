package utils

import (
	"fmt"
	"os"
)

// EnsureDir creates a directory and all parents if needed.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// SafeCreateFile creates a new file and fails if it already exists.
func SafeCreateFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	return f, nil
}
