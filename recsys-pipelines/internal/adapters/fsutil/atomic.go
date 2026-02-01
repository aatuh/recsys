package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// WriteFileAtomic writes a file by writing to a temporary file in the same
// directory and then renaming it into place.
//
// This provides atomic replace semantics on POSIX filesystems when the
// temporary file is created on the same mount as the target.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return err
	}
	// Best-effort durability.
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return err
	}
	if err := syncDir(dir); err != nil {
		return err
	}
	return nil
}

// CreateAtomicWriter creates a temporary file in the target directory and
// returns a commit function that renames it into place.
//
// The caller must close the returned file before calling commit.
func CreateAtomicWriter(path string, perm fs.FileMode) (*os.File, func() error, func() error, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, nil, err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return nil, nil, nil, err
	}
	tmpName := tmp.Name()

	rollback := func() error {
		_ = tmp.Close()
		return os.Remove(tmpName)
	}
	commit := func() error {
		if err := tmp.Chmod(perm); err != nil {
			_ = rollback()
			return err
		}
		if err := tmp.Sync(); err != nil {
			_ = rollback()
			return err
		}
		if err := tmp.Close(); err != nil {
			_ = rollback()
			return err
		}
		if err := os.Rename(tmpName, path); err != nil {
			_ = rollback()
			return err
		}
		if err := syncDir(dir); err != nil {
			return err
		}
		return nil
	}
	return tmp, commit, rollback, nil
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open dir for sync: %w", err)
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		// Some filesystems may not support syncing directories.
		if errors.Is(err, syscall.EINVAL) {
			return nil
		}
		return err
	}
	return nil
}
