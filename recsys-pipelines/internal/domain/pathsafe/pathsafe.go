package pathsafe

import (
	"fmt"
	"path"
	"strings"
	"unicode"
)

// Segment normalizes and validates a logical ID before it is used as one
// filesystem path segment.
func Segment(name, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s must be set", name)
	}
	if trimmed == "." || trimmed == ".." || strings.ContainsAny(trimmed, `/\`) {
		return "", fmt.Errorf("%s contains invalid path segment", name)
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("%s contains invalid path segment", name)
		}
	}
	return trimmed, nil
}

// RelativePath normalizes an object key before it is joined to a filesystem root.
func RelativePath(name, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimLeft(trimmed, "/")
	if trimmed == "" {
		return "", fmt.Errorf("%s must be set", name)
	}
	if strings.Contains(trimmed, `\`) {
		return "", fmt.Errorf("%s contains invalid path segment", name)
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("%s contains invalid path segment", name)
		}
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", fmt.Errorf("%s contains invalid path segment", name)
	}
	return cleaned, nil
}
