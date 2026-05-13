package fsutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Confine joins rel under root and rejects paths that escape root after
// cleaning. It does not resolve symlinks; callers should use private roots.
func Confine(root, rel string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("root must be set")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes root")
	}
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate := filepath.Join(cleanRoot, filepath.Clean(rel))
	cleanCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if cleanCandidate != cleanRoot && !strings.HasPrefix(cleanCandidate, cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return cleanCandidate, nil
}

// ConfineAbsolute rejects an absolute path unless it stays under root.
func ConfineAbsolute(root, path string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("root must be set")
	}
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	if candidate != cleanRoot && !strings.HasPrefix(candidate, cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return candidate, nil
}
