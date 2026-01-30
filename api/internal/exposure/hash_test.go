package exposure

import "testing"

func TestHasherStable(t *testing.T) {
	hasher := NewHasher("salt")
	first := hasher.Hash("user")
	second := hasher.Hash("user")
	if first == "" {
		t.Fatalf("expected hash value")
	}
	if first != second {
		t.Fatalf("expected stable hash, got %q and %q", first, second)
	}
}

func TestHasherEmpty(t *testing.T) {
	hasher := NewHasher("salt")
	if hasher.Hash("") != "" {
		t.Fatalf("expected empty hash for empty value")
	}
}
