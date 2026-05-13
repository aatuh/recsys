package pathsafe

import (
	"fmt"
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
