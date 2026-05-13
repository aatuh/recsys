package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFSObjectStorePutGetConfinesNestedKeys(t *testing.T) {
	store := New(t.TempDir())

	uri, err := store.Put(context.Background(), "/tenant/home/object.json", "application/json", []byte(`{"ok":true}`))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	got, err := store.Get(context.Background(), uri)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("Get() = %s", got)
	}
}

func TestFSObjectStoreRejectsTraversalKeys(t *testing.T) {
	store := New(t.TempDir())

	for _, key := range []string{"../escape.json", "tenant/../../escape.json", `tenant\escape.json`} {
		t.Run(key, func(t *testing.T) {
			_, err := store.Put(context.Background(), key, "application/json", []byte(`{}`))
			if err == nil {
				t.Fatal("Put() error = nil")
			}
			if !strings.Contains(err.Error(), "invalid path segment") {
				t.Fatalf("Put() error = %q, want invalid path segment", err.Error())
			}
		})
	}
}

func TestFSObjectStoreRejectsFileURIEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	store := New(root)

	_, err := store.Get(context.Background(), "file://"+outside)
	if err == nil {
		t.Fatal("Get() error = nil")
	}
	if strings.Contains(err.Error(), outside) {
		t.Fatalf("Get() leaked rejected path: %q", err.Error())
	}
}
