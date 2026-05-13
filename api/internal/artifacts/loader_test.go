package artifacts

import (
	"strings"
	"testing"
)

func TestManifestURIValidatesTemplateSegments(t *testing.T) {
	loader := NewLoader(nil, LoaderConfig{ManifestTemplate: "/artifacts/{tenant}/{surface}/manifest.json"})

	_, err := loader.ManifestURI("tenant-a", "../surface")
	if err == nil {
		t.Fatal("ManifestURI() error = nil")
	}
	if strings.Contains(err.Error(), "/artifacts") {
		t.Fatalf("ManifestURI() leaked template path: %q", err.Error())
	}
}

func TestManifestURISubstitutesCleanSegments(t *testing.T) {
	loader := NewLoader(nil, LoaderConfig{ManifestTemplate: "/artifacts/{tenant}/{surface}/manifest.json"})

	got, err := loader.ManifestURI("tenant-a", "home")
	if err != nil {
		t.Fatalf("ManifestURI() error = %v", err)
	}
	want := "/artifacts/tenant-a/home/manifest.json"
	if got != want {
		t.Fatalf("ManifestURI() = %q, want %q", got, want)
	}
}
