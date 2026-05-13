package fs

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStoreSetGetLastIngested(t *testing.T) {
	store := New(t.TempDir())
	day := time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC)

	if err := store.SetLastIngested(context.Background(), "tenant-a", "home", day); err != nil {
		t.Fatalf("SetLastIngested() error = %v", err)
	}
	got, ok, err := store.GetLastIngested(context.Background(), "tenant-a", "home")
	if err != nil {
		t.Fatalf("GetLastIngested() error = %v", err)
	}
	if !ok {
		t.Fatal("GetLastIngested() ok = false")
	}
	if got.Format("2006-01-02") != "2026-01-02" {
		t.Fatalf("day = %s", got)
	}
}

func TestStoreRejectsUnsafeCheckpointSegments(t *testing.T) {
	store := New(t.TempDir())

	err := store.SetLastIngested(context.Background(), "../tenant", "home", time.Now())
	if err == nil {
		t.Fatal("SetLastIngested() error = nil")
	}
	if !strings.Contains(err.Error(), "invalid path segment") {
		t.Fatalf("error = %q, want invalid path segment", err.Error())
	}
}
