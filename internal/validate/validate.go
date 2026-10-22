package validate

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

const MaxDescriptionLength = 500

func JavaPackage(value string) error {
	if value == "" {
		return fmt.Errorf("package name is required")
	}
	for _, part := range strings.Split(value, ".") {
		if part == "" || !isIdentifier(part) {
			return fmt.Errorf("%q is not a valid Java package name", value)
		}
	}
	return nil
}

func GroupID(value string) error { return JavaPackage(value) }

func SafeName(value string) error {
	if value == "" || value == "." || value == ".." || strings.Contains(value, "..") {
		return fmt.Errorf("%q is not a safe project name", value)
	}
	for _, r := range value {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_') {
			return fmt.Errorf("%q is not a safe project name", value)
		}
	}
	return nil
}

func Description(value string) error {
	if len(value) > MaxDescriptionLength {
		return fmt.Errorf("description must be at most %d characters", MaxDescriptionLength)
	}
	return nil
}

func OutputPath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("output directory is required")
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("invalid output directory: %w", err)
	}
	return filepath.Clean(abs), nil
}

func isIdentifier(value string) bool {
	for i, r := range value {
		if (i == 0 && !(r == '_' || unicode.IsLetter(r))) || (i > 0 && !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r))) {
			return false
		}
	}
	return true
}
