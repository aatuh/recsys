package validation

import (
	"testing"

	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

func TestValidateConfigPayload(t *testing.T) {
	if err := ValidateConfigPayload([]byte{}); err == nil {
		t.Fatalf("expected error for empty payload")
	}
	if err := ValidateConfigPayload([]byte(`{"weights":{"pop":-1}}`)); err == nil {
		t.Fatalf("expected error for negative weights")
	}
	if err := ValidateConfigPayload([]byte(`{"weights":{"pop":1,"cooc":0,"emb":0}}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRulesPayload(t *testing.T) {
	if err := ValidateRulesPayload([]byte(``)); err == nil {
		t.Fatalf("expected error for empty rules")
	}
	if err := ValidateRulesPayload([]byte(`"nope"`)); err == nil {
		t.Fatalf("expected error for non-object rules")
	}
	if err := ValidateRulesPayload([]byte(`[]`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCacheInvalidate(t *testing.T) {
	req := types.CacheInvalidateRequest{Targets: []string{"config", "rules"}}
	if err := ValidateCacheInvalidate(&req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req = types.CacheInvalidateRequest{Targets: []string{"unknown"}}
	if err := ValidateCacheInvalidate(&req); err == nil {
		t.Fatalf("expected error for invalid target")
	}
}
