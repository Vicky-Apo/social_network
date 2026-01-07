package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// LoadDotEnv reads key=value pairs from a local .env file without extra deps.
// Existing environment variables are not overwritten.
func LoadDotEnv(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = trimSpaces(line)
		if line == "" || line[0] == '#' {
			continue
		}
		key, val, ok := splitKeyValue(line)
		if !ok || key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// GetString returns the env value or the provided default if unset.
func GetString(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// GetInt returns the env value parsed as int or the default when missing/invalid.
func GetInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return def
}

// GetBool parses env value as boolean with common truthy/falsey strings.
// Recognizes: true/false, 1/0, yes/no (case-insensitive).
func GetBool(key string, def bool) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if val == "" {
		return def
	}
	switch val {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func trimSpaces(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func splitKeyValue(line string) (string, string, bool) {
	for i := 0; i < len(line); i++ {
		if line[i] == '=' {
			key := trimSpaces(line[:i])
			val := trimSpaces(line[i+1:])
			// support optional surrounding quotes
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			return key, val, true
		}
	}
	return "", "", false
}
