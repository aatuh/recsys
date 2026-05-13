package objectstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRootedFSReaderReadsRelativePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o700); err != nil {
		t.Fatalf("create artifacts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "artifacts", "manifest.json"), []byte(`{"ok":true}`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	reader, err := NewRootedFSReader(root, 0)
	if err != nil {
		t.Fatalf("NewRootedFSReader() error = %v", err)
	}

	data, err := reader.Get(context.Background(), "file://artifacts/manifest.json")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("Get() = %q", string(data))
	}
}

func TestRootedFSReaderRejectsTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	reader, err := NewRootedFSReader(root, 0)
	if err != nil {
		t.Fatalf("NewRootedFSReader() error = %v", err)
	}

	_, err = reader.Get(context.Background(), "file://../secret.json")
	if err == nil {
		t.Fatalf("expected traversal error")
	}
	if strings.Contains(err.Error(), outside) {
		t.Fatalf("error leaked outside path: %v", err)
	}
}

func TestRootedFSReaderRejectsAbsolutePathOutsideRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	reader, err := NewRootedFSReader(root, 0)
	if err != nil {
		t.Fatalf("NewRootedFSReader() error = %v", err)
	}

	_, err = reader.Get(context.Background(), "file://"+outside)
	if err == nil {
		t.Fatalf("expected outside-root error")
	}
	if strings.Contains(err.Error(), outside) {
		t.Fatalf("error leaked outside path: %v", err)
	}
}

func TestRootedFSReaderRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink privileges vary on Windows")
	}
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.json")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	reader, err := NewRootedFSReader(root, 0)
	if err != nil {
		t.Fatalf("NewRootedFSReader() error = %v", err)
	}

	_, err = reader.Get(context.Background(), "file://link.json")
	if err == nil {
		t.Fatalf("expected symlink escape error")
	}
	var notFound ErrNotFound
	if errors.As(err, &notFound) {
		t.Fatalf("expected confinement error, got not found")
	}
	if strings.Contains(err.Error(), outside) {
		t.Fatalf("error leaked outside path: %v", err)
	}
}
