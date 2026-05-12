package staging

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

func TestStorePutLoadCurrent(t *testing.T) {
	store := New(t.TempDir())
	key := testKey()
	win := testWindow()
	wantBlob := []byte(`{"ok":true}`)
	ref := artifacts.Ref{Key: key, Window: win, Version: "v1"}

	path, err := store.Put(context.Background(), ref, wantBlob)
	if err != nil {
		t.Fatalf("put artifact: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join("popularity", "2026-01-01_2026-01-02", "v1.json")) {
		t.Fatalf("unexpected staged path: %s", path)
	}

	gotRef, gotBlob, ok, err := store.LoadCurrent(context.Background(), key, win)
	if err != nil {
		t.Fatalf("load current: %v", err)
	}
	if !ok {
		t.Fatalf("expected current artifact")
	}
	if gotRef.Version != "v1" {
		t.Fatalf("version = %q, want v1", gotRef.Version)
	}
	if string(gotBlob) != string(wantBlob) {
		t.Fatalf("blob = %s, want %s", gotBlob, wantBlob)
	}
}

func TestStoreLoadCurrentFallsBackToSingleJSON(t *testing.T) {
	store := New(t.TempDir())
	key := testKey()
	win := testWindow()
	winDir := store.windowDir(key, win)
	if err := os.MkdirAll(winDir, 0o755); err != nil {
		t.Fatalf("mkdir staging dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(winDir, "fallback.json"), []byte(`{"fallback":true}`), 0o600); err != nil {
		t.Fatalf("write fallback artifact: %v", err)
	}

	gotRef, gotBlob, ok, err := store.LoadCurrent(context.Background(), key, win)
	if err != nil {
		t.Fatalf("load current: %v", err)
	}
	if !ok {
		t.Fatalf("expected fallback artifact")
	}
	if gotRef.Version != "fallback" {
		t.Fatalf("version = %q, want fallback", gotRef.Version)
	}
	if string(gotBlob) != `{"fallback":true}` {
		t.Fatalf("blob = %s", gotBlob)
	}
}

func TestStoreRejectsTraversalVersion(t *testing.T) {
	store := New(t.TempDir())
	key := testKey()
	win := testWindow()
	winDir := store.windowDir(key, win)
	if err := os.MkdirAll(winDir, 0o755); err != nil {
		t.Fatalf("mkdir staging dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(winDir, "current.version"), []byte("../outside\n"), 0o600); err != nil {
		t.Fatalf("write current version: %v", err)
	}

	_, _, ok, err := store.LoadCurrent(context.Background(), key, win)
	if err == nil {
		t.Fatalf("expected traversal version to fail")
	}
	if ok {
		t.Fatalf("expected traversal version not to load an artifact")
	}
	if !strings.Contains(err.Error(), "invalid path segment") {
		t.Fatalf("error = %q, want invalid path segment", err.Error())
	}
	if strings.Contains(err.Error(), "outside") {
		t.Fatalf("error leaked rejected version: %q", err.Error())
	}

	_, err = store.Put(context.Background(), artifacts.Ref{Key: key, Window: win, Version: "../outside"}, []byte(`{}`))
	if err == nil {
		t.Fatalf("expected put with traversal version to fail")
	}
}

func testKey() artifacts.Key {
	return artifacts.Key{
		Tenant:  "tenant-a",
		Surface: "home",
		Type:    artifacts.TypePopularity,
	}
}

func testWindow() windows.Window {
	return windows.DayWindowUTC(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
}
